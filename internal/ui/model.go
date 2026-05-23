package ui

import (
	"time"

	"github.com/Beaver-family/beaver/internal/ai"
	"github.com/Beaver-family/beaver/internal/ui/filetree"
	"github.com/Beaver-family/beaver/internal/ui/search"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
)

type opMode int

const (
	opNone opMode = iota
	opConfirmDelete
	opConfirmPaste
	opInputRename
	opInputNewFile
	opInputNewDir
)

type model struct {
	width    int
	height   int
	now      time.Time
	filetree *filetree.Tree

	// preview
	previewLines    []string
	previewPath     string
	previewScroll   int
	previewPageSize int
	previewErr      string
	focus           int // 0 = sidebar, 1 = main

	// editor
	editMode     bool
	editor       textarea.Model
	editPath     string
	editModified bool
	editConfirm  bool // waiting for discard y/n
	editErr      string

	// file operations
	opMode     opMode
	opInput    textinput.Model
	opPasteDst string // destination dir for pending paste confirmation
	clipboard  *filetree.Node
	clipCut    bool
	opErr      string

	// fuzzy file search
	searchMode     bool
	searchInput    textinput.Model
	searchResults  []search.FileResult
	searchSelected int

	// content grep
	grepMode     bool
	grepInput    textinput.Model
	grepResults  []search.GrepResult // nil = input phase, non-nil = results phase
	grepQuery    string
	grepSelected int
	grepRunning  bool

	// ai chat
	chatMode       bool
	chatMessages   []ai.Message
	chatInput      textinput.Model
	chatScroll     int
	chatStreaming   bool
	chatStreamCh   <-chan ai.AgentEvent
	chatBuf        string // accumulating current streamed response
	anthropicClient *anthropic.Client

	// api key setup
	apiKeyMode  bool
	apiKeyInput textinput.Model
	apiKeyErr   string

	// model selection
	modelSelectMode bool
	modelSelectIdx  int
	activeModel     string // model ID currently in use

	// theme selection
	themeSelectMode bool
	themeSelectIdx  int
}
