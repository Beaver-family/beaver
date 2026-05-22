package search

import (
	"os"
	"path/filepath"
	"strings"
)

const maxFuzzyResults = 100

// FileResult is a file or directory matched by fuzzy search.
type FileResult struct {
	Name  string
	Path  string
	IsDir bool
}

// FuzzyFiles walks the entire tree under rootPath and returns files whose
// name contains query (case-insensitive). Returns empty slice for empty query.
func FuzzyFiles(rootPath, query string) []FileResult {
	if query == "" {
		return []FileResult{}
	}
	out := make([]FileResult, 0, 32)
	walkForFuzzy(rootPath, strings.ToLower(query), &out)
	return out
}

func walkForFuzzy(dir, query string, out *[]FileResult) {
	if len(*out) >= maxFuzzyResults {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, e := range entries {
		if len(*out) >= maxFuzzyResults {
			return
		}
		if e.Name()[0] == '.' {
			continue
		}
		fullPath := filepath.Join(dir, e.Name())
		if strings.Contains(strings.ToLower(e.Name()), query) {
			*out = append(*out, FileResult{
				Name:  e.Name(),
				Path:  fullPath,
				IsDir: e.IsDir(),
			})
		}
		if e.IsDir() {
			walkForFuzzy(fullPath, query, out)
		}
	}
}
