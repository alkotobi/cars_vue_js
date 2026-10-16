// Package chatbotstate is the SHARED state of the cars-app chat AI, kept in the
// `ai` DB so several machines can run `whatsapp-tool chat-poll` against one tenant
// with a single active writer (tables chatbot_group_cursors / chatbot_leases,
// migration 009). It has two concerns:
//
//   - the per-group cursor: the highest client message id already answered, so a
//     machine that takes over resumes at the right message (writes are monotonic);
//   - the single-writer lease: one row per tenant, held by the active daemon and
//     renewed by a heartbeat. A standby may claim it only once expires_at (a DB
//     clock value) has passed, i.e. the holder stopped renewing.
//
// Every time comparison uses the DB clock (NOW(3)) — never the caller's — so a
// machine with a skewed clock cannot split the lock.
package chatbotstate

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/db"
)

// ---------------------------------------------------------------- cursors

// LoadCursors returns the tenant's per-group cursors keyed by group id. A group
// with no row is simply absent (treated as 0 by the caller).
func LoadCursors(tenant string) (map[int64]int64, error) {
	rows, err := db.Query("SELECT id_chat_group, last_message_id FROM chatbot_group_cursors WHERE tenant = ?", tenant)
	if err != nil {
		return nil, err
	}
	out := make(map[int64]int64, len(rows))
	for _, r := range rows {
		out[int64(num(r["id_chat_group"]))] = int64(num(r["last_message_id"]))
	}
	return out, nil
}

// SaveCursor records the last-answered id for one group. It is MONOTONIC — a
// GREATEST upsert — so a stale writer can never move a shared cursor backwards
// (and two machines importing local state at once keep the larger value).
func SaveCursor(tenant string, groupID, lastID int64) error {
	_, err := db.Execute(
		"INSERT INTO chatbot_group_cursors (tenant, id_chat_group, last_message_id) VALUES (?,?,?) "+
			"ON DUPLICATE KEY UPDATE last_message_id = GREATEST(last_message_id, VALUES(last_message_id))",
		tenant, groupID, lastID)
	return err
}

// ---------------------------------------------------------------- lease

// Owner builds this process's lease owner string: host:pid:profile. It is stable
// for the life of the process and makes "who is the active daemon" observable.
func Owner() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknown"
	}
	return fmt.Sprintf("%s:%d:%s", host, os.Getpid(), config.WAProfile())
}

// Acquire tries to take the lease for the tenant, returning true when this owner
// holds it afterwards. It holds the row with SELECT ... FOR UPDATE so two
// standbys starting at once serialise: the first claims it, the second sees a
// live lease owned by someone else and is refused. A lease owned by us is always
// (re)claimed; another owner's lease is claimed only once expires_at has passed.
func Acquire(tenant, owner string, ttl time.Duration) (bool, error) {
	secs := int64(ttl / time.Second)
	if secs <= 0 {
		secs = 1
	}
	return db.WithTransactionR(func(tx *sql.Tx) (bool, error) {
		rows, err := db.TxQuery(tx,
			"SELECT owner, (expires_at <= NOW(3)) AS expired FROM chatbot_leases WHERE tenant = ? FOR UPDATE",
			tenant)
		if err != nil {
			return false, err
		}
		if len(rows) == 0 {
			_, _, err := db.TxExec(tx,
				"INSERT INTO chatbot_leases (tenant, owner, acquired_at, heartbeat_at, expires_at) "+
					"VALUES (?,?,NOW(3),NOW(3),DATE_ADD(NOW(3), INTERVAL ? SECOND))",
				tenant, owner, secs)
			return true, err
		}
		if str(rows[0]["owner"]) != owner && !boolish(rows[0]["expired"]) {
			return false, nil // held by another, live owner — stand by
		}
		// Claim/renew. acquired_at keeps the original when we already held it, and
		// is reset when we take over an expired lease from someone else — so its
		// condition must read `owner` BEFORE it is overwritten below.
		_, _, err = db.TxExec(tx,
			"UPDATE chatbot_leases SET acquired_at = IF(owner = ?, acquired_at, NOW(3)), owner = ?, "+
				"heartbeat_at = NOW(3), expires_at = DATE_ADD(NOW(3), INTERVAL ? SECOND) WHERE tenant = ?",
			owner, owner, secs, tenant)
		return true, err
	})
}

// Renew extends the lease's expires_at by ttl. It returns false when this owner
// no longer holds a live lease (someone else took over, or it already expired) —
// the caller must then stop writing. It re-checks with a SELECT rather than
// trusting RowsAffected, which MySQL reports as 0 for a no-op update.
func Renew(tenant, owner string, ttl time.Duration) (bool, error) {
	secs := int64(ttl / time.Second)
	if secs <= 0 {
		secs = 1
	}
	if _, err := db.Execute(
		"UPDATE chatbot_leases SET heartbeat_at = NOW(3), expires_at = DATE_ADD(NOW(3), INTERVAL ? SECOND) "+
			"WHERE tenant = ? AND owner = ?",
		secs, tenant, owner); err != nil {
		return false, err
	}
	row, err := db.QueryOne(
		"SELECT 1 AS ok FROM chatbot_leases WHERE tenant = ? AND owner = ? AND expires_at > NOW(3)",
		tenant, owner)
	if err != nil {
		return false, err
	}
	return row != nil, nil
}

// Holds reports whether this owner still holds a live lease.
func Holds(tenant, owner string) (bool, error) {
	row, err := db.QueryOne(
		"SELECT 1 AS ok FROM chatbot_leases WHERE tenant = ? AND owner = ? AND expires_at > NOW(3)",
		tenant, owner)
	if err != nil {
		return false, err
	}
	return row != nil, nil
}

// Release gives the lease up immediately (expires_at = now) so a standby can take
// over without waiting for the TTL. Only the current owner can release.
func Release(tenant, owner string) error {
	_, err := db.Execute(
		"UPDATE chatbot_leases SET expires_at = NOW(3) WHERE tenant = ? AND owner = ?",
		tenant, owner)
	return err
}

// ---------------------------------------------------------------- value helpers

func num(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case float32:
		return float64(t)
	case int64:
		return float64(t)
	case uint64:
		return float64(t)
	case int:
		return float64(t)
	case []byte:
		f, _ := strconv.ParseFloat(string(t), 64)
		return f
	case string:
		f, _ := strconv.ParseFloat(t, 64)
		return f
	}
	return 0
}

func str(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	}
	return ""
}

func boolish(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case float64:
		return t != 0
	case float32:
		return t != 0
	case int64:
		return t != 0
	case uint64:
		return t != 0
	case int:
		return t != 0
	case []byte:
		s := string(t)
		return s != "0" && s != ""
	case string:
		return t != "0" && t != ""
	}
	return false
}
