package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const todoSectionHeader = "## 実装タスク"

// RunTodoCleanup handles the SessionStart hook for TODO cleanup.
// On main branch, removes completed TODO tasks (all criteria checked) from TODO.md.
// 履歴は git log と GitHub Release で追跡する。
func RunTodoCleanup(input *HookInput) error {
	cwd := input.CWD
	if cwd == "" {
		cwd, _ = os.Getwd()
	}

	branch := getCurrentBranch(cwd)
	if branch != "main" {
		return nil
	}

	todoPath := filepath.Join(cwd, "TODO.md")

	content, err := os.ReadFile(todoPath)
	if err != nil || !strings.Contains(string(content), "[x]") {
		return nil
	}

	remaining, completedNames := ParseTodoAndExtract(string(content))
	if len(completedNames) == 0 {
		return nil
	}

	if err := os.WriteFile(todoPath, []byte(remaining), 0644); err != nil {
		return err
	}

	fmt.Printf("完了済みタスク %d 件を TODO.md から削除しました。\n", len(completedNames))
	return nil
}

func getCurrentBranch(cwd string) string {
	cmd := exec.Command("git", "-C", cwd, "branch", "--show-current")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// todoTask represents a task block being parsed in TODO.md.
type todoTask struct {
	lines     []string
	name      string
	criteria  int
	unchecked int
}

// ParseTodoAndExtract parses TODO.md content, removes tasks with all criteria
// checked from the "## 実装タスク" section, and returns remaining content and
// completed task names.
func ParseTodoAndExtract(content string) (remaining string, completedNames []string) {
	lines := strings.Split(content, "\n")

	var out []string
	var current *todoTask
	inSection := false

	flush := func() {
		if current == nil {
			return
		}
		if current.criteria > 0 && current.unchecked == 0 {
			completedNames = append(completedNames, current.name)
		} else {
			out = append(out, current.lines...)
		}
		current = nil
	}

	for _, line := range lines {
		if line == todoSectionHeader {
			inSection = true
			out = append(out, line)
			continue
		}

		if strings.HasPrefix(line, "## ") {
			if inSection {
				flush()
				inSection = false
			}
			out = append(out, line)
			continue
		}

		if !inSection {
			out = append(out, line)
			continue
		}

		if strings.HasPrefix(line, "- ") {
			flush()
			current = &todoTask{
				lines: []string{line},
				name:  strings.TrimPrefix(line, "- "),
			}
			continue
		}

		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "- ") && current != nil {
			current.lines = append(current.lines, line)
			if strings.Contains(line, "[x]") {
				current.criteria++
			}
			if strings.Contains(line, "[ ]") {
				current.criteria++
				current.unchecked++
			}
			continue
		}

		if strings.TrimSpace(line) == "" {
			if current != nil {
				flush()
			}
			out = append(out, line)
			continue
		}

		out = append(out, line)
	}

	if inSection {
		flush()
	}

	return strings.Join(out, "\n"), completedNames
}
