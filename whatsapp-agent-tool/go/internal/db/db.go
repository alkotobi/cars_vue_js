// Package db is the Go twin of src/db/index.ts — a MySQL access layer (single
// lazily-created pool, connection-glitch retries, the shared toIso normaliser).
// Pure I/O: no business logic.
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"whatsapp-agent-tool/internal/config"
)

var pool *sql.DB
var openErr error

// Open returns the shared pool, creating it once. Mirrors getPool().
func Open() (*sql.DB, error) {
	if pool != nil {
		return pool, nil
	}
	if openErr != nil {
		return nil, openErr
	}
	db, err := sql.Open("mysql", DSN())
	if err != nil {
		openErr = err
		return nil, err
	}
	db.SetMaxOpenConns(10) // connectionLimit: 10, waitForConnections: true, queueLimit: 0
	pool = db
	return pool, nil
}

// DSN builds the DSN from config with the same options the pool uses:
// charset utf8mb4, timezone Z (UTC), parseTimes so DATETIME scans are t=UTC.
func DSN() string {
	cfg := config.C.DB
	mc := mysql.NewConfig()
	mc.User = cfg.User
	mc.Passwd = cfg.Password
	mc.Net = "tcp"
	mc.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mc.DBName = cfg.Database
	// parseTime=false: DATE/DATETIME come back as []byte strings, exactly like
	// mysql2's dateStrings:true — we control the UTC formatting via ToIso().
	mc.ParseTime = false
	mc.Loc = time.UTC
	mc.Timeout = 8 * time.Second // CONNECT_TIMEOUT_MS
	mc.Params = map[string]string{"charset": "utf8mb4"}
	return mc.FormatDSN()
}

var retryDelays = []time.Duration{300 * time.Millisecond, 800 * time.Millisecond, 1500 * time.Millisecond}
var transientRe = regexp.MustCompile(`ETIMEDOUT|ECONNRESET|connection (?:lost|clos)`)

// isTransientError mirrors src/db/index.ts: connection-level failures worth
// retrying, never statement-level (syntax, constraint…).
func isTransientError(err error) bool {
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		return me.Number == 1040 // ER_CON_COUNT_ERROR
	}
	if err == nil {
		return false
	}
	msg := err.Error()
	if strings.Contains(msg, "connection refused") { // ECONNREFUSED
		return true
	}
	if strings.Contains(msg, "no route to host") { // EHOSTUNREACH
		return true
	}
	if strings.Contains(msg, "network is unreachable") { // ENETUNREACH
		return true
	}
	if strings.Contains(msg, "i/o timeout") || strings.Contains(msg, "timeout") { // ETIMEDOUT
		return true
	}
	if strings.Contains(msg, "bad connection") || strings.Contains(msg, "invalid connection") {
		return true
	}
	return transientRe.MatchString(msg)
}

// withRetry runs fn, retrying ONLY transient connection errors with a short
// backoff. Safe for idempotent reads; writes go through exactly-once paths.
func withRetry[T any](fn func() (T, error)) (T, error) {
	var zero T
	for attempt := 0; ; attempt++ {
		v, err := fn()
		if err == nil {
			return v, nil
		}
		if attempt >= len(retryDelays) || !isTransientError(err) {
			return zero, err
		}
		fmt.Fprintf(os.Stderr, "[db] transient error (%v) — retrying\n", err)
		time.Sleep(retryDelays[attempt])
	}
}

// Query runs an idempotent SELECT, retrying transient handshake glitches.
// Returns rows as []map[string]any with the same type mapping as mysql2
// (DECIMAL/INT -> float64, DATETIME -> time.Time, NULL -> nil, strings as-is).
func Query(query string, params ...any) ([]map[string]any, error) {
	return withRetry(func() ([]map[string]any, error) {
		db, err := Open()
		if err != nil {
			return nil, err
		}
		rows, err := db.QueryContext(context.Background(), query, params...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanRows(rows)
	})
}

// QueryOne returns the first row or nil.
func QueryOne(query string, params ...any) (map[string]any, error) {
	rows, err := Query(query, params...)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

// Execute runs a write EXACTLY ONCE on a dedicated connection (mirrors
// acquireConnection + execute). Returns affected rows; error on transient only
// retries the handshake, never the statement.
func Execute(query string, params ...any) (int64, error) {
	db, err := Open()
	if err != nil {
		return 0, err
	}
	conn, err := withRetry(func() (*sql.Conn, error) { return db.Conn(context.Background()) })
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	res, err := conn.ExecContext(context.Background(), query, params...)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Insert runs an INSERT exactly once and returns the last insert id.
func Insert(query string, params ...any) (int64, error) {
	_, id, err := execExact(query, params...)
	return id, err
}

// execExact runs a statement exactly once and reports affected rows + insert id.
func execExact(query string, params ...any) (int64, int64, error) {
	db, err := Open()
	if err != nil {
		return 0, 0, err
	}
	conn, err := withRetry(func() (*sql.Conn, error) { return db.Conn(context.Background()) })
	if err != nil {
		return 0, 0, err
	}
	defer conn.Close()
	res, err := conn.ExecContext(context.Background(), query, params...)
	if err != nil {
		return 0, 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, 0, err
	}
	return affected, id, nil
}

// WithTransaction runs fn inside a transaction on a dedicated connection.
func WithTransaction(fn func(tx *sql.Tx) error) error {
	_, err := WithTransactionR(func(tx *sql.Tx) (any, error) { return nil, fn(tx) })
	return err
}

// WithTransactionR runs fn inside a transaction and returns its value. Write
// statements inside the transaction run exactly once on the transaction's
// connection (no statement-level retry).
func WithTransactionR[T any](fn func(tx *sql.Tx) (T, error)) (T, error) {
	var zero T
	db, err := Open()
	if err != nil {
		return zero, err
	}
	conn, err := withRetry(func() (*sql.Conn, error) { return db.Conn(context.Background()) })
	if err != nil {
		return zero, err
	}
	defer conn.Close()
	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		return zero, err
	}
	done := false
	defer func() {
		if !done {
			_ = tx.Rollback()
		}
	}()
	v, err := fn(tx)
	if err != nil {
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		return zero, err
	}
	done = true
	return v, nil
}

// TxQuery runs a SELECT inside a transaction (pricing writes read the previous
// value before updating).
func TxQuery(tx *sql.Tx, query string, params ...any) ([]map[string]any, error) {
	rows, err := tx.QueryContext(context.Background(), query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRows(rows)
}

// TxExec runs a write inside a transaction; returns affected rows + insert id.
func TxExec(tx *sql.Tx, query string, params ...any) (int64, int64, error) {
	res, err := tx.ExecContext(context.Background(), query, params...)
	if err != nil {
		return 0, 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, 0, err
	}
	return affected, id, nil
}

// ClosePool mirrors closePool().
func ClosePool() {
	if pool != nil {
		pool.Close()
		pool = nil
	}
}

// scanRows converts driver values to the mysql2-style shapes the loaders expect.
func scanRows(rows *sql.Rows) ([]map[string]any, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	var out []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any, len(cols))
		for i, c := range cols {
			row[c] = convert(colTypes[i].DatabaseTypeName(), vals[i])
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func convert(dbType string, v any) any {
	if v == nil {
		return nil
	}
	switch t := v.(type) {
	case []byte:
		s := string(t)
		if isNumericType(dbType) {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return f
			}
		}
		return s
	case time.Time:
		return t
	default:
		return v
	}
}

func isNumericType(dbType string) bool {
	u := strings.ToUpper(dbType)
	if strings.Contains(u, "INT") {
		return true
	}
	for _, p := range []string{"DECIMAL", "NUMERIC", "FLOAT", "DOUBLE", "YEAR"} {
		if strings.HasPrefix(u, p) {
			return true
		}
	}
	return false
}

// ToIso mirrors src/db/index.ts toIso(): DATETIME(3) stored as UTC becomes an
// ISO-8601 UTC string ("2026-10-10T20:22:28.853Z") — indistinguishable from the
// old JSON `timestamp`. Accepts nil, time.Time, or driver strings/bytes.
func ToIso(v any) *string {
	if v == nil {
		return nil
	}
	var s string
	switch t := v.(type) {
	case time.Time:
		out := t.UTC().Format("2006-01-02T15:04:05.000Z07:00")
		return &out
	case []byte:
		s = string(t)
	case string:
		s = t
	default:
		return nil
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	normalized := s
	if !strings.Contains(s, "T") {
		normalized = strings.Replace(s, " ", "T", 1)
	}
	if !msSuffixRe(normalized) {
		normalized += ".000"
	}
	normalized = strings.TrimSuffix(normalized, "Z") + "Z"
	return &normalized
}

func msSuffixRe(s string) bool {
	i := strings.LastIndex(s, ".")
	if i < 0 {
		return false
	}
	for j := i + 1; j < len(s); j++ {
		c := s[j]
		if c < '0' || c > '9' {
			return j > i+1 // at least one digit
		}
	}
	return i+1 < len(s)
}
