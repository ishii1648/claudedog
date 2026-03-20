package syncdb

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestRunWithPaths(t *testing.T) {
	dir := t.TempDir()

	// Create transcript files so sessions are not ghost
	t1Path := filepath.Join(dir, "t1.jsonl")
	os.WriteFile(t1Path, []byte(
		`{"type":"user","message":{"content":"hello"}}`+"\n"+
			`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read"}]}}`+"\n",
	), 0644)
	t2Path := filepath.Join(dir, "t2.jsonl")
	os.WriteFile(t2Path, []byte(
		`{"type":"user","message":{"content":"hello"}}`+"\n"+
			`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Edit"}]}}`+"\n",
	), 0644)
	t3Path := filepath.Join(dir, "t3.jsonl")
	os.WriteFile(t3Path, []byte(
		`{"type":"user","message":{"content":"hello"}}`+"\n",
	), 0644)

	// Create session-index.jsonl
	indexPath := filepath.Join(dir, "session-index.jsonl")
	os.WriteFile(indexPath, []byte(
		`{"timestamp":"2026-03-01 10:00:00","session_id":"s1","cwd":"/tmp","repo":"user/repo","branch":"feat","pr_urls":["https://github.com/user/repo/pull/1"],"transcript":"`+t1Path+`","parent_session_id":"","backfill_checked":false}`+"\n"+
			`{"timestamp":"2026-03-01 11:00:00","session_id":"s2","cwd":"/tmp","repo":"user/repo","branch":"feat","pr_urls":["https://github.com/user/repo/pull/1"],"transcript":"`+t2Path+`","parent_session_id":"","backfill_checked":false}`+"\n"+
			`{"timestamp":"2026-03-01 12:00:00","session_id":"s3","cwd":"/tmp","repo":"ishii1648/dotfiles","branch":"main","pr_urls":["https://github.com/ishii1648/dotfiles/pull/5"],"transcript":"`+t3Path+`","parent_session_id":"","backfill_checked":false}`+"\n",
	), 0644)

	// Create permission.log
	permPath := filepath.Join(dir, "permission.log")
	os.WriteFile(permPath, []byte(
		"2026-03-01T10:05:00Z session=s1 tool=Bash(git:internal)\n"+
			"2026-03-01T10:10:00Z session=s1 tool=Edit\n"+
			"2026-03-01T11:05:00Z session=s2 tool=Write\n",
	), 0644)

	dbPath := filepath.Join(dir, "claudedog.db")
	err := RunWithPaths(indexPath, permPath, dbPath)
	if err != nil {
		t.Fatal(err)
	}

	// Verify DB
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Check sessions count
	var sessionCount int
	db.QueryRow("SELECT COUNT(*) FROM sessions").Scan(&sessionCount)
	if sessionCount != 3 {
		t.Errorf("sessions count: got %d, want 3", sessionCount)
	}

	// Check permission_events count
	var permCount int
	db.QueryRow("SELECT COUNT(*) FROM permission_events").Scan(&permCount)
	if permCount != 3 {
		t.Errorf("permission_events count: got %d, want 3", permCount)
	}

	// Check transcript_stats count
	var statsCount int
	db.QueryRow("SELECT COUNT(*) FROM transcript_stats").Scan(&statsCount)
	if statsCount != 3 {
		t.Errorf("transcript_stats count: got %d, want 3", statsCount)
	}

	// Check pr_metrics VIEW excludes dotfiles repo
	var prMetricsCount int
	db.QueryRow("SELECT COUNT(*) FROM pr_metrics").Scan(&prMetricsCount)
	if prMetricsCount != 1 {
		t.Errorf("pr_metrics count: got %d, want 1 (dotfiles excluded)", prMetricsCount)
	}

	// Check pr_metrics aggregation
	var prURL string
	var sessCount, permTotal int
	db.QueryRow("SELECT pr_url, session_count, perm_count FROM pr_metrics").Scan(&prURL, &sessCount, &permTotal)
	if prURL != "https://github.com/user/repo/pull/1" {
		t.Errorf("pr_url: got %s", prURL)
	}
	if sessCount != 2 {
		t.Errorf("session_count: got %d, want 2", sessCount)
	}
	if permTotal != 3 {
		t.Errorf("perm_count: got %d, want 3", permTotal)
	}
}

func TestRunWithPaths_DummyPRURL(t *testing.T) {
	dir := t.TempDir()

	indexPath := filepath.Join(dir, "session-index.jsonl")
	os.WriteFile(indexPath, []byte(
		`{"timestamp":"2026-03-01 10:00:00","session_id":"s1","cwd":"/tmp","repo":"user/repo","branch":"feat","pr_urls":["https://github.com/org/repo/pull/123"],"transcript":"","parent_session_id":"","backfill_checked":false}`+"\n",
	), 0644)

	permPath := filepath.Join(dir, "permission.log")
	os.WriteFile(permPath, []byte(""), 0644)

	dbPath := filepath.Join(dir, "claudedog.db")
	err := RunWithPaths(indexPath, permPath, dbPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	// Dummy PR URL should be stored as empty
	var prURL string
	db.QueryRow("SELECT pr_url FROM sessions WHERE session_id = 's1'").Scan(&prURL)
	if prURL != "" {
		t.Errorf("dummy PR URL should be empty, got: %s", prURL)
	}
}

func TestShortenPRURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://github.com/user/repo/pull/42", "user/repo#42"},
		{"https://example.com/foo", "https://example.com/foo"},
	}
	for _, tt := range tests {
		got := ShortenPRURL(tt.input)
		if got != tt.want {
			t.Errorf("ShortenPRURL(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
