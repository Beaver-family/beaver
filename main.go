package main

import (
	"fmt"
	"os"

	"github.com/Beaver-family/tui/internal/config"
	"github.com/Beaver-family/tui/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	config.LoadConfig()
	
	p := tea.NewProgram(ui.New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
