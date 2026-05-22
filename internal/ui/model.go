package ui

import (
	"time"

	"github.com/Beaver-family/tui/internal/ui/filetree"
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
	opMode    opMode
	opInput   textinput.Model
	opPasteDst string        // destination dir for pending paste confirmation
	clipboard *filetree.Node
	clipCut   bool
	opErr     string
}
