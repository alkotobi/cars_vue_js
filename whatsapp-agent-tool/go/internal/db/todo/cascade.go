package todo

import (
	"context"
	"database/sql"

	"whatsapp-agent-tool/internal/db"
)

// systemActor is the changed_by value for cascade changes (not a human).
const systemActor = "system"

// cascadeNode is the minimal row the cascade needs while walking the tree.
type cascadeNode struct {
	id       int64
	parentID *int64
	status   string
	ownerID  int64
}

func txGetNode(tx *sql.Tx, id int64) (*cascadeNode, error) {
	rows, err := db.TxQuery(tx, "SELECT id, parent_id, status, owner_id FROM todos WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, nil
	}
	r := rows[0]
	n := &cascadeNode{id: int64(fnum(r["id"])), status: s(r["status"]), ownerID: int64(fnum(r["owner_id"]))}
	if pid := fnum(r["parent_id"]); pid > 0 {
		v := int64(pid)
		n.parentID = &v
	}
	return n, nil
}

// txChildStats returns how many children a node has and how many are resolved.
func txChildStats(tx *sql.Tx, id int64) (total, resolved int, err error) {
	rows, err := db.TxQuery(tx, "SELECT COUNT(*) AS total, SUM(status IN ('done','cancelled')) AS resolved FROM todos WHERE parent_id = ?", id)
	if err != nil {
		return 0, 0, err
	}
	if len(rows) == 0 {
		return 0, 0, nil
	}
	return int(fnum(rows[0]["total"])), int(fnum(rows[0]["resolved"])), nil
}

// txSetStatus writes a status change and its history row (auto = cascade).
func txSetStatus(tx *sql.Tx, id int64, from, to, by string, auto bool, reason string) error {
	ctx := context.Background()
	if to == StatusDone {
		if _, err := tx.ExecContext(ctx, "UPDATE todos SET status = ?, done_at = ?, updated_at = ? WHERE id = ?", to, nowSQL(), nowSQL(), id); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, "UPDATE todos SET status = ?, done_at = NULL, updated_at = ? WHERE id = ?", to, nowSQL(), id); err != nil {
			return err
		}
	}
	_, err := tx.ExecContext(ctx,
		"INSERT INTO todo_status_history (todo_id, from_status, to_status, changed_by, auto, reason, changed_at) VALUES (?,?,?,?,?,?,?)",
		id, nullIfEmpty(from), to, by, boolToTiny(auto), nullIfEmpty(reason), nowSQL())
	return err
}

// recomputeFrom walks from startID up the ancestor chain, flipping any node whose
// derived status (done iff all children resolved) no longer matches. A cancelled
// node is terminal, and a node that just became childless keeps its manual status.
func recomputeFrom(tx *sql.Tx, startID int64, by string) error {
	cur := startID
	for cur != 0 {
		n, err := txGetNode(tx, cur)
		if err != nil {
			return err
		}
		if n == nil || n.status == StatusCancelled {
			return nil
		}
		total, resolved, err := txChildStats(tx, n.id)
		if err != nil {
			return err
		}
		if total == 0 {
			return nil
		}
		desired := StatusOpen
		if resolved == total {
			desired = StatusDone
		}
		if desired == n.status {
			return nil
		}
		if err := txSetStatus(tx, n.id, n.status, desired, by, true, "auto cascade"); err != nil {
			return err
		}
		if n.parentID == nil {
			return nil
		}
		cur = *n.parentID
	}
	return nil
}

func boolToTiny(b bool) int {
	if b {
		return 1
	}
	return 0
}
