package search

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	maxGrepResults  = 500
	maxGrepFileSize = 512 * 1024 // 500 KB
)

// GrepResult is a single matching line found during content search.
type GrepResult struct {
	Path    string
	LineNum int
	Line    string
}

// GrepDoneMsg is sent when content search completes successfully.
type GrepDoneMsg struct {
	Query   string
	Results []GrepResult
}

// GrepErrMsg is sent when content search fails entirely.
type GrepErrMsg struct {
	Text string
}

var errGrepDone = errors.New("grep cap reached")

// GrepCmd asynchronously searches all text files under rootPath for query.
func GrepCmd(rootPath, query string) tea.Cmd {
	return func() tea.Msg {
		results := make([]GrepResult, 0, 32)
		if query == "" {
			return GrepDoneMsg{Query: query, Results: results}
		}
		err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if d.Name()[0] == '.' {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Name()[0] == '.' {
				return nil
			}
			info, err := d.Info()
			if err != nil || info.Size() > maxGrepFileSize {
				return nil
			}
			matches := grepFile(path, query)
			results = append(results, matches...)
			if len(results) >= maxGrepResults {
				results = results[:maxGrepResults]
				return errGrepDone
			}
			return nil
		})
		if err != nil && !errors.Is(err, errGrepDone) {
			return GrepErrMsg{Text: err.Error()}
		}
		return GrepDoneMsg{Query: query, Results: results}
	}
}

func grepFile(path, query string) []GrepResult {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	lower := strings.ToLower(query)
	var results []GrepResult
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if !utf8.ValidString(line) {
			return nil // binary file — bail out
		}
		if strings.Contains(strings.ToLower(line), lower) {
			results = append(results, GrepResult{
				Path:    path,
				LineNum: lineNum,
				Line:    strings.TrimSpace(line),
			})
		}
	}
	return results
}
