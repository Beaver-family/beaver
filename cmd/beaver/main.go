package main

import (
	"fmt"
	"os"

	"github.com/Beaver-family/beaver/internal/config"
	"github.com/Beaver-family/beaver/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("beaver", version)
		return
	}

	config.LoadConfig()

	p := tea.NewProgram(ui.New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
