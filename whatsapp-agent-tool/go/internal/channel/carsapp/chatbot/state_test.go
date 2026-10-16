package chatbot

import (
	"os"
	"path/filepath"
	"testing"
)

// fakeBackend is an in-memory cursorBackend that persists across LoadCursorStore
// calls within a test, so the shared-store logic can be exercised without a DB.
type fakeBackend struct {
	// data[tenant][groupID] = lastID
	data map[string]map[int64]int64
	// saves counts every save call (to assert dirty-only flushing).
	saves int
}

func newFakeBackend() *fakeBackend {
	return &fakeBackend{data: map[string]map[int64]int64{}}
}

func (f *fakeBackend) load(tenant string) (map[int64]int64, error) {
	out := map[int64]int64{}
	for k, v := range f.data[tenant] {
		out[k] = v
	}
	return out, nil
}

func (f *fakeBackend) save(tenant string, groupID, lastID int64) error {
	f.saves++
	if f.data[tenant] == nil {
		f.data[tenant] = map[int64]int64{}
	}
	if cur, ok := f.data[tenant][groupID]; !ok || lastID > cur {
		f.data[tenant][groupID] = lastID
	}
	return nil
}

// withFakeBackend swaps in a fake shared backend for the duration of a test.
func withFakeBackend(t *testing.T) *fakeBackend {
	t.Helper()
	prev := sharedBackend
	f := newFakeBackend()
	sharedBackend = f
	t.Cleanup(func() { sharedBackend = prev })
	return f
}

func TestSharedCursorStoreRoundTrip(t *testing.T) {
	restore, _ := useTempDataDir(t)
	defer restore()
	withFakeBackend(t)

	s, err := LoadCursorStore("mig_26")
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Cursor(12); got != 0 {
		t.Fatalf("fresh store cursor = %d, want 0", got)
	}
	s.SetCursor(12, 345)
	s.SetCursor(7, 99)
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	s2, err := LoadCursorStore("mig_26")
	if err != nil {
		t.Fatal(err)
	}
	if got := s2.Cursor(12); got != 345 {
		t.Fatalf("after reload cursor 12 = %d, want 345", got)
	}
	if got := s2.Cursor(7); got != 99 {
		t.Fatalf("after reload cursor 7 = %d, want 99", got)
	}

	// A second tenant must not see the first tenant's cursors.
	s3, err := LoadCursorStore("other_tenant")
	if err != nil {
		t.Fatal(err)
	}
	if got := s3.Cursor(12); got != 0 {
		t.Fatalf("tenant isolation broken: cursor = %d, want 0", got)
	}
}

// A shared cursor must never move backwards, even if a stale writer saves an
// older value (a machine that fell behind, or two importing at once).
func TestSharedCursorMonotonic(t *testing.T) {
	restore, _ := useTempDataDir(t)
	defer restore()
	f := withFakeBackend(t)

	s, _ := LoadCursorStore("mig_26")
	s.SetCursor(12, 500)
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	// A stale writer for the same group must not regress the cursor.
	if err := f.save("mig_26", 12, 400); err != nil {
		t.Fatal(err)
	}
	if got := f.data["mig_26"][12]; got != 500 {
		t.Fatalf("cursor regressed to %d, want 500", got)
	}
}

// The first load of an empty shared store seeds it from the legacy local file.
func TestSharedStoreImportsLegacyFile(t *testing.T) {
	restore, dir := useTempDataDir(t)
	defer restore()
	f := withFakeBackend(t)

	legacy := `{"version":1,"tenants":{"mig_26":{"12":345,"7":99}}}`
	if err := os.WriteFile(filepath.Join(dir, cursorFile), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadCursorStore("mig_26")
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Cursor(12); got != 345 {
		t.Fatalf("imported cursor 12 = %d, want 345", got)
	}
	if got := f.data["mig_26"][12]; got != 345 {
		t.Fatalf("legacy cursor not persisted to shared store: %d", got)
	}
	if got := f.data["mig_26"][7]; got != 99 {
		t.Fatalf("legacy cursor 7 not persisted: %d", got)
	}
}

// A non-empty shared store must NOT re-import the local file (the DB wins).
func TestSharedStoreSkipsImportWhenPopulated(t *testing.T) {
	restore, dir := useTempDataDir(t)
	defer restore()
	f := withFakeBackend(t)
	f.data["mig_26"] = map[int64]int64{12: 900}

	legacy := `{"version":1,"tenants":{"mig_26":{"12":345}}}`
	if err := os.WriteFile(filepath.Join(dir, cursorFile), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	s, err := LoadCursorStore("mig_26")
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Cursor(12); got != 900 {
		t.Fatalf("shared cursor = %d, want 900 (DB must win over the stale file)", got)
	}
}

func TestLegacyCursorStoreCorruptFileFails(t *testing.T) {
	restore, dir := useTempDataDir(t)
	defer restore()

	if err := os.WriteFile(filepath.Join(dir, cursorFile), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := loadFileCursorStore("mig_26", filepath.Join(dir, cursorFile)); err == nil {
		t.Fatal("corrupt cursor file must error, not silently reset")
	}
}

// A corrupt legacy file must fail the import (an empty shared store + corrupt
// file must not silently start over, which would re-answer history).
func TestSharedStoreCorruptLegacyFails(t *testing.T) {
	restore, dir := useTempDataDir(t)
	defer restore()
	withFakeBackend(t)

	if err := os.WriteFile(filepath.Join(dir, cursorFile), []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCursorStore("mig_26"); err == nil {
		t.Fatal("corrupt legacy file must error on import")
	}
}

// useTempDataDir points the package dataDir at a fresh temp dir and restores
// it afterwards, returning the temp dir for file placement.
func useTempDataDir(t *testing.T) (restore func(), dir string) {
	t.Helper()
	prev := dataDir
	dir = t.TempDir()
	dataDir = dir
	return func() { dataDir = prev }, dir
}
