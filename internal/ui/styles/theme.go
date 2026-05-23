package styles

import (
	"github.com/Beaver-family/beaver/internal/config"
	"github.com/charmbracelet/lipgloss"
)

// Theme holds every color used across the UI.
type Theme struct {
	Name string

	// Base palette
	Background lipgloss.Color
	Border     lipgloss.Color
	Text       lipgloss.Color
	TextMuted  lipgloss.Color
	Accent     lipgloss.Color
	Highlight  lipgloss.Color

	// File tree
	FileIcon lipgloss.Color
	Danger   lipgloss.Color // cut indicator, warnings, errors

	// Git status
	GitStaged    lipgloss.Color
	GitModified  lipgloss.Color
	GitUntracked lipgloss.Color

	// Editor
	EditorCursorBg lipgloss.Color
	EditorLineNum  lipgloss.Color

	// Prompt bar (nano-style confirm dialogs)
	PromptBg lipgloss.Color
}

// Themes is the registry of all built-in themes.
var Themes = map[string]Theme{
	"default": {
		Name:           "Default",
		Background:     "#0a0b0d",
		Border:         "#2e2f34",
		Text:           "#c9cad1",
		TextMuted:      "#5a5c68",
		Accent:         "#5DCAA5",
		Highlight:      "#534AB7",
		FileIcon:       "#EF9F27",
		Danger:         "#ff6b6b",
		GitStaged:      "#98C379",
		GitModified:    "#E5C07B",
		GitUntracked:   "#E06C75",
		EditorCursorBg: "#1e1f24",
		EditorLineNum:  "#3e3f4a",
		PromptBg:       "#c0c1c7",
	},
	"dracula": {
		Name:           "Dracula",
		Background:     "#282a36",
		Border:         "#44475a",
		Text:           "#f8f8f2",
		TextMuted:      "#6272a4",
		Accent:         "#50fa7b",
		Highlight:      "#bd93f9",
		FileIcon:       "#ffb86c",
		Danger:         "#ff5555",
		GitStaged:      "#50fa7b",
		GitModified:    "#f1fa8c",
		GitUntracked:   "#ff5555",
		EditorCursorBg: "#44475a",
		EditorLineNum:  "#44475a",
		PromptBg:       "#e2e2e6",
	},
	"nord": {
		Name:           "Nord",
		Background:     "#2e3440",
		Border:         "#3b4252",
		Text:           "#d8dee9",
		TextMuted:      "#4c566a",
		Accent:         "#88c0d0",
		Highlight:      "#5e81ac",
		FileIcon:       "#ebcb8b",
		Danger:         "#bf616a",
		GitStaged:      "#a3be8c",
		GitModified:    "#ebcb8b",
		GitUntracked:   "#bf616a",
		EditorCursorBg: "#3b4252",
		EditorLineNum:  "#4c566a",
		PromptBg:       "#d8dee9",
	},
	"catppuccin": {
		Name:           "Catppuccin Mocha",
		Background:     "#1e1e2e",
		Border:         "#313244",
		Text:           "#cdd6f4",
		TextMuted:      "#6c7086",
		Accent:         "#89dceb",
		Highlight:      "#cba6f7",
		FileIcon:       "#fab387",
		Danger:         "#f38ba8",
		GitStaged:      "#a6e3a1",
		GitModified:    "#f9e2af",
		GitUntracked:   "#f38ba8",
		EditorCursorBg: "#313244",
		EditorLineNum:  "#45475a",
		PromptBg:       "#cdd6f4",
	},
	"tokyonight": {
		Name:           "Tokyo Night",
		Background:     "#1a1b26",
		Border:         "#292e42",
		Text:           "#c0caf5",
		TextMuted:      "#565f89",
		Accent:         "#7dcfff",
		Highlight:      "#bb9af7",
		FileIcon:       "#e0af68",
		Danger:         "#f7768e",
		GitStaged:      "#9ece6a",
		GitModified:    "#e0af68",
		GitUntracked:   "#f7768e",
		EditorCursorBg: "#292e42",
		EditorLineNum:  "#3b3d57",
		PromptBg:       "#c0caf5",
	},
}

// ThemeNames is the canonical display order for the theme picker.
var ThemeNames = []string{"default", "dracula", "nord", "catppuccin", "tokyonight"}

// Current is the active theme. Defaults to "default"; updated by SetTheme.
var Current = Themes["default"]

// Exported color vars — kept for backwards compatibility with existing call sites.
// SetTheme keeps these in sync with Current.
var (
	ColorBorder    = Current.Border
	ColorMuted     = Current.TextMuted
	ColorText      = Current.Text
	ColorAccent    = Current.Accent
	ColorHighlight = Current.Highlight
)

// SetTheme activates a named theme and syncs all exported vars.
// Returns false (and keeps current theme) if name is not recognised.
func SetTheme(name string) bool {
	t, ok := Themes[name]
	if !ok {
		return false
	}
	Current = t
	ColorBorder    = t.Border
	ColorMuted     = t.TextMuted
	ColorText      = t.Text
	ColorAccent    = t.Accent
	ColorHighlight = t.Highlight
	return true
}

// StyleSidebar returns a fresh sidebar style using the current theme.
func StyleSidebar() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Current.Border).
		PaddingLeft(1).
		MaxHeight(0)
}

// StyleMain returns a fresh main-panel style using the current theme.
func StyleMain() lipgloss.Style {
	return lipgloss.NewStyle().
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(Current.Border).
		PaddingLeft(1)
}

func SidebarWidth() int {
	return config.GetUIConfig().SidebarWidth
}
