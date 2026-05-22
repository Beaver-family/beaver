package ui

import (
	"time"

	"github.com/Beaver-family/tui/internal/ui/filetree"
	"github.com/charmbracelet/bubbles/textarea"
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
}
