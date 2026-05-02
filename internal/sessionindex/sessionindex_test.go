package sessionindex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestWriteAll_AtomicNoTempLeftover ensures successful WriteAll does not
// leave behind any .tmp files in the destination directory.
func TestWriteAll_AtomicNoTempLeftover(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "session-index.jsonl")

	raws := []json.RawMessage{
		json.RawMessage(`{"session_id":"s1"}`),
		json.RawMessage(`{"session_id":"s2"}`),
	}
	if err := WriteAll(path, raws); err != nil {
		t.Fatalf("WriteAll: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".tmp") {
			t.Errorf("temp file leftover: %s", e.Name())
		}
	}
}

// TestWriteAll_FailurePreservesExisting verifies that when WriteAll fails to
// create / write the temp file, the existing session-index.jsonl is left
// untouched (not truncated, not partially overwritten).
func TestWriteAll_FailurePreservesExisting(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("read-only directory mode does not block root; skipping")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "session-index.jsonl")

	original := `{"session_id":"keep-me","pr_urls":["u1","u2"]}` + "\n"
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	// Make the directory read-only so os.CreateTemp inside WriteAll fails.
	if err := os.Chmod(dir, 0555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0755) })

	err := WriteAll(path, []json.RawMessage{json.RawMessage(`{"session_id":"new"}`)})
	if err == nil {
		t.Fatal("WriteAll: expected error in read-only directory")
	}

	// Restore write perm so we can read the file back (read should already work).
	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != original {
		t.Errorf("existing file was modified after failed WriteAll\ngot:  %q\nwant: %q", string(got), original)
	}
}
