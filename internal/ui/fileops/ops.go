package fileops

import (
	"io"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

// DoneMsg is sent when a file operation succeeds.
type DoneMsg struct {
	Op     string
	Dir    string // destination directory to refresh
	SrcDir string // source directory to also refresh (move only)
}

// ErrMsg is sent when a file operation fails.
type ErrMsg struct {
	Op   string
	Text string
}

func DeleteCmd(path string) tea.Cmd {
	return func() tea.Msg {
		dir := filepath.Dir(path)
		if err := os.RemoveAll(path); err != nil {
			return ErrMsg{Op: "delete", Text: err.Error()}
		}
		return DoneMsg{Op: "delete", Dir: dir}
	}
}

func RenameCmd(oldPath, newName string) tea.Cmd {
	return func() tea.Msg {
		dir := filepath.Dir(oldPath)
		newPath := filepath.Join(dir, newName)
		if err := os.Rename(oldPath, newPath); err != nil {
			return ErrMsg{Op: "rename", Text: err.Error()}
		}
		return DoneMsg{Op: "rename", Dir: dir}
	}
}

func CreateFileCmd(dir, name string) tea.Cmd {
	return func() tea.Msg {
		path := filepath.Join(dir, name)
		f, err := os.Create(path)
		if err != nil {
			return ErrMsg{Op: "new file", Text: err.Error()}
		}
		f.Close()
		return DoneMsg{Op: "new file", Dir: dir}
	}
}

func CreateDirCmd(dir, name string) tea.Cmd {
	return func() tea.Msg {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(path, 0755); err != nil {
			return ErrMsg{Op: "new folder", Text: err.Error()}
		}
		return DoneMsg{Op: "new folder", Dir: dir}
	}
}

func CopyCmd(src, dstDir string) tea.Cmd {
	return func() tea.Msg {
		dst := filepath.Join(dstDir, filepath.Base(src))
		if err := copyPath(src, dst); err != nil {
			return ErrMsg{Op: "copy", Text: err.Error()}
		}
		return DoneMsg{Op: "copy", Dir: dstDir}
	}
}

func MoveCmd(src, dstDir string) tea.Cmd {
	return func() tea.Msg {
		srcDir := filepath.Dir(src)
		dst := filepath.Join(dstDir, filepath.Base(src))
		// try a cheap rename first (works within the same filesystem)
		if err := os.Rename(src, dst); err != nil {
			// cross-device: copy then delete
			if err2 := copyPath(src, dst); err2 != nil {
				return ErrMsg{Op: "move", Text: err2.Error()}
			}
			if err2 := os.RemoveAll(src); err2 != nil {
				return ErrMsg{Op: "move", Text: err2.Error()}
			}
		}
		return DoneMsg{Op: "move", Dir: dstDir, SrcDir: srcDir}
	}
}

func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyPath(
			filepath.Join(src, e.Name()),
			filepath.Join(dst, e.Name()),
		); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
