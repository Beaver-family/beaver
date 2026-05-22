package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/Beaver-family/tui/internal/ui/styles"
	xansi "github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) View() string {
	if m.width == 0 {
		return "loading..."
	}

	timeStr := m.now.Format("15:04:05")
	dateStr := m.now.Format("Mon, 02 Jan 2006")

	bar := lipgloss.NewStyle().
		Background(lipgloss.Color("#0a0b0d")).
		Foreground(styles.ColorMuted)

	sep := lipgloss.NewStyle().Foreground(styles.ColorBorder).Render("  │  ")
	k := func(s string) string {
		return lipgloss.NewStyle().Foreground(styles.ColorAccent).Render(s)
	}
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b6b"))

	var l1L, l2L string
	if m.editMode {
		mode := lipgloss.NewStyle().Foreground(styles.ColorHighlight).Bold(true).Render("EDIT")
		indent := strings.Repeat(" ", lipgloss.Width(mode))
		extras := ""
		if m.editConfirm {
			extras = sep + warnStyle.Render("discard changes?") + "  " + k("y") + " yes  " + k("n/esc") + " no"
		} else if m.editErr != "" {
			extras = sep + warnStyle.Render(m.editErr)
		}
		l1L = mode + sep + k("ctrl+s") + " save" + sep + k("esc") + " discard" + extras
		l2L = indent + sep + lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(displayPath(m.editPath, m.width/2))
	} else if m.focus == focusMain {
		mode := lipgloss.NewStyle().Foreground(styles.ColorAccent).Bold(true).Render("PREVIEW")
		indent := strings.Repeat(" ", lipgloss.Width(mode))
		l1L = mode + sep + k("↑/↓") + " scroll" + sep + k("ctrl+d/u") + " half page" + sep + k("g/G") + " top/bot" + sep + k("e") + " edit"
		l2L = indent + sep + k("tab") + " → files" + sep + k("esc/←") + " sidebar" + sep + k("pgup/dn") + " page"
	} else {
		mode := lipgloss.NewStyle().Foreground(styles.ColorText).Bold(true).Render("SIDEBAR")
		indent := strings.Repeat(" ", lipgloss.Width(mode))
		l1L = mode + sep + k("↑/↓") + " navigate" + sep + k("enter") + " open" + sep + k("l/→") + " preview"
		l2L = indent + sep + k("tab") + " switch focus" + sep + k("q") + " quit"
	}
	l1L = "  " + l1L
	l2L = "  " + l2L

	l1R := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(dateStr + "  ")
	l2R := lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(timeStr + "  ")
	g1 := m.width - lipgloss.Width(l1L) - lipgloss.Width(l1R)
	if g1 < 0 {
		g1 = 0
	}
	g2 := m.width - lipgloss.Width(l2L) - lipgloss.Width(l2R)
	if g2 < 0 {
		g2 = 0
	}
	statusBar := bar.Width(m.width).Render(l1L+strings.Repeat(" ", g1)+l1R) + "\n" +
		bar.Width(m.width).Render(l2L+strings.Repeat(" ", g2)+l2R)

	panelHeight := m.height - 2
	treeHeight := panelHeight - 2
	if treeHeight < 1 {
		treeHeight = 1
	}
	sw := styles.SidebarWidth()

	// focus-coloured borders
	sidebarStyle := styles.StyleSidebar.BorderForeground(styles.ColorBorder)
	mainStyle := styles.StyleMain.BorderForeground(styles.ColorBorder)
	if m.editMode {
		mainStyle = styles.StyleMain.BorderForeground(styles.ColorHighlight)
	} else if m.focus == focusSidebar {
		sidebarStyle = styles.StyleSidebar.BorderForeground(styles.ColorAccent)
	} else {
		mainStyle = styles.StyleMain.BorderForeground(styles.ColorAccent)
	}

	// ── sidebar ───────────────────────────────────────────────────────────────
	sidebarContent := sidebarLabel("files")
	if m.filetree != nil {
		sidebarContent += "\n" + m.filetree.Render(sw-2, treeHeight)
	}

	sidebar := sidebarStyle.
		Width(sw).
		Height(panelHeight).
		Render(capLines(sidebarContent, panelHeight))

	// ── right panel ───────────────────────────────────────────────────────────
	rightContent := strings.Join([]string{
		sidebarLabel("system"),
		statRow("cpu", "12%"),
		statRow("memory", "34%"),
		statRow("disk", "55%"),
		"",
		sidebarLabel("network"),
		statRow("latency", "18ms"),
		statRow("uptime", "4h32m"),
	}, "\n")

	rightPanel := styles.StyleRightPanel.
		Width(styles.RightPanelWidth).
		Height(panelHeight).
		Render(capLines(rightContent, panelHeight))

	// ── main area ─────────────────────────────────────────────────────────────
	mainWidth := m.width - sw - styles.RightPanelWidth - 4
	if mainWidth < 0 {
		mainWidth = 0
	}

	var mainContent string
	if m.editMode {
		mainContent = m.renderEditor()
	} else {
		mainContent = capLines(m.renderPreview(mainWidth, panelHeight), panelHeight)
	}

	mainArea := mainStyle.
		Width(mainWidth).
		Height(panelHeight).
		Render(mainContent)

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		sidebar,
		mainArea,
		rightPanel,
	)

	return body + "\n" + statusBar
}

func sidebarLabel(label string) string {
	return "\n" + lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Render(strings.ToUpper(label))
}


func statRow(label, value string) string {
	l := lipgloss.NewStyle().Foreground(styles.ColorMuted).Width(10).Render(label)
	v := lipgloss.NewStyle().Foreground(styles.ColorText).Render(value)
	return l + v
}

func (m *model) renderEditor() string {
	accent := lipgloss.NewStyle().Foreground(styles.ColorHighlight)
	header := accent.Render(" " + displayPath(m.editPath, 60) + "  [editing]")
	return header + "\n" + m.editor.View()
}

func (m *model) renderPreview(width, height int) string {
	muted := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	accent := lipgloss.NewStyle().Foreground(styles.ColorAccent)

	// nothing selected yet
	if m.previewPath == "" {
		return muted.Render("  select a file")
	}

	// file that can't be shown
	if m.previewErr != "" {
		name := m.previewPath
		return accent.Render("  "+name) + "\n\n" + muted.Render("  "+m.previewErr)
	}

	// header: shorten path with ~ then left-truncate so filename always shows
	header := accent.Render(" " + displayPath(m.previewPath, width-2))

	// clamp scroll
	contentLines := height - 1 // 1 line for header
	if contentLines < 1 {
		contentLines = 1
	}
	maxScroll := len(m.previewLines) - contentLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.previewScroll > maxScroll {
		m.previewScroll = maxScroll
	}

	end := m.previewScroll + contentLines
	if end > len(m.previewLines) {
		end = len(m.previewLines)
	}
	visible := m.previewLines[m.previewScroll:end]

	// truncate lines that are wider than the panel
	out := make([]string, len(visible))
	for i, l := range visible {
		out[i] = truncateLine(l, width-1)
	}

	// scroll indicator: replace last line's tail with pct if file is scrollable
	total := len(m.previewLines)
	if total > contentLines && len(out) > 0 {
		pct := 0
		if maxScroll > 0 {
			pct = m.previewScroll * 100 / maxScroll
		}
		tag := muted.Render(" " + fmt.Sprintf("%d", pct) + "% ")
		last := out[len(out)-1]
		tagW := lipgloss.Width(tag)
		lastW := lipgloss.Width(last)
		// content area = width-1 (PaddingLeft eats 1 col), never exceed it
		avail := width - 1
		if lastW+tagW <= avail {
			last = last + strings.Repeat(" ", avail-lastW-tagW) + tag
		} else {
			last = truncateLine(last, avail-tagW) + tag
		}
		out[len(out)-1] = last
	}

	body := strings.Join(out, "\n")
	return header + "\n" + body
}

// truncateLine cuts a string to max visible columns, ANSI-safe.
// Uses charmbracelet/x/ansi.Truncate which preserves and closes open escape sequences.
func truncateLine(s string, maxW int) string {
	if maxW <= 0 {
		return ""
	}
	return xansi.Truncate(s, maxW, "")
}


// displayPath shortens path to fit maxW columns.
// Replaces home dir with ~, then left-truncates with ... so the filename is always visible.
func displayPath(path string, maxW int) string {
	if home, err := os.UserHomeDir(); err == nil {
		if strings.HasPrefix(path, home) {
			path = "~" + path[len(home):]
		}
	}
	if len(path) <= maxW {
		return path
	}
	// left-truncate: keep the tail (filename visible)
	return "..." + path[len(path)-(maxW-3):]
}

func capLines(s string, maxLines int) string {
	if maxLines <= 0 {
		return ""
	}
	lines := strings.SplitN(s, "\n", maxLines+1)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}
