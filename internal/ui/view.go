package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Beaver-family/beaver/internal/ui/styles"
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

	var statusBar string
	if m.opMode != opNone || m.editConfirm {
		statusBar = m.renderOpStatusBar(m.width, bar)
	} else {
		var l1L, l2L string
		if m.searchMode {
			cnt := fmt.Sprintf("  %d results", len(m.searchResults))
			if m.searchInput.Value() == "" {
				cnt = "  type to search"
			}
			l1L = k("/") + " " + m.searchInput.View() + lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(cnt)
			l2L = k("↑/↓") + " navigate" + sep + k("enter") + " open" + sep + k("esc") + " exit"
		} else if m.grepMode {
			if m.grepResults != nil {
				l1L = lipgloss.NewStyle().Foreground(styles.ColorAccent).Render(fmt.Sprintf("  %d matches", len(m.grepResults))) +
					lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(fmt.Sprintf("  for \"%s\"", m.grepQuery))
				l2L = k("↑/↓") + " navigate" + sep + k("enter") + " open" + sep + k("esc") + " new search"
			} else if m.grepRunning {
				l1L = lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(fmt.Sprintf("  searching for \"%s\"...", m.grepQuery))
				l2L = ""
			} else {
				l1L = k("ctrl+f") + " " + m.grepInput.View()
				l2L = k("enter") + " search" + sep + k("esc") + " cancel"
			}
		} else if m.editMode {
			extras := ""
			if m.editErr != "" {
				extras = sep + warnStyle.Render(m.editErr)
			}
			l1L = k("ctrl+s") + " save" + sep + k("esc") + " discard" + extras
			l2L = lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(displayPath(m.editPath, m.width/2))
		} else if m.focus == focusMain {
			l1L = k("↑/↓") + " scroll" + sep + k("ctrl+d/u") + " half page" + sep + k("g/G") + " top/bot" + sep + k("e") + " edit"
			l2L = k("tab") + " → files" + sep + k("esc/←") + " sidebar" + sep + k("pgup/dn") + " page"
		} else {
			l1L = k("↑/↓") + " navigate" + sep + k("enter") + " open" + sep + k("d") + " delete" + sep + k("r") + " rename" + sep + k("/") + " search"
			l2L = k("n") + " new file" + sep + k("N") + " new dir" + sep + k("y/x") + " copy/cut" + sep + k("p") + " paste" + sep + k("ctrl+f") + " grep" + sep + k("q") + " quit"
		}
		l1L = "  " + l1L
		l2L = "  " + l2L

		l1R := lipgloss.NewStyle().Foreground(styles.ColorText).Render(dateStr + "  ")
		l2R := lipgloss.NewStyle().Foreground(styles.ColorAccent).Render(timeStr + "  ")
		g1 := m.width - lipgloss.Width(l1L) - lipgloss.Width(l1R)
		if g1 < 0 {
			g1 = 0
		}
		g2 := m.width - lipgloss.Width(l2L) - lipgloss.Width(l2R)
		if g2 < 0 {
			g2 = 0
		}
		statusBar = bar.Width(m.width).Render(l1L+strings.Repeat(" ", g1)+l1R) + "\n" +
			bar.Width(m.width).Render(l2L+strings.Repeat(" ", g2)+l2R)
	}

	panelHeight := m.height - 2
	treeHeight := panelHeight - 2
	if treeHeight < 1 {
		treeHeight = 1
	}
	sw := styles.SidebarWidth()

	// focus-coloured borders
	sidebarStyle := styles.StyleSidebar.BorderForeground(styles.ColorBorder)
	mainStyle := styles.StyleMain.BorderForeground(styles.ColorBorder)
	if m.editMode || m.grepMode {
		mainStyle = styles.StyleMain.BorderForeground(styles.ColorHighlight)
	} else if m.searchMode || m.focus == focusSidebar {
		sidebarStyle = styles.StyleSidebar.BorderForeground(styles.ColorAccent)
	} else {
		mainStyle = styles.StyleMain.BorderForeground(styles.ColorAccent)
	}

	// ── sidebar ───────────────────────────────────────────────────────────────
	// reserve lines at the bottom for op prompts / clipboard indicator / errors
	bottomLines := m.sidebarBottomLineCount()
	treeCap := treeHeight - bottomLines
	if treeCap < 1 {
		treeCap = 1
	}

	var sidebarContent string
	if m.searchMode {
		sidebarContent = m.renderSearchSidebar(sw-2, panelHeight-2)
	} else {
		sidebarContent = sidebarLabel("files")
		if m.filetree != nil {
			sidebarContent += "\n" + m.filetree.Render(sw-2, treeCap)
		}
		sidebarContent += m.renderSidebarBottom()
	}

	sidebar := sidebarStyle.
		Width(sw).
		Height(panelHeight).
		Render(capLines(sidebarContent, panelHeight))

	// ── main area ─────────────────────────────────────────────────────────────
	mainWidth := m.width - sw - 2
	if mainWidth < 0 {
		mainWidth = 0
	}

	var mainContent string
	if m.grepMode {
		mainContent = capLines(m.renderGrepPanel(mainWidth, panelHeight), panelHeight)
	} else if m.editMode {
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
	)

	return body + "\n" + statusBar
}

// renderOpStatusBar renders a nano-style two-line prompt in the status bar area.
// Line 1: inverted/highlighted question bar.
// Line 2: dark bar with Cancel on the left, action keys on the right.
func (m *model) renderOpStatusBar(width int, bar lipgloss.Style) string {
	promptBar := lipgloss.NewStyle().
		Background(lipgloss.Color("#c0c1c7")).
		Foreground(lipgloss.Color("#0a0b0d")).
		Bold(true)

	kOpt := func(s string) string {
		return lipgloss.NewStyle().
			Background(lipgloss.Color("#0a0b0d")).
			Foreground(styles.ColorAccent).
			Bold(true).
			Render(" " + s + " ")
	}

	cancelOpt := kOpt("esc") + " Cancel"

	optLine := func(question, opts string) string {
		line1 := promptBar.Width(width).Render(question)
		line2 := bar.Width(width).Render("  " + cancelOpt + "    " + opts)
		return line1 + "\n" + line2
	}

	if m.editConfirm {
		question := fmt.Sprintf(`  Discard unsaved changes to "%s" ?`, displayPath(m.editPath, 40))
		return optLine(question, kOpt("Y")+" Yes      "+kOpt("N/esc")+" No")
	}

	switch m.opMode {
	case opConfirmDelete:
		name := ""
		if m.filetree != nil {
			if node := m.filetree.SelectedNode(); node != nil {
				name = node.Name
			}
		}
		question := fmt.Sprintf(`  Delete "%s"  (ANSWERING "No" WILL KEEP THE FILE) ?`, name)
		return optLine(question, kOpt("Y")+" Yes      "+kOpt("N")+" No")

	case opConfirmPaste:
		name := ""
		action := "copy"
		if m.clipboard != nil {
			name = m.clipboard.Name
		}
		if m.clipCut {
			action = "move"
		}
		dst := displayPath(m.opPasteDst, 40)
		question := fmt.Sprintf(`  %s "%s"  →  "%s" ?`, strings.ToUpper(action), name, dst)
		return optLine(question, kOpt("Y")+" Yes      "+kOpt("N")+" No")

	case opInputRename:
		return optLine("  Rename →  "+m.opInput.View(), kOpt("enter")+" Confirm")
	case opInputNewFile:
		return optLine("  New File →  "+m.opInput.View(), kOpt("enter")+" Confirm")
	case opInputNewDir:
		return optLine("  New Folder →  "+m.opInput.View(), kOpt("enter")+" Confirm")
	}
	return bar.Width(width).Render("") + "\n" + bar.Width(width).Render("")
}

// sidebarBottomLineCount returns lines reserved at the bottom of the sidebar
// for clipboard indicator and op errors (prompts moved to status bar).
func (m *model) sidebarBottomLineCount() int {
	n := 0
	if m.opErr != "" {
		n++
	}
	if m.clipboard != nil {
		n++
	}
	return n
}

// renderSidebarBottom renders the clipboard indicator and last op error at the
// bottom of the sidebar. Op prompts live in the status bar (nano-style).
func (m *model) renderSidebarBottom() string {
	if m.sidebarBottomLineCount() == 0 {
		return ""
	}

	muted := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	warn := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff6b6b"))

	var lines []string

	if m.opErr != "" {
		lines = append(lines, warn.Render("  ✗ "+m.opErr))
	}

	if m.clipboard != nil {
		icon := "copy"
		if m.clipCut {
			icon = "cut"
		}
		lines = append(lines, muted.Render("  ["+icon+"] "+m.clipboard.Name))
	}

	return "\n" + strings.Join(lines, "\n")
}

func sidebarLabel(label string) string {
	return "\n" + lipgloss.NewStyle().
		Foreground(styles.ColorMuted).
		Render(strings.ToUpper(label))
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
		tag := accent.Render(" " + fmt.Sprintf("%d", pct) + "% ")
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

// renderSearchSidebar renders the fuzzy-search result list in the sidebar.
func (m *model) renderSearchSidebar(width, height int) string {
	muted := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	dirStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
	fileStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	selStyle := lipgloss.NewStyle().Foreground(styles.ColorAccent).Bold(true)

	header := "\n" + lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("SEARCH")

	if len(m.searchResults) == 0 {
		msg := "  type to search"
		if m.searchInput.Value() != "" {
			msg = "  no results"
		}
		return header + "\n\n" + muted.Render(msg)
	}

	var lines []string
	for i, r := range m.searchResults {
		if len(lines) >= height-2 {
			break
		}
		icon := "  "
		style := fileStyle
		if r.IsDir {
			icon = "  "
			style = dirStyle
		}
		if i == m.searchSelected {
			style = selStyle
		}
		line := icon + style.Render(truncateLine(r.Name, width-4))
		lines = append(lines, line)
	}
	return header + "\n" + strings.Join(lines, "\n")
}

// renderGrepPanel renders grep results (or the waiting state) in the main panel.
func (m *model) renderGrepPanel(width, height int) string {
	muted := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	accent := lipgloss.NewStyle().Foreground(styles.ColorAccent)
	selStyle := lipgloss.NewStyle().Foreground(styles.ColorHighlight).Bold(true)

	if m.grepResults == nil {
		if m.grepRunning {
			return muted.Render(fmt.Sprintf("  searching for \"%s\"...", m.grepQuery))
		}
		return muted.Render("  enter a search term and press enter")
	}

	if len(m.grepResults) == 0 {
		return accent.Render(fmt.Sprintf("  no matches for \"%s\"", m.grepQuery))
	}

	header := accent.Render(fmt.Sprintf("  %d matches  ", len(m.grepResults))) +
		muted.Render(fmt.Sprintf("for \"%s\"", m.grepQuery))

	// keep selected row visible
	visibleRows := height - 2
	if visibleRows < 1 {
		visibleRows = 1
	}
	start := 0
	if m.grepSelected >= visibleRows {
		start = m.grepSelected - visibleRows + 1
	}

	var lines []string
	for i := start; i < len(m.grepResults) && len(lines) < visibleRows; i++ {
		r := m.grepResults[i]
		loc := filepath.Base(r.Path) + ":" + fmt.Sprintf("%d", r.LineNum)
		locW := 22
		locStr := accent.Render("  " + truncateLine(loc, locW))
		var content string
		if i == m.grepSelected {
			content = selStyle.Render(truncateLine(r.Line, width-locW-4))
		} else {
			content = muted.Render(truncateLine(r.Line, width-locW-4))
		}
		lines = append(lines, locStr+"  "+content)
	}
	return header + "\n" + strings.Join(lines, "\n")
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
