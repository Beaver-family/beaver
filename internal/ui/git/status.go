package git

import (
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// FileState is the git status of a path.
// Values are ordered so that higher = more urgent (used for directory propagation).
type FileState uint8

const (
	StateClean     FileState = 0
	StateStaged    FileState = 1 // in index (green)
	StateModified  FileState = 2 // tracked, modified in worktree (yellow)
	StateUntracked FileState = 3 // not tracked by git (red)
)

// StatusMap maps absolute paths to their git state.
// Both files and their ancestor directories (up to repo root) are included.
type StatusMap map[string]FileState

// StatusMsg is delivered to the bubbletea model when the status load finishes.
type StatusMsg struct{ Status StatusMap }

// LoadCmd runs git status asynchronously and returns a StatusMsg.
func LoadCmd(repoPath string) tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{Status: Load(repoPath)}
	}
}

// Load runs git status and returns the StatusMap.
// Returns an empty map (not an error) if repoPath is not inside a git repo.
func Load(repoPath string) StatusMap {
	// resolve the actual repo root — repoPath may be a subdirectory
	out, err := exec.Command("git", "-C", repoPath, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return StatusMap{}
	}
	repoRoot := strings.TrimSpace(string(out))

	out, err = exec.Command("git", "-C", repoRoot, "status", "--porcelain").Output()
	if err != nil {
		return StatusMap{}
	}

	sm := make(StatusMap)
	for _, line := range strings.Split(string(out), "\n") {
		if len(line) < 4 {
			continue
		}
		x, y := line[0], line[1]
		rel := strings.TrimSpace(line[3:])

		// renames: "old -> new" — keep only the new (destination) path
		if i := strings.LastIndex(rel, " -> "); i >= 0 {
			rel = rel[i+4:]
		}
		// strip surrounding quotes git may add for unusual paths
		rel = strings.Trim(rel, `"`)

		absPath := filepath.Join(repoRoot, filepath.FromSlash(rel))
		sm[absPath] = fileState(x, y)
	}

	// propagate worst child state up to each ancestor directory
	dirs := make(map[string]FileState)
	repoRootSlash := repoRoot + string(filepath.Separator)
	for absPath, state := range sm {
		dir := filepath.Dir(absPath)
		for dir == repoRoot || strings.HasPrefix(dir, repoRootSlash) {
			if cur, ok := dirs[dir]; !ok || state > cur {
				dirs[dir] = state
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	for dir, state := range dirs {
		if cur, ok := sm[dir]; !ok || state > cur {
			sm[dir] = state
		}
	}

	return sm
}

// fileState maps the two-char porcelain XY code to a FileState.
func fileState(x, y byte) FileState {
	if x == '?' && y == '?' {
		return StateUntracked
	}
	// index has a change — staged
	if x != ' ' && x != '?' {
		return StateStaged
	}
	// worktree change only
	return StateModified
}
