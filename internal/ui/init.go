package ui

import (
	"os"
	"time"

	"github.com/Beaver-family/beaver/internal/ui/filetree"
	"github.com/Beaver-family/beaver/internal/ui/git"
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
	cmds := []tea.Cmd{tickCmd()}
	if m.filetree != nil {
		cmds = append(cmds, git.LoadCmd(m.filetree.Root.Path))
	}
	return tea.Batch(cmds...)
}
