package styles

import (
	"github.com/Beaver-family/beaver/internal/config"
	"github.com/charmbracelet/lipgloss"
)

var (
	ColorBorder    = lipgloss.Color("#2e2f34")
	ColorMuted     = lipgloss.Color("#5a5c68")
	ColorText      = lipgloss.Color("#c9cad1")
	ColorAccent    = lipgloss.Color("#5DCAA5")
	ColorHighlight = lipgloss.Color("#534AB7")

	StyleTopBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#0a0b0d")).
			Foreground(ColorText).
			PaddingLeft(2)

	StyleStatusBar = lipgloss.NewStyle().
			Background(lipgloss.Color("#0a0b0d")).
			Foreground(ColorMuted).
			PaddingLeft(2)

	StyleSidebar = lipgloss.NewStyle().
			BorderRight(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder).
			PaddingLeft(1).
			MaxHeight(0)

	StyleMain = lipgloss.NewStyle().
			BorderRight(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder).
			PaddingLeft(1)

)

func SidebarWidth() int {
	return config.GetUIConfig().SidebarWidth
}
