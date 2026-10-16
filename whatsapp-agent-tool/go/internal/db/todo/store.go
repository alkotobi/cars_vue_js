package todo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"whatsapp-agent-tool/internal/db"
)

// CreateInput is the payload for Create. OwnerProfile is the login slug; when
// ParentID is set the new todo joins that parent's owner (the tree is one person's).
type CreateInput struct {
	OwnerProfile string
	OwnerDisplay string
	OwnerNumber  string
	ParentID     *int64
	Title        string
	Description  string
	Priority     string
	DueAt        string
	RequestedBy  string
	ChatRef      string
}

// Create inserts a todo and, when it has a parent, re-runs the ancestor cascade
// (a new open child reopens a done parent). Returns the new todo id.
func Create(in CreateInput) (int64, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return 0, fmt.Errorf("a todo needs a title")
	}
	priority := strings.TrimSpace(in.Priority)
	if priority == "" {
		priority = "normal"
	}
	if !ValidPriority(priority) {
		return 0, fmt.Errorf("invalid priority %q", priority)
	}
	// A root todo owns itself to the login; a subtask inherits its parent's owner,
	// so only resolve (and create) the login row when this is a root.
	var ownerID int64
	if in.ParentID == nil {
		var err error
		ownerID, err = EnsureUser(in.OwnerProfile, in.OwnerDisplay, in.OwnerNumber)
		if err != nil {
			return 0, err
		}
	}

	var newID int64
	err := db.WithTransaction(func(tx *sql.Tx) error {
		ownerForRow := ownerID
		if in.ParentID != nil {
			parent, err := txGetNode(tx, *in.ParentID)
			if err != nil {
				return err
			}
			if parent == nil {
				return fmt.Errorf("parent todo %d not found", *in.ParentID)
			}
			ownerForRow = parent.ownerID
		}
		res, err := tx.ExecContext(context.Background(),
			"INSERT INTO todos (parent_id, owner_id, title, description, status, priority, due_at, requested_by, chat_ref, created_at, updated_at) "+
				"VALUES (?,?,?,?,?,?,?,?,?,?,?)",
			in.ParentID, ownerForRow, strings.TrimSpace(in.Title), nullIfEmpty(in.Description),
			StatusOpen, priority, sqlDatetime(in.DueAt), nullIfEmpty(in.RequestedBy), nullIfEmpty(in.ChatRef), nowSQL(), nowSQL())
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		newID = id
		if in.ParentID != nil {
			return recomputeFrom(tx, *in.ParentID, systemActor)
		}
		return nil
	})
	return newID, err
}

// SetStatus changes one todo by hand (open/done/cancelled), then re-derives its
// ancestors. A cancelled node is terminal for the cascade.
func SetStatus(id int64, status, by string) error {
	if !ValidStatus(status) {
		return fmt.Errorf("invalid status %q; one of open, done, cancelled", status)
	}
	return db.WithTransaction(func(tx *sql.Tx) error {
		n, err := txGetNode(tx, id)
		if err != nil {
			return err
		}
		if n == nil {
			return fmt.Errorf("todo %d not found", id)
		}
		if n.status == status {
			return nil
		}
		if err := txSetStatus(tx, id, n.status, status, by, false, "manual"); err != nil {
			return err
		}
		if n.parentID != nil {
			return recomputeFrom(tx, *n.parentID, by)
		}
		return nil
	})
}

// Move re-parents a todo (nil = make it a root), rejecting a move under its own
// descendant, then re-derives both the old and the new ancestor chain.
func Move(id int64, newParent *int64, by string) error {
	return db.WithTransaction(func(tx *sql.Tx) error {
		n, err := txGetNode(tx, id)
		if err != nil {
			return err
		}
		if n == nil {
			return fmt.Errorf("todo %d not found", id)
		}
		oldParent := n.parentID
		if newParent != nil {
			if *newParent == id {
				return fmt.Errorf("todo %d cannot be its own parent", id)
			}
			descendant, err := isDescendant(tx, id, *newParent)
			if err != nil {
				return err
			}
			if descendant {
				return fmt.Errorf("cannot move todo %d under its own descendant %d", id, *newParent)
			}
			parent, err := txGetNode(tx, *newParent)
			if err != nil {
				return err
			}
			if parent == nil {
				return fmt.Errorf("parent todo %d not found", *newParent)
			}
			if _, err := tx.ExecContext(context.Background(), "UPDATE todos SET parent_id = ?, updated_at = ? WHERE id = ?", *newParent, nowSQL(), id); err != nil {
				return err
			}
		} else {
			if _, err := tx.ExecContext(context.Background(), "UPDATE todos SET parent_id = NULL, updated_at = ? WHERE id = ?", nowSQL(), id); err != nil {
				return err
			}
		}
		if oldParent != nil {
			if err := recomputeFrom(tx, *oldParent, by); err != nil {
				return err
			}
		}
		if newParent != nil {
			return recomputeFrom(tx, *newParent, by)
		}
		return nil
	})
}

// isDescendant reports whether candidate sits somewhere under ancestor's subtree.
func isDescendant(tx *sql.Tx, ancestor, candidate int64) (bool, error) {
	cur := candidate
	for cur != 0 {
		if cur == ancestor {
			return true, nil
		}
		n, err := txGetNode(tx, cur)
		if err != nil {
			return false, err
		}
		if n == nil || n.parentID == nil {
			return false, nil
		}
		cur = *n.parentID
	}
	return false, nil
}

// Assign re-owns a single todo (children keep their own owner). The target login
// is auto-registered when unknown.
func Assign(id int64, ownerProfile, ownerDisplay, ownerNumber string) (int64, error) {
	ownerID, err := EnsureUser(ownerProfile, ownerDisplay, ownerNumber)
	if err != nil {
		return 0, err
	}
	affected, err := db.Execute("UPDATE todos SET owner_id = ?, updated_at = ? WHERE id = ?", ownerID, nowSQL(), id)
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		return 0, fmt.Errorf("todo %d not found", id)
	}
	return ownerID, nil
}

// UpdatePatch carries the editable fields (nil pointer = leave unchanged).
type UpdatePatch struct {
	Title       *string
	Description *string
	Priority    *string
	DueAt       *string // set "" to clear
	RequestedBy *string
	ChatRef     *string
}

// Update edits a todo's descriptive fields (not status/parent — use SetStatus/Move).
func Update(id int64, p UpdatePatch) (bool, error) {
	sets := []string{}
	var params []any
	if p.Title != nil {
		if strings.TrimSpace(*p.Title) == "" {
			return false, fmt.Errorf("title cannot be empty")
		}
		sets = append(sets, "title = ?")
		params = append(params, strings.TrimSpace(*p.Title))
	}
	if p.Description != nil {
		sets = append(sets, "description = ?")
		params = append(params, nullIfEmpty(*p.Description))
	}
	if p.Priority != nil {
		if !ValidPriority(*p.Priority) {
			return false, fmt.Errorf("invalid priority %q; one of %s", *p.Priority, strings.Join(Priorities, ", "))
		}
		sets = append(sets, "priority = ?")
		params = append(params, *p.Priority)
	}
	if p.DueAt != nil {
		sets = append(sets, "due_at = ?")
		params = append(params, sqlDatetime(*p.DueAt))
	}
	if p.RequestedBy != nil {
		sets = append(sets, "requested_by = ?")
		params = append(params, nullIfEmpty(*p.RequestedBy))
	}
	if p.ChatRef != nil {
		sets = append(sets, "chat_ref = ?")
		params = append(params, nullIfEmpty(*p.ChatRef))
	}
	sets = append(sets, "updated_at = ?")
	params = append(params, nowSQL())
	if len(sets) == 1 { // only the always-present updated_at → nothing to change
		return false, nil
	}
	if len(sets) == 0 {
		return false, nil
	}
	params = append(params, id)
	affected, err := db.Execute(fmt.Sprintf("UPDATE todos SET %s WHERE id = ?", strings.Join(sets, ", ")), params...)
	if err != nil {
		return false, err
	}
	if affected == 0 {
		return false, fmt.Errorf("todo %d not found", id)
	}
	return true, nil
}

// Delete removes a todo; the FK cascade removes its whole subtree, then the
// former parent's derived status is recomputed.
func Delete(id int64) (int64, error) {
	var affected int64
	err := db.WithTransaction(func(tx *sql.Tx) error {
		n, err := txGetNode(tx, id)
		if err != nil {
			return err
		}
		if n == nil {
			return fmt.Errorf("todo %d not found", id)
		}
		res, err := tx.ExecContext(context.Background(), "DELETE FROM todos WHERE id = ?", id)
		if err != nil {
			return err
		}
		affected, _ = res.RowsAffected()
		if n.parentID != nil {
			return recomputeFrom(tx, *n.parentID, systemActor)
		}
		return nil
	})
	return affected, err
}
