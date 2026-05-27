package fileselect

import (
	"sort"
	"strings"
)

// TreeNode represents a single node (file or directory) in a tree.
type TreeNode struct {
	Name     string
	IsDir    bool
	Children map[string]*TreeNode
}

// BuildTree organises a list of slash-separated paths into a directory tree.
func BuildTree(paths []string) *TreeNode {
	root := &TreeNode{Name: "", IsDir: true, Children: map[string]*TreeNode{}}
	for _, p := range paths {
		segments := strings.Split(p, "/")
		cur := root
		for i, seg := range segments {
			isLast := i == len(segments)-1
			child, ok := cur.Children[seg]
			if !ok {
				child = &TreeNode{Name: seg, IsDir: !isLast, Children: map[string]*TreeNode{}}
				cur.Children[seg] = child
			}
			if !isLast {
				child.IsDir = true
			}
			cur = child
		}
	}
	return root
}

// RenderTree renders a tree node as a flat indented string. Directories are
// listed before files, both in alphabetical order, with directories suffixed
// by `/`.
func RenderTree(root *TreeNode) string {
	var b strings.Builder
	renderTreeInto(&b, root, "")
	return strings.TrimRight(b.String(), "\n")
}

func renderTreeInto(b *strings.Builder, n *TreeNode, indent string) {
	keys := make([]string, 0, len(n.Children))
	for k := range n.Children {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, c := n.Children[keys[i]], n.Children[keys[j]]
		if a.IsDir != c.IsDir {
			return a.IsDir
		}
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		child := n.Children[k]
		b.WriteString(indent)
		b.WriteString(child.Name)
		if child.IsDir {
			b.WriteString("/")
		}
		b.WriteString("\n")
		if child.IsDir {
			renderTreeInto(b, child, indent+"  ")
		}
	}
}

// PathsOf returns the slash-separated paths of the selected files.
func PathsOf(files []Result) []string {
	out := make([]string, 0, len(files))
	for _, f := range files {
		out = append(out, f.RelPath)
	}
	return out
}
