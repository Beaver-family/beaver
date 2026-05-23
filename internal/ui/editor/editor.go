package editor

import (
	"github.com/Beaver-family/beaver/internal/ui/styles"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

// New creates a textarea configured for code editing using the active theme.
func New(content string, width, height int) textarea.Model {
	ta := textarea.New()
	ta.SetValue(content)
	ta.SetWidth(width)
	ta.SetHeight(height)
	ta.ShowLineNumbers = true
	ta.CharLimit = 0

	focused := ta.FocusedStyle
	focused.Base = lipgloss.NewStyle()
	focused.CursorLine = lipgloss.NewStyle().Background(styles.Current.EditorCursorBg)
	focused.CursorLineNumber = lipgloss.NewStyle().Foreground(styles.Current.Accent)
	focused.LineNumber = lipgloss.NewStyle().Foreground(styles.Current.EditorLineNum)
	focused.Text = lipgloss.NewStyle().Foreground(styles.Current.Text)
	ta.FocusedStyle = focused

	blurred := ta.BlurredStyle
	blurred.Base = lipgloss.NewStyle()
	blurred.Text = lipgloss.NewStyle().Foreground(styles.Current.TextMuted)
	ta.BlurredStyle = blurred

	ta.Focus()
	return ta
}

// Resize updates the textarea dimensions to match a new panel size.
func Resize(ta *textarea.Model, width, height int) {
	ta.SetWidth(width)
	ta.SetHeight(height)
}
