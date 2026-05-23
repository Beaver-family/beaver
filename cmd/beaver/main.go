package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/Beaver-family/beaver/internal/config"
	"github.com/Beaver-family/beaver/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

// version is overridden by GoReleaser via -X main.version={{.Version}}.
// For go install builds, we fall back to the module version from build info.
var version = "dev"

func resolveVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Println("beaver", resolveVersion())
		return
	}

	config.LoadConfig()

	p := tea.NewProgram(ui.New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
