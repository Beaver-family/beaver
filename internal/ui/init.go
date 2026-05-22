package ui

import (
	"os"
	"time"

	"github.com/Beaver-family/beaver/internal/ui/filetree"
	tea "github.com/charmbracelet/bubbletea"
)

func New() *model {
	cwd, err := os.Getwd()
	if err != nil {
		cwd, _ = os.UserHomeDir()
	}
	tree, _ := filetree.New(cwd)

	return &model{
		now:      time.Now(),
		filetree: tree,
	}
}

func (m *model) Init() tea.Cmd {
	return tickCmd()
}
