package filetree

import (
	"strings"

	"github.com/Beaver-family/beaver/internal/ui/git"
	"github.com/Beaver-family/beaver/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

func (t *Tree) Render(width, height int) string {
	if len(t.Flat) == 0 {
		return lipgloss.NewStyle().
			Foreground(styles.Current.TextMuted).
			Render("  empty directory")
	}

	if height <= 0 {
		height = 1
	}

	lines := make([]string, 0, len(t.Flat))
	for i, node := range t.Flat {
		line := renderNode(node, t, i, t.Flat, i == t.Selected, width)
		lines = append(lines, line)
	}

	// keep selected always in view
	start := 0
	if t.Selected >= height {
		start = t.Selected - height + 1
	}

	end := start + height
	if end > len(lines) {
		end = len(lines)
	}

	visible := lines[start:end]
	if len(visible) > height {
		visible = visible[:height]
	}

	return strings.Join(visible, "\n")
}

func renderNode(node *Node, t *Tree, idx int, flat []*Node, selected bool, width int) string {
	_ = idx

	// colors read fresh each frame from the active theme
	colorFolder    := styles.Current.Highlight
	colorFile      := styles.Current.FileIcon
	colorSelected  := styles.Current.Text
	colorMuted     := styles.Current.TextMuted
	colorConnector := styles.Current.Border
	colorCut       := styles.Current.Danger
	colorCopy      := styles.Current.Accent

	prefix := buildPrefix(node, idx, flat)

	chevron := " "
	if node.IsDir {
		if node.Expanded {
			chevron = "▼"
		} else {
			chevron = "▶"
		}
	}

	icon := "  "
	iconStyle := lipgloss.NewStyle().Foreground(colorFile)
	if node.IsDir {
		iconStyle = lipgloss.NewStyle().Foreground(colorFolder)
		icon = " "
	}

	inClipboard := t.ClipboardPath != "" && node.Path == t.ClipboardPath
	nameStyle := lipgloss.NewStyle().Foreground(colorMuted)
	if selected {
		nameStyle = lipgloss.NewStyle().Foreground(colorSelected)
	} else if inClipboard {
		if t.ClipboardCut {
			nameStyle = lipgloss.NewStyle().Foreground(colorCut)
		} else {
			nameStyle = lipgloss.NewStyle().Foreground(colorCopy)
		}
	} else if gs, ok := t.GitStatus[node.Path]; ok && gs != git.StateClean {
		c := gitColor(gs)
		nameStyle = lipgloss.NewStyle().Foreground(c)
		iconStyle = lipgloss.NewStyle().Foreground(c)
	}

	connStyle := lipgloss.NewStyle().Foreground(colorConnector)
	chevStyle := lipgloss.NewStyle().Foreground(colorMuted)

	available := width - lipgloss.Width(prefix) - 1 - lipgloss.Width(icon)
	if available < 4 {
		available = 4
	}

	return connStyle.Render(prefix) +
		chevStyle.Render(chevron) +
		iconStyle.Render(icon) +
		nameStyle.Render(truncate(node.Name, available))
}

func buildPrefix(node *Node, idx int, flat []*Node) string {
	if node.Depth == 0 {
		return ""
	}

	ancestors := make([]bool, node.Depth)
	current := node
	for d := node.Depth - 1; d >= 0; d-- {
		if current.Parent != nil {
			ancestors[d] = hasNextSibling(current, flat)
			current = current.Parent
		}
	}

	var sb strings.Builder
	for d := 0; d < node.Depth-1; d++ {
		if ancestors[d] {
			sb.WriteString("│ ")
		} else {
			sb.WriteString("  ")
		}
	}
	if hasNextSibling(node, flat) {
		sb.WriteString("├─")
	} else {
		sb.WriteString("└─")
	}
	return sb.String()
}

func hasNextSibling(node *Node, flat []*Node) bool {
	found := false
	for _, n := range flat {
		if found && n.Depth == node.Depth && n.Parent == node.Parent {
			return true
		}
		if n == node {
			found = true
		}
	}
	return false
}

func gitColor(gs git.FileState) lipgloss.Color {
	switch gs {
	case git.StateStaged:
		return styles.Current.GitStaged
	case git.StateModified:
		return styles.Current.GitModified
	default:
		return styles.Current.GitUntracked
	}
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
