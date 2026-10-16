// Per-group cursors for the chat AI. Each group keeps the highest client message
// id already answered, so the daemon never re-reads (nor re-answers) history.
//
// The cursors are SHARED (table chatbot_group_cursors in the `ai` DB, migration
// 009) so several machines can run the daemon against one tenant and a machine
// that takes over the lease resumes at the right message. A one-time import from
// the legacy per-machine data/chatbot-state.json seeds the shared store when it
// is still empty, so upgrading does not re-answer history.
package chatbot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/db/chatbotstate"
)

// cursorFile is the legacy on-disk store (relative to the repo's data dir), kept
// only as the one-time seed for the shared store.
const cursorFile = "chatbot-state.json"

// dataDir is the runtime data directory (config.Root()/data). Tests re-point it.
var dataDir = filepath.Join(config.Root(), "data")

// CursorStore is the daemon's view of the cursors: an in-memory map it reads and
// mutates, flushed by Save. Backed by the shared DB in production.
type CursorStore interface {
	Cursor(groupID int64) int64
	SetCursor(groupID, lastID int64)
	Save() error
}

// stateOnDisk is the legacy file format. Group cursors are keyed by group id string.
type stateOnDisk struct {
	Version int                         `json:"version"`
	Tenants map[string]map[string]int64 `json:"tenants"`
}

// cursorBackend persists the shared cursors. The MySQL implementation talks to
// the `ai` DB; tests swap in a fake via sharedBackend.
type cursorBackend interface {
	load(tenant string) (map[int64]int64, error)
	save(tenant string, groupID, lastID int64) error
}

// sharedBackend is the production backend; tests replace it with a fake.
var sharedBackend cursorBackend = mysqlBackend{}

type mysqlBackend struct{}

func (mysqlBackend) load(tenant string) (map[int64]int64, error) {
	return chatbotstate.LoadCursors(tenant)
}

func (mysqlBackend) save(tenant string, groupID, lastID int64) error {
	return chatbotstate.SaveCursor(tenant, groupID, lastID)
}

// dbCursorStore is the production CursorStore: the map is the working copy, Save
// flushes the entries changed since the last flush to the shared store. Save is
// safe to call more than once — only dirty groups are written.
type dbCursorStore struct {
	tenant  string
	backend cursorBackend
	cursors map[int64]int64
	dirty   map[int64]bool
}

// LoadCursorStore reads the tenant's shared cursors, seeding them once from the
// legacy local file when the shared store is still empty. A corrupt local file
// yields an error instead of silently resetting (a reset would make the bot
// answer every historical client message).
func LoadCursorStore(tenant string) (CursorStore, error) {
	cursors, err := sharedBackend.load(tenant)
	if err != nil {
		return nil, fmt.Errorf("chat: load shared cursors: %w", err)
	}
	s := &dbCursorStore{tenant: tenant, backend: sharedBackend, cursors: cursors, dirty: map[int64]bool{}}
	if len(cursors) == 0 {
		if err := s.importLegacy(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// importLegacy seeds the shared store from the legacy local file, if any. Each
// group is written through the monotonic backend, so a stale local file can never
// move a shared cursor backwards.
func (s *dbCursorStore) importLegacy() error {
	file, err := loadFileCursorStore(s.tenant, filepath.Join(dataDir, cursorFile))
	if err != nil {
		return err
	}
	for _, g := range file.Groups() {
		s.cursors[g] = file.Cursor(g)
		s.dirty[g] = true
	}
	return s.Save()
}

func (s *dbCursorStore) Cursor(groupID int64) int64 { return s.cursors[groupID] }

func (s *dbCursorStore) SetCursor(groupID, lastID int64) {
	s.cursors[groupID] = lastID
	s.dirty[groupID] = true
}

func (s *dbCursorStore) Save() error {
	for g := range s.dirty {
		if err := s.backend.save(s.tenant, g, s.cursors[g]); err != nil {
			return fmt.Errorf("chat: save cursor %d: %w", g, err)
		}
		delete(s.dirty, g)
	}
	return nil
}

// ---------------------------------------------------------------- legacy file store

// fileCursorStore is the legacy per-machine JSON cursor store, now used only to
// seed the shared store on first run. Its logic mirrors the previous state.go.
type fileCursorStore struct {
	tenant  string
	path    string
	cursors map[string]int64
}

// loadFileCursorStore reads the legacy file, or starts empty when it does not
// exist yet.
func loadFileCursorStore(tenant, path string) (*fileCursorStore, error) {
	s := &fileCursorStore{tenant: tenant, path: path, cursors: map[string]int64{}}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("chat: read cursor file: %w", err)
	}
	var disk stateOnDisk
	if err := json.Unmarshal(raw, &disk); err != nil {
		return nil, fmt.Errorf("chat: corrupt cursor file %s: %w", path, err)
	}
	if group := disk.Tenants[tenant]; group != nil {
		for k, v := range group {
			s.cursors[k] = v
		}
	}
	return s, nil
}

// Cursor returns the stored id for a group (0 = absent).
func (s *fileCursorStore) Cursor(groupID int64) int64 {
	return s.cursors[strconv.FormatInt(groupID, 10)]
}

// Groups returns the group ids present in the file.
func (s *fileCursorStore) Groups() []int64 {
	out := make([]int64, 0, len(s.cursors))
	for k := range s.cursors {
		if id, err := strconv.ParseInt(k, 10, 64); err == nil {
			out = append(out, id)
		}
	}
	return out
}
