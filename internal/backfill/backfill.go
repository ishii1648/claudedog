package backfill

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/ishii1648/claudedog/internal/sessionindex"
)

type group struct {
	repo    string
	branch  string
	entries []sessionindex.Session
}

type result struct {
	group       group
	url         string
	markChecked bool
}

// Run executes the backfill batch. It finds sessions without pr_urls,
// groups them by (repo, branch), and fetches PR URLs via gh pr list in parallel.
func Run(indexPath string, recheck bool) error {
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		return nil
	}

	_, sessions, err := sessionindex.ReadAll(indexPath)
	if err != nil {
		return err
	}

	// Collect entries with empty pr_urls
	var entries []sessionindex.Session
	for _, s := range sessions {
		if len(s.PRURLs) == 0 && (!s.BackfillChecked || recheck) {
			entries = append(entries, s)
		}
	}

	if len(entries) == 0 {
		fmt.Println("backfill: 対象エントリなし（全件 pr_urls 補完済み or backfill_checked 済み）")
		return nil
	}

	// Group by (repo, branch)
	type key struct{ repo, branch string }
	groupMap := make(map[key][]sessionindex.Session)
	for _, e := range entries {
		if e.Repo == "" || e.Branch == "" {
			continue
		}
		k := key{e.Repo, e.Branch}
		groupMap[k] = append(groupMap[k], e)
	}

	var groups []group
	for k, es := range groupMap {
		groups = append(groups, group{repo: k.repo, branch: k.branch, entries: es})
	}

	fmt.Printf("backfill: %d エントリ / %d グループを処理中...\n", len(entries), len(groups))

	// Parallel fetch with max 8 workers
	results := make(chan result, len(groups))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8)

	for _, g := range groups {
		wg.Add(1)
		go func(g group) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			r := fetchPRURL(g)
			results <- r
		}(g)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	found, skipped, retried := 0, 0, 0
	for r := range results {
		if r.url != "" {
			found++
			for _, e := range r.group.entries {
				if e.SessionID != "" {
					sessionindex.Update(indexPath, e.SessionID, []string{r.url})
				}
			}
		} else if r.markChecked {
			skipped++
			var ids []string
			for _, e := range r.group.entries {
				if e.SessionID != "" {
					ids = append(ids, e.SessionID)
				}
			}
			if len(ids) > 0 {
				sessionindex.MarkChecked(indexPath, ids)
			}
		} else {
			retried++
		}
	}

	fmt.Printf("backfill: 完了 — URL取得成功 %d グループ / cwd消滅スキップ %d グループ / 再試行待ち %d グループ\n",
		found, skipped, retried)
	return nil
}

func fetchPRURL(g group) result {
	// Use the last entry's cwd (matches Python behavior)
	cwd := g.entries[len(g.entries)-1].CWD
	if cwd == "" || !isDir(cwd) {
		return result{group: g, markChecked: true}
	}

	cmd := exec.Command("gh", "pr", "list",
		"--head", g.branch,
		"--author", "@me",
		"--state", "all",
		"--json", "url",
		"-q", ".[0].url",
	)
	cmd.Dir = cwd

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()

	select {
	case err := <-done:
		url := strings.TrimSpace(stdout.String())
		if err != nil || !strings.Contains(url, "github.com") {
			return result{group: g}
		}
		return result{group: g, url: url}
	case <-time.After(8 * time.Second):
		cmd.Process.Kill()
		return result{group: g}
	}
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
