package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Beaver-family/beaver/internal/ai"
	"github.com/Beaver-family/beaver/internal/config"
	"github.com/Beaver-family/beaver/internal/ui/editor"
	"github.com/Beaver-family/beaver/internal/ui/fileops"
	"github.com/Beaver-family/beaver/internal/ui/filetree"
	"github.com/Beaver-family/beaver/internal/ui/git"
	"github.com/Beaver-family/beaver/internal/ui/preview"
	"github.com/Beaver-family/beaver/internal/ui/search"
	"github.com/Beaver-family/beaver/internal/ui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
)

const (
	focusSidebar = 0
	focusMain    = 1
)

const anthropicKeyMissingMsg = `No Anthropic API key found.

  Set it via environment variable:
    export ANTHROPIC_API_KEY=sk-ant-api03-...

  Or press ctrl+k to enter your key here.
  Get a free key at: console.anthropic.com`

const openAIKeyMissingMsg = `No OpenAI API key found.

  Set it via environment variable:
    export OPENAI_API_KEY=sk-...

  Or press ctrl+k to enter your key here.
  Get a key at: platform.openai.com/api-keys`

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case spinnerTickMsg:
		if m.chatStreaming {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(spinnerFrames)
			return m, spinnerTickCmd()
		}
		return m, nil

	case ai.StreamStartMsg:
		m.chatStreamCh = msg.Ch
		m.spinnerFrame = 0
		return m, tea.Batch(ai.ReadAgentEvent(m.chatStreamCh), spinnerTickCmd())

	case ai.StreamTokenMsg:
		m.chatBuf += msg.Text
		return m, ai.ReadAgentEvent(m.chatStreamCh)

	case ai.AgentToolStarted:
		// flush any text the model produced before this tool call so it appears first
		if m.chatBuf != "" {
			m.chatMessages = append(m.chatMessages, ai.Message{Role: "assistant", Content: m.chatBuf, Provider: ai.ProviderForModel(m.activeModel)})
			m.chatBuf = ""
		}
		m.chatMessages = append(m.chatMessages, ai.Message{Role: "tool", Content: "⚙ " + msg.Name})
		return m, ai.ReadAgentEvent(m.chatStreamCh)

	case ai.AgentToolDone:
		// update the matching pending tool message
		for i := len(m.chatMessages) - 1; i >= 0; i-- {
			if m.chatMessages[i].Role == "tool" && m.chatMessages[i].Content == "⚙ "+msg.Name {
				icon := "✓"
				if msg.IsError {
					icon = "✗"
				}
				m.chatMessages[i].Content = icon + " " + msg.Name + ": " + msg.Summary
				break
			}
		}
		// refresh filetree and git status when agent writes or creates files
		var extraCmd tea.Cmd
		if (msg.Name == "write_file" || msg.Name == "create_file") && m.filetree != nil {
			m.filetree.Refresh(m.filetree.Root.Path)
			extraCmd = git.LoadCmd(m.filetree.Root.Path)
		}
		return m, tea.Batch(ai.ReadAgentEvent(m.chatStreamCh), extraCmd)

	case ai.StreamDoneMsg:
		m.chatStreaming = false
		if m.chatBuf != "" {
			m.chatMessages = append(m.chatMessages, ai.Message{Role: "assistant", Content: m.chatBuf, Provider: ai.ProviderForModel(m.activeModel)})
			m.chatBuf = ""
		}
		m.chatScroll = 0
		return m, nil

	case validateKeyMsg:
		if msg.err != nil {
			m.apiKeyErr = friendlyAPIKeyErr(msg.err)
			m.apiKeyInput.Focus()
			return m, nil
		}
		if msg.provider == ai.ProviderOpenAI {
			if err := ai.SaveOpenAIKey(msg.key); err != nil {
				m.apiKeyErr = "could not save key: " + err.Error()
				return m, nil
			}
			m.openaiClient = ai.NewOpenAIClient(msg.key)
			if m.activeModel == "" || ai.ProviderForModel(m.activeModel) != ai.ProviderOpenAI {
				m.activeModel = ai.DefaultOpenAIModel
			}
		} else {
			if err := ai.SaveKey(msg.key); err != nil {
				m.apiKeyErr = "could not save key: " + err.Error()
				return m, nil
			}
			m.anthropicClient = ai.NewClient(msg.key)
			m.activeModel = ai.LoadSavedModel()
		}
		m.apiKeyMode = false
		m.apiKeyErr = ""
		if !m.chatMode {
			m.chatMode = true
			m.chatInput = newChatInput()
		}
		return m, nil

	case tea.KeyMsg:
		// ── theme selection ──────────────────────────────────────────────────
		if m.themeSelectMode {
			return m.updateThemeSelect(msg)
		}

		// ── model selection ──────────────────────────────────────────────────
		if m.modelSelectMode {
			return m.updateModelSelect(msg)
		}

		// ── api key setup ────────────────────────────────────────────────────
		if m.apiKeyMode {
			return m.updateAPIKey(msg)
		}

		// ── file op input/confirm mode ───────────────────────────────────────
		if m.opMode != opNone {
			return m.updateOp(msg)
		}

		// ── edit mode ────────────────────────────────────────────────────────
		if m.editMode {
			return m.updateEditor(msg)
		}

		// ── chat mode ────────────────────────────────────────────────────────
		if m.chatMode {
			return m.updateChat(msg)
		}

		// ── search / grep modes ──────────────────────────────────────────────
		if m.searchMode {
			return m.updateSearch(msg)
		}
		if m.grepMode {
			return m.updateGrep(msg)
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
				maxS := len(m.previewLines) - m.previewPageSize
				if maxS < 0 {
					maxS = 0
				}
				m.previewScroll = maxS
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

		// ── ai chat ──────────────────────────────────────────────────────────
		case "c":
			// always open chat — key setup is on-demand via ctrl+k inside chat
			provider := ai.ProviderForModel(m.activeModel)
			if provider == ai.ProviderOpenAI {
				if m.openaiClient == nil {
					if key := ai.ResolveOpenAIKey(); key != "" {
						m.openaiClient = ai.NewOpenAIClient(key)
					}
				}
			} else {
				if m.anthropicClient == nil {
					if key := ai.ResolveAPIKey(); key != "" {
						m.anthropicClient = ai.NewClient(key)
					}
				}
			}
			m.chatMode = true
			m.chatInput = newChatInput()
			return m, nil

		case "t":
			m.themeSelectMode = true
			m.themeSelectIdx = 0
			for i, key := range styles.ThemeNames {
				if styles.Themes[key].Name == styles.Current.Name {
					m.themeSelectIdx = i
					break
				}
			}
			return m, nil

		case "m":
			m.modelSelectMode = true
			// pre-select current model
			m.modelSelectIdx = 0
			for i, opt := range ai.Models {
				if opt.ID == m.activeModel {
					m.modelSelectIdx = i
					break
				}
			}
			return m, nil

		// ── search ───────────────────────────────────────────────────────────
		case "f", "/":
			if m.focus == focusSidebar && m.filetree != nil {
				m.searchInput = newOpInput("search files...")
				m.searchMode = true
				m.searchResults = []search.FileResult{}
				m.searchSelected = 0
			}

		case "ctrl+f":
			if m.filetree != nil {
				m.grepInput = newOpInput("search file contents...")
				m.grepMode = true
				m.grepResults = nil
				m.grepSelected = 0
				m.grepQuery = ""
				m.grepRunning = false
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
		if m.filetree != nil {
			return m, git.LoadCmd(m.filetree.Root.Path)
		}

	case fileops.ErrMsg:
		m.opErr = msg.Op + ": " + msg.Text

	case search.GrepDoneMsg:
		m.grepRunning = false
		m.grepQuery = msg.Query
		if msg.Results == nil {
			m.grepResults = []search.GrepResult{}
		} else {
			m.grepResults = msg.Results
		}
		m.grepSelected = 0

	case search.GrepErrMsg:
		m.grepRunning = false
		m.grepMode = false
		m.opErr = "grep: " + msg.Text

	case git.StatusMsg:
		if m.filetree != nil {
			m.filetree.GitStatus = msg.Status
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// content area = panelHeight - 1 header line = (height-2) - 1 = height - 3
		m.previewPageSize = msg.Height - 3
		if m.previewPageSize < 1 {
			m.previewPageSize = 1
		}
		// clamp scroll so resize never leaves it out of bounds
		if maxS := len(m.previewLines) - m.previewPageSize; m.previewScroll > maxS {
			if maxS < 0 {
				m.previewScroll = 0
			} else {
				m.previewScroll = maxS
			}
		}
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

// updateSearch handles keystrokes while fuzzy file search is active.
func (m *model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.searchResults = nil
		m.searchSelected = 0
		return m, nil
	case "enter":
		if m.searchSelected < len(m.searchResults) {
			r := m.searchResults[m.searchSelected]
			if !r.IsDir {
				m.searchMode = false
				m.searchResults = nil
				m.focus = focusMain
				return m, preview.Load(r.Path)
			}
		}
		return m, nil
	case "up", "k":
		if m.searchSelected > 0 {
			m.searchSelected--
		}
		return m, nil
	case "down", "j":
		if m.searchSelected < len(m.searchResults)-1 {
			m.searchSelected++
		}
		return m, nil
	default:
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		if m.filetree != nil {
			m.searchResults = search.FuzzyFiles(m.filetree.Root.Path, m.searchInput.Value())
			m.searchSelected = 0
		}
		return m, cmd
	}
}

// updateGrep handles keystrokes while content grep is active.
func (m *model) updateGrep(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// results navigation phase
	if m.grepResults != nil {
		switch msg.String() {
		case "esc":
			// back to input phase
			m.grepResults = nil
			m.grepSelected = 0
			return m, nil
		case "up", "k":
			if m.grepSelected > 0 {
				m.grepSelected--
			}
			return m, nil
		case "down", "j":
			if m.grepSelected < len(m.grepResults)-1 {
				m.grepSelected++
			}
			return m, nil
		case "enter":
			if m.grepSelected < len(m.grepResults) {
				path := m.grepResults[m.grepSelected].Path
				m.grepMode = false
				m.grepResults = nil
				m.focus = focusMain
				return m, preview.Load(path)
			}
			return m, nil
		}
		return m, nil
	}

	// input phase
	switch msg.String() {
	case "esc":
		m.grepMode = false
		m.grepRunning = false
		return m, nil
	case "enter":
		query := strings.TrimSpace(m.grepInput.Value())
		if query == "" || m.grepRunning {
			return m, nil
		}
		m.grepRunning = true
		m.grepQuery = query
		return m, search.GrepCmd(m.filetree.Root.Path, query)
	default:
		var cmd tea.Cmd
		m.grepInput, cmd = m.grepInput.Update(msg)
		return m, cmd
	}
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
		cmds := []tea.Cmd{preview.Load(m.editPath)}
		if m.filetree != nil {
			cmds = append(cmds, git.LoadCmd(m.filetree.Root.Path))
		}
		return m, tea.Batch(cmds...)

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

// ── api key setup ─────────────────────────────────────────────────────────────

type validateKeyMsg struct {
	key      string
	provider string
	err      error
}

func validateKeyCmd(key, provider string) tea.Cmd {
	return func() tea.Msg {
		var err error
		if provider == ai.ProviderOpenAI {
			err = ai.ValidateOpenAIKey(key)
		} else {
			err = ai.ValidateKey(key)
		}
		return validateKeyMsg{key: key, provider: provider, err: err}
	}
}

func (m *model) updateAPIKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.apiKeyMode = false
		m.apiKeyErr = ""
		return m, nil
	case "enter":
		key := strings.TrimSpace(m.apiKeyInput.Value())
		if key == "" {
			return m, nil
		}
		m.apiKeyErr = "validating..."
		return m, validateKeyCmd(key, m.apiKeyProvider)
	default:
		var cmd tea.Cmd
		m.apiKeyInput, cmd = m.apiKeyInput.Update(msg)
		return m, cmd
	}
}

func friendlyAPIKeyErr(err error) string {
	s := err.Error()
	switch {
	case strings.Contains(s, "401"), strings.Contains(s, "authentication_error"), strings.Contains(s, "invalid x-api-key"):
		return "API key not recognised — double-check it at console.anthropic.com"
	case strings.Contains(s, "403"):
		return "Access denied — your key may not have permission to use this API"
	case strings.Contains(s, "429"):
		return "Rate limited — wait a moment and try again"
	case strings.Contains(s, "no such host"), strings.Contains(s, "connection refused"), strings.Contains(s, "dial"):
		return "No internet connection — check your network and try again"
	default:
		return "Could not validate key — check your connection and try again"
	}
}

func newAPIKeyInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 200
	ti.Prompt = "> "
	ti.EchoMode = textinput.EchoPassword
	ti.Focus()
	return ti
}

// ── model selection ───────────────────────────────────────────────────────────

func (m *model) updateModelSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.modelSelectMode = false
		return m, nil
	case "up", "k":
		if m.modelSelectIdx > 0 {
			m.modelSelectIdx--
		}
		return m, nil
	case "down", "j":
		if m.modelSelectIdx < len(ai.Models)-1 {
			m.modelSelectIdx++
		}
		return m, nil
	case "enter":
		chosen := ai.Models[m.modelSelectIdx]
		m.activeModel = chosen.ID
		_ = ai.SaveModel(chosen.ID)
		m.modelSelectMode = false
		return m, nil
	}
	return m, nil
}

// ── theme selection ───────────────────────────────────────────────────────────

func (m *model) updateThemeSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.themeSelectMode = false
		return m, nil
	case "up", "k":
		if m.themeSelectIdx > 0 {
			m.themeSelectIdx--
		}
		return m, nil
	case "down", "j":
		if m.themeSelectIdx < len(styles.ThemeNames)-1 {
			m.themeSelectIdx++
		}
		return m, nil
	case "enter":
		key := styles.ThemeNames[m.themeSelectIdx]
		styles.SetTheme(key)
		_ = config.SaveTheme(key)
		m.themeSelectMode = false
		return m, nil
	}
	return m, nil
}

// ── ai chat ───────────────────────────────────────────────────────────────────

func (m *model) updateChat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+o":
		if !m.chatStreaming {
			m.modelSelectMode = true
			m.modelSelectIdx = 0
			for i, opt := range ai.Models {
				if opt.ID == m.activeModel {
					m.modelSelectIdx = i
					break
				}
			}
		}
		return m, nil

	case "esc":
		if m.chatStreaming {
			return m, nil // don't close while streaming
		}
		m.chatMode = false
		return m, nil

	case "up":
		m.chatScroll++
		return m, nil

	case "down":
		if m.chatScroll > 0 {
			m.chatScroll--
		}
		return m, nil

	case "ctrl+k":
		if !m.chatStreaming {
			m.apiKeyProvider = ai.ProviderForModel(m.activeModel)
			m.apiKeyMode = true
			m.apiKeyErr = ""
			if m.apiKeyProvider == ai.ProviderOpenAI {
				m.apiKeyInput = newAPIKeyInput("sk-...")
			} else {
				m.apiKeyInput = newAPIKeyInput("sk-ant-api03-...")
			}
		}
		return m, nil

	case "enter":
		text := strings.TrimSpace(m.chatInput.Value())
		if text == "" || m.chatStreaming {
			return m, nil
		}
		// guard: if no client, key isn't configured — show inline instructions
		provider := ai.ProviderForModel(m.activeModel)
		if provider == ai.ProviderOpenAI && m.openaiClient == nil {
			m.chatMessages = append(m.chatMessages, ai.Message{Role: "system", Content: openAIKeyMissingMsg})
			m.chatScroll = 0
			return m, nil
		}
		if provider != ai.ProviderOpenAI && m.anthropicClient == nil {
			m.chatMessages = append(m.chatMessages, ai.Message{Role: "system", Content: anthropicKeyMissingMsg})
			m.chatScroll = 0
			return m, nil
		}
		m.chatMessages = append(m.chatMessages, ai.Message{Role: "user", Content: text})
		m.chatInput.SetValue("")
		m.chatStreaming = true
		m.chatBuf = ""
		m.chatScroll = 0
		sys := m.buildSystemPrompt()
		root := m.filetree.Root.Path
		if provider == ai.ProviderOpenAI {
			return m, ai.OpenAIAgentCmd(m.openaiClient, m.chatMessages, sys, m.activeModel, root)
		}
		return m, ai.AgentCmd(m.anthropicClient, m.chatMessages, sys, m.activeModel, root)

	default:
		var cmd tea.Cmd
		m.chatInput, cmd = m.chatInput.Update(msg)
		return m, cmd
	}
}

// buildSystemPrompt builds context for Claude from the current file/directory.
func (m *model) buildSystemPrompt() string {
	var sb strings.Builder
	sb.WriteString("You are an AI agent embedded in Beaver, a terminal file manager.\n")
	sb.WriteString("You can read, write, and create files using the provided tools.\n")
	sb.WriteString("Prefer using tools to look at actual file contents before answering questions about code.\n")

	if m.filetree != nil {
		fmt.Fprintf(&sb, "Current directory: %s\n", m.filetree.Root.Path)
	}

	if m.previewPath != "" && len(m.previewLines) > 0 {
		fmt.Fprintf(&sb, "The user is currently viewing: %s\n", m.previewPath)
		sb.WriteString("File contents (first 200 lines):\n```\n")
		limit := len(m.previewLines)
		if limit > 200 {
			limit = 200
		}
		for _, line := range m.previewLines[:limit] {
			sb.WriteString(line)
			sb.WriteByte('\n')
		}
		sb.WriteString("```\n")
	}

	sb.WriteString("\nBe concise. Use markdown fences for code.")
	return sb.String()
}

func newChatInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "ask anything about this file or directory..."
	ti.CharLimit = 2000
	ti.Prompt = "> "
	ti.Focus()
	return ti
}
