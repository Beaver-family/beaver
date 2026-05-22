package filetree

import (
	"os"
	"path/filepath"
	"sort"
)

type Tree struct {
	Root     *Node
	Flat     []*Node  // flattened visible nodes — what gets rendered
	Selected int      // index into Flat
}

func New(rootPath string) (*Tree, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}

	root := &Node{
		Name:     info.Name(),
		Path:     rootPath,
		IsDir:    true,
		Depth:    0,
		Expanded: true,
	}

	if err := loadChildren(root); err != nil {
		return nil, err
	}

	t := &Tree{Root: root}
	t.flatten()
	return t, nil
}

// loadChildren reads one level of directory entries into node.Children.
func loadChildren(node *Node) error {
	entries, err := os.ReadDir(node.Path)
	if err != nil {
		return err
	}

	// sort: folders first, then files, both alphabetical
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir() != entries[j].IsDir() {
			return entries[i].IsDir()
		}
		return entries[i].Name() < entries[j].Name()
	})

	for _, e := range entries {
		// skip hidden files
		if e.Name()[0] == '.' {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		child := NewNode(info, filepath.Join(node.Path, e.Name()), node.Depth+1, node)
		node.Children = append(node.Children, child)
	}
	return nil
}

// flatten walks the tree and builds Flat — only expanded nodes are included.
func (t *Tree) flatten() {
	t.Flat = nil
	var walk func(node *Node)
	walk = func(node *Node) {
		t.Flat = append(t.Flat, node)
		if node.IsDir && node.Expanded {
			for _, child := range node.Children {
				walk(child)
			}
		}
	}
	// start from root's children so root itself isn't in the list
	for _, child := range t.Root.Children {
		walk(child)
	}
}

// Toggle opens or closes a directory node.
func (t *Tree) Toggle(idx int) error {
	if idx < 0 || idx >= len(t.Flat) {
		return nil
	}
	node := t.Flat[idx]
	if !node.IsDir {
		return nil
	}

	node.Expanded = !node.Expanded

	// load children on first open
	if node.Expanded && len(node.Children) == 0 {
		if err := loadChildren(node); err != nil {
			return err
		}
	}

	t.flatten()
	return nil
}

// Up moves selection up.
func (t *Tree) Up() {
	if t.Selected > 0 {
		t.Selected--
	}
}

// Down moves selection down.
func (t *Tree) Down() {
	if t.Selected < len(t.Flat)-1 {
		t.Selected++
	}
}

// SelectedNode returns the currently selected node.
func (t *Tree) SelectedNode() *Node {
	if len(t.Flat) == 0 {
		return nil
	}
	return t.Flat[t.Selected]
}
