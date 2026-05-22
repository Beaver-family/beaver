package editor

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// New creates a textarea configured for code editing.
func New(content string, width, height int) textarea.Model {
	ta := textarea.New()
	ta.SetValue(content)
	ta.SetWidth(width)
	ta.SetHeight(height)
	ta.ShowLineNumbers = true
	ta.CharLimit = 0

	// style the focused state to match the app theme
	focused := ta.FocusedStyle
	focused.Base = lipgloss.NewStyle()
	focused.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("#1e1f24"))
	focused.CursorLineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color("#5DCAA5"))
	focused.LineNumber = lipgloss.NewStyle().Foreground(lipgloss.Color("#3e3f4a"))
	focused.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("#c9cad1"))
	ta.FocusedStyle = focused

	blurred := ta.BlurredStyle
	blurred.Base = lipgloss.NewStyle()
	blurred.Text = lipgloss.NewStyle().Foreground(lipgloss.Color("#5a5c68"))
	ta.BlurredStyle = blurred

	ta.Focus()
	return ta
}

// Resize updates the textarea dimensions to match a new panel size.
func Resize(ta *textarea.Model, width, height int) {
	ta.SetWidth(width)
	ta.SetHeight(height)
}
