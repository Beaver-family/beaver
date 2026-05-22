package filetree

import "os"

type Node struct {
	Name     string
	Path     string
	IsDir    bool
	Depth    int
	Expanded bool
	Children []*Node
	Parent   *Node
}

func NewNode(info os.FileInfo, path string, depth int, parent *Node) *Node {
	return &Node{
		Name:   info.Name(),
		Path:   path,
		IsDir:  info.IsDir(),
		Depth:  depth,
		Parent: parent,
	}
}