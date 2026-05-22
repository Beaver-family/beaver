package ui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Beaver-family/tui/internal/ui/editor"
	"github.com/Beaver-family/tui/internal/ui/fileops"
	"github.com/Beaver-family/tui/internal/ui/filetree"
	"github.com/Beaver-family/tui/internal/ui/preview"
	"github.com/Beaver-family/tui/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
)

const (
	focusSidebar = 0
	focusMain    = 1
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		// ── file op input/confirm mode ───────────────────────────────────────
		if m.opMode != opNone {
			return m.updateOp(msg)
		}

		// ── edit mode ────────────────────────────────────────────────────────
		if m.editMode {
			return m.updateEditor(msg)
		}

		// ── normal mode ──────────────────────────────────────────────────────
		switch msg.String() {
		case "q":
			return m, tea.Quit

		case "tab":
			if m.focus == focusSidebar {
				m.focus = focusMain
			} else {
				m.focus = focusSidebar
			}

		case "esc", "left", "h":
			m.focus = focusSidebar

		case "up", "k":
			if m.focus == focusMain {
				if m.previewScroll > 0 {
					m.previewScroll--
				}
			} else if m.filetree != nil {
				m.filetree.Up()
				return m, loadSelected(m)
			}

		case "down", "j":
			if m.focus == focusMain {
				m.previewScroll++
			} else if m.filetree != nil {
				m.filetree.Down()
				return m, loadSelected(m)
			}

		case "ctrl+d":
			if m.focus == focusMain {
				m.previewScroll += m.previewPageSize / 2
			}
		case "ctrl+u":
			if m.focus == focusMain {
				half := m.previewPageSize / 2
				if m.previewScroll > half {
					m.previewScroll -= half
				} else {
					m.previewScroll = 0
				}
			}

		case "pgdown", " ":
			if m.focus == focusMain {
				m.previewScroll += m.previewPageSize
			}
		case "pgup":
			if m.focus == focusMain {
				if m.previewScroll > m.previewPageSize {
					m.previewScroll -= m.previewPageSize
				} else {
					m.previewScroll = 0
				}
			}

		case "g":
			if m.focus == focusMain {
				m.previewScroll = 0
			}
		case "G":
			if m.focus == focusMain {
				m.previewScroll = len(m.previewLines)
			}

		case "e":
			if m.focus == focusMain && m.previewPath != "" && m.previewErr == "" {
				return m, m.enterEditMode()
			}

		case "enter", "right", "l":
			if m.focus == focusSidebar && m.filetree != nil {
				node := m.filetree.SelectedNode()
				if node != nil && node.IsDir {
					m.filetree.Toggle(m.filetree.Selected)
				} else if node != nil {
					m.focus = focusMain
					return m, preview.Load(node.Path)
				}
			}

		// ── file operations ──────────────────────────────────────────────────
		case "d":
			if m.focus == focusSidebar && m.filetree != nil {
				if node := m.filetree.SelectedNode(); node != nil {
					m.opMode = opConfirmDelete
					m.opErr = ""
				}
			}

		case "r":
			if m.focus == focusSidebar && m.filetree != nil {
				if node := m.filetree.SelectedNode(); node != nil {
					m.opInput = newOpInput("new name")
					m.opInput.SetValue(node.Name)
					m.opMode = opInputRename
					m.opErr = ""
				}
			}

		case "n":
			if m.focus == focusSidebar && m.filetree != nil {
				m.opInput = newOpInput("filename.txt")
				m.opMode = opInputNewFile
				m.opErr = ""
			}

		case "N":
			if m.focus == focusSidebar && m.filetree != nil {
				m.opInput = newOpInput("folder-name")
				m.opMode = opInputNewDir
				m.opErr = ""
			}

		case "y":
			if m.focus == focusSidebar && m.filetree != nil {
				if node := m.filetree.SelectedNode(); node != nil {
					m.clipboard = node
					m.clipCut = false
					m.filetree.ClipboardPath = node.Path
					m.filetree.ClipboardCut = false
					m.opErr = ""
				}
			}

		case "x":
			if m.focus == focusSidebar && m.filetree != nil {
				if node := m.filetree.SelectedNode(); node != nil {
					m.clipboard = node
					m.clipCut = true
					m.filetree.ClipboardPath = node.Path
					m.filetree.ClipboardCut = true
					m.opErr = ""
				}
			}

		case "p":
			if m.focus == focusSidebar && m.filetree != nil && m.clipboard != nil {
				node := m.filetree.SelectedNode()
				m.opPasteDst = nodeDir(node, m.filetree.Root)
				m.opMode = opConfirmPaste
				m.opErr = ""
			}
		}

	case editReadyMsg:
		w, h := editorDimensions(m)
		m.editor = editor.New(msg.content, w, h)
		m.editPath = msg.path
		m.editMode = true
		m.editModified = false
		m.editErr = ""

	case preview.LoadedMsg:
		m.previewLines = msg.Lines
		m.previewPath = msg.Path
		m.previewScroll = 0
		m.previewErr = ""

	case preview.ErrMsg:
		m.previewLines = nil
		m.previewPath = msg.Path
		m.previewScroll = 0
		m.previewErr = msg.Text

	case fileops.DoneMsg:
		m.opErr = ""
		if m.filetree != nil {
			m.filetree.Refresh(msg.Dir)
			// for moves, also refresh the source directory
			if msg.SrcDir != "" && msg.SrcDir != msg.Dir {
				m.filetree.Refresh(msg.SrcDir)
			}
		}
		// clear preview if the file it was showing no longer exists
		if m.previewPath != "" {
			if _, err := os.Stat(m.previewPath); err != nil {
				m.previewPath = ""
				m.previewLines = nil
				m.previewErr = ""
			}
		}

	case fileops.ErrMsg:
		m.opErr = msg.Op + ": " + msg.Text

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.previewPageSize = msg.Height - 4
		if m.editMode {
			w, h := editorDimensions(m)
			editor.Resize(&m.editor, w, h)
		}

	case tickMsg:
		m.now = msg.Time()
		return m, tickCmd()
	}

	return m, nil
}

// updateOp handles key events while a file operation prompt is active.
func (m *model) updateOp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.opMode {
	case opConfirmDelete:
		switch msg.String() {
		case "y", "Y":
			node := m.filetree.SelectedNode()
			m.opMode = opNone
			if node == nil {
				return m, nil
			}
			return m, fileops.DeleteCmd(node.Path)
		case "n", "N", "esc":
			m.opMode = opNone
		}
		return m, nil

	case opConfirmPaste:
		switch msg.String() {
		case "y", "Y":
			src := m.clipboard.Path
			isCut := m.clipCut
			dst := m.opPasteDst
			m.opMode = opNone
			if isCut {
				// clear clipboard — a cut can only be pasted once
				m.clipboard = nil
				m.clipCut = false
				m.filetree.ClipboardPath = ""
			}
			if isCut {
				return m, fileops.MoveCmd(src, dst)
			}
			return m, fileops.CopyCmd(src, dst)
		case "n", "N", "esc":
			m.opMode = opNone
		}
		return m, nil

	case opInputRename, opInputNewFile, opInputNewDir:
		switch msg.String() {
		case "enter":
			val := strings.TrimSpace(m.opInput.Value())
			mode := m.opMode
			m.opMode = opNone
			if val == "" {
				return m, nil
			}
			return m, m.executeInputOp(mode, val)
		case "esc":
			m.opMode = opNone
			return m, nil
		default:
			var cmd tea.Cmd
			m.opInput, cmd = m.opInput.Update(msg)
			return m, cmd
		}
	}
	return m, nil
}

// executeInputOp dispatches the appropriate file operation after the user
// confirms an input prompt.
func (m *model) executeInputOp(mode opMode, val string) tea.Cmd {
	node := m.filetree.SelectedNode()
	switch mode {
	case opInputRename:
		if node == nil {
			return nil
		}
		return fileops.RenameCmd(node.Path, val)
	case opInputNewFile:
		return fileops.CreateFileCmd(nodeDir(node, m.filetree.Root), val)
	case opInputNewDir:
		return fileops.CreateDirCmd(nodeDir(node, m.filetree.Root), val)
	}
	return nil
}

// nodeDir returns the directory to use for operations: if node is a directory
// use it directly, otherwise use its parent. Falls back to tree root.
func nodeDir(node *filetree.Node, root *filetree.Node) string {
	if node == nil {
		return root.Path
	}
	if node.IsDir {
		return node.Path
	}
	return filepath.Dir(node.Path)
}

// newOpInput creates a focused single-line text input for file op prompts.
func newOpInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 255
	ti.Prompt = "> "
	ti.Focus()
	return ti
}

// updateEditor handles all key events while in edit mode.
func (m *model) updateEditor(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// discard-confirm sub-mode
	if m.editConfirm {
		switch msg.String() {
		case "y", "Y":
			m.exitEditMode(false)
		case "n", "N", "esc":
			m.editConfirm = false
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+s":
		if err := saveFile(m.editPath, m.editor.Value()); err != nil {
			m.editErr = "save failed: " + err.Error()
			return m, nil
		}
		m.exitEditMode(true)
		return m, preview.Load(m.editPath)

	case "esc":
		if m.editModified {
			m.editConfirm = true
			return m, nil
		}
		m.exitEditMode(false)
		return m, nil
	}

	m.editErr = ""
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	m.editModified = true
	return m, cmd
}

// enterEditMode reads the raw file and opens it in the editor.
func (m *model) enterEditMode() tea.Cmd {
	return func() tea.Msg {
		raw, err := os.ReadFile(m.previewPath)
		if err != nil {
			return preview.ErrMsg{Path: m.previewPath, Text: "cannot open for editing: " + err.Error()}
		}
		return editReadyMsg{path: m.previewPath, content: string(raw)}
	}
}

type editReadyMsg struct {
	path    string
	content string
}

// exitEditMode tears down editor state.
func (m *model) exitEditMode(saved bool) {
	m.editMode = false
	m.editModified = false
	m.editConfirm = false
	m.editErr = ""
	_ = saved
}

// editorDimensions computes usable width/height for the textarea.
func editorDimensions(m *model) (int, int) {
	sw := styles.SidebarWidth()
	mainWidth := m.width - sw - 2
	if mainWidth < 4 {
		mainWidth = 4
	}
	h := m.height - 4
	if h < 1 {
		h = 1
	}
	return mainWidth - 2, h
}

// saveFile writes content back to disk preserving original permissions.
func saveFile(path, content string) error {
	perm := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode()
	}
	return os.WriteFile(path, []byte(content), perm)
}

// loadSelected auto-previews files as the cursor moves (files only, not dirs).
func loadSelected(m *model) tea.Cmd {
	node := m.filetree.SelectedNode()
	if node == nil || node.IsDir {
		return nil
	}
	return preview.Load(node.Path)
}
