package ui

import (
	"os"

	"github.com/Beaver-family/tui/internal/ui/editor"
	"github.com/Beaver-family/tui/internal/ui/preview"
	"github.com/Beaver-family/tui/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	focusSidebar = 0
	focusMain    = 1
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
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
			// enter edit mode — only when a file is open in the preview
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

	// clear any previous save error on next keystroke
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
	mainWidth := m.width - sw - styles.RightPanelWidth - 4
	if mainWidth < 4 {
		mainWidth = 4
	}
	// subtract 2: header line + padding
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
