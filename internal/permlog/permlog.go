package permlog

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Entry represents a single permission log event.
type Entry struct {
	Timestamp time.Time
	SessionID string
	Tool      string
}

var (
	tsRe   = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z)`)
	sidRe  = regexp.MustCompile(`session=(\S+)`)
	toolRe = regexp.MustCompile(`tool=(\S+)`)
)

// LogFile returns the default path to permission.log.
func LogFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".claude", "logs", "permission.log")
}

// Parse reads the permission.log and returns entries grouped by session ID.
func Parse(path string) (map[string][]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string][]Entry), nil
		}
		return nil, err
	}
	defer f.Close()

	result := make(map[string][]Entry)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		tsMatch := tsRe.FindStringSubmatch(line)
		sidMatch := sidRe.FindStringSubmatch(line)
		if tsMatch == nil || sidMatch == nil {
			continue
		}

		ts, err := parseTS(tsMatch[1])
		if err != nil {
			continue
		}

		tool := "unknown"
		if toolMatch := toolRe.FindStringSubmatch(line); toolMatch != nil {
			tool = toolMatch[1]
		}

		sid := sidMatch[1]
		result[sid] = append(result[sid], Entry{
			Timestamp: ts,
			SessionID: sid,
			Tool:      tool,
		})
	}

	return result, scanner.Err()
}

func parseTS(s string) (time.Time, error) {
	s = strings.Replace(s, "Z", "+00:00", 1)
	return time.Parse("2006-01-02T15:04:05+00:00", s)
}
