package filetree

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	colorFolder    = lipgloss.Color("#534AB7")
	colorFile      = lipgloss.Color("#EF9F27")
	colorSelected  = lipgloss.Color("#c9cad1")
	colorMuted     = lipgloss.Color("#7b7e8a")
	colorConnector = lipgloss.Color("#2e2f34")
)

func (t *Tree) Render(width, height int) string {
	if len(t.Flat) == 0 {
		return lipgloss.NewStyle().
			Foreground(colorMuted).
			Render("  empty directory")
	}

	if height <= 0 {
		height = 1
	}

	lines := make([]string, 0, len(t.Flat))
	for i, node := range t.Flat {
		line := renderNode(node, i, t.Flat, i == t.Selected, width)
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

	// hard cap — never return more than height lines
	visible := lines[start:end]
	if len(visible) > height {
		visible = visible[:height]
	}

	return strings.Join(visible, "\n")
}


func renderNode(node *Node, idx int, flat []*Node, selected bool, width int) string {
	// build connector prefix
	prefix := buildPrefix(node, idx, flat)

	// chevron for dirs
	chevron := " "
	if node.IsDir {
		if node.Expanded {
			chevron = "▼"
		} else {
			chevron = "▶"
		}
	}

	// icon
	icon := "  "
	iconStyle := lipgloss.NewStyle().Foreground(colorFile)
	if node.IsDir {
		iconStyle = lipgloss.NewStyle().Foreground(colorFolder)
		icon = " "
	}

	// name
	nameStyle := lipgloss.NewStyle().Foreground(colorMuted)
	if selected {
		nameStyle = lipgloss.NewStyle().Foreground(colorSelected)
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

// buildPrefix builds the ├─ └─ │ connector string for a node.
func buildPrefix(node *Node, idx int, flat []*Node) string {
	if node.Depth == 0 {
		return ""
	}

	// collect ancestors
	ancestors := make([]bool, node.Depth)
	current := node
	for d := node.Depth - 1; d >= 0; d-- {
		if current.Parent != nil {
			ancestors[d] = hasNextSibling(current, flat)
			current = current.Parent
		}
	}

	prefix := ""
	for d := 0; d < node.Depth-1; d++ {
		if ancestors[d] {
			prefix += "│ "
		} else {
			prefix += "  "
		}
	}

	if hasNextSibling(node, flat) {
		prefix += "├─"
	} else {
		prefix += "└─"
	}

	return prefix
}

// hasNextSibling checks if a node has a sibling after it in the flat list
// at the same depth.
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
