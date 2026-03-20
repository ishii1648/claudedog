package transcript

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempTranscript(t *testing.T, lines []string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "transcript.jsonl")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range lines {
		f.WriteString(l + "\n")
	}
	f.Close()
	return p
}

func TestParse_BasicStats(t *testing.T) {
	p := writeTempTranscript(t, []string{
		`{"type":"user","message":{"content":"hello"}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read"},{"type":"tool_use","name":"Edit"}]}}`,
		`{"type":"user","message":{"content":"fix this"}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"ask-user-question"}]}}`,
	})

	stats := Parse(p)
	if stats.ToolUseTotal != 3 {
		t.Errorf("tool_use_total: got %d, want 3", stats.ToolUseTotal)
	}
	if stats.MidSessionMsgs != 1 {
		t.Errorf("mid_session_msgs: got %d, want 1", stats.MidSessionMsgs)
	}
	if stats.AskUserQuestion != 1 {
		t.Errorf("ask_user_question: got %d, want 1", stats.AskUserQuestion)
	}
	if stats.IsGhost {
		t.Error("should not be ghost")
	}
}

func TestParse_GhostSession(t *testing.T) {
	// No user messages at all
	p := writeTempTranscript(t, []string{
		`{"type":"assistant","message":{"content":[{"type":"text","text":"hello"}]}}`,
	})

	stats := Parse(p)
	if !stats.IsGhost {
		t.Error("should be ghost")
	}
}

func TestParse_ExcludesToolResultOnly(t *testing.T) {
	p := writeTempTranscript(t, []string{
		`{"type":"user","message":{"content":"initial prompt"}}`,
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Read"}]}}`,
		`{"type":"user","message":{"content":[{"type":"tool_result"}]}}`,
		`{"type":"user","message":{"content":"real message"}}`,
	})

	stats := Parse(p)
	// tool_result-only message should not count as mid-session
	if stats.MidSessionMsgs != 1 {
		t.Errorf("mid_session_msgs: got %d, want 1 (tool_result excluded)", stats.MidSessionMsgs)
	}
}

func TestParse_ExcludesLocalCommand(t *testing.T) {
	p := writeTempTranscript(t, []string{
		`{"type":"user","message":{"content":"initial"}}`,
		`{"type":"user","message":{"content":"<local-command-foo>bar</local-command-foo>"}}`,
	})

	stats := Parse(p)
	if stats.MidSessionMsgs != 0 {
		t.Errorf("mid_session_msgs: got %d, want 0 (local-command excluded)", stats.MidSessionMsgs)
	}
}

func TestParse_NonExistentFile(t *testing.T) {
	stats := Parse("/nonexistent/path/transcript.jsonl")
	if !stats.IsGhost {
		t.Error("non-existent file should produce ghost stats")
	}
}

func TestParse_EmptyPath(t *testing.T) {
	stats := Parse("")
	if !stats.IsGhost {
		t.Error("empty path should produce ghost stats")
	}
}
