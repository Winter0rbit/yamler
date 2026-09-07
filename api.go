// Document-level query and mutation helpers: Has, Keys, Copy, Delete,
// DeleteAll.

package yamler

import (
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// nodeAt resolves a path to its node, for mapping-root and array-root
// documents alike. An empty path returns the root node.
func (d *Document) nodeAt(path string) (*yaml.Node, error) {
	if d.root == nil || len(d.root.Content) == 0 {
		return nil, wrapErr(ErrRoot, "empty document root")
	}
	node := d.root.Content[0]
	if path == "" {
		return node, nil
	}
	for _, part := range strings.Split(path, ".") {
		var err error
		node, err = navigateToNode(node, part, path)
		if err != nil {
			return nil, err
		}
	}
	return node, nil
}

// parentAndLast splits a path into the path of the parent node and the last
// segment ("a.b[2]" -> "a.b", "[2]"; "a.b" -> "a", "b"; "[0]" -> "", "[0]").
func parentAndLast(path string) (parent, last string) {
	if strings.HasSuffix(path, "]") {
		open := strings.LastIndex(path, "[")
		if open < 0 {
			return "", path
		}
		return path[:open], path[open:]
	}
	if dot := strings.LastIndex(path, "."); dot >= 0 {
		return path[:dot], path[dot+1:]
	}
	return "", path
}

// Has reports whether a value exists at the path.
func (d *Document) Has(path string) bool {
	_, err := d.nodeAt(path)
	return err == nil
}

// Keys returns the keys of the mapping at the path in document order. An
// empty path lists the top-level keys.
func (d *Document) Keys(path string) ([]string, error) {
	node, err := d.nodeAt(path)
	if err != nil {
		return nil, err
	}
	if node.Kind != yaml.MappingNode {
		return nil, wrapErr(ErrType, "path %s: expected map, got %s", path, kindName(node))
	}
	keys := make([]string, 0, len(node.Content)/2)
	for i := 0; i+1 < len(node.Content); i += 2 {
		keys = append(keys, node.Content[i].Value)
	}
	return keys, nil
}

// Copy returns an independent deep copy of the document, including its
// formatting information. Changes to the copy do not affect the original.
func (d *Document) Copy() *Document {
	c := *d
	if d.root != nil {
		c.root, _ = cloneNode(d.root)
		relinkAliases(c.root)
	}
	if d.formattingCache != nil {
		c.formattingCache = d.formattingCache.clone()
	}
	return &c
}

// relinkAliases points alias nodes of a cloned tree at the anchors of the
// same tree instead of the tree they were cloned from.
func relinkAliases(root *yaml.Node) {
	anchors := make(map[string]*yaml.Node)
	var collect func(n *yaml.Node)
	collect = func(n *yaml.Node) {
		if n.Anchor != "" {
			anchors[n.Anchor] = n
		}
		for _, c := range n.Content {
			collect(c)
		}
	}
	collect(root)
	var fix func(n *yaml.Node)
	fix = func(n *yaml.Node) {
		if n.Kind == yaml.AliasNode {
			if target, ok := anchors[n.Value]; ok {
				n.Alias = target
			}
		}
		for _, c := range n.Content {
			fix(c)
		}
	}
	fix(root)
}

// Delete removes the key or array element at the path. Deleting a missing
// path returns ErrNotFound.
func (d *Document) Delete(path string) error {
	if path == "" {
		return wrapErr(ErrPath, "empty path")
	}
	parentPath, last := parentAndLast(path)
	parent, err := d.nodeAt(parentPath)
	if err != nil {
		return err
	}
	if strings.HasPrefix(last, "[") {
		if parent.Kind != yaml.SequenceNode {
			return wrapErr(ErrType, "path %s: expected array, got %s", path, kindName(parent))
		}
		idx, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(last, "["), "]"))
		if err != nil {
			return wrapErr(ErrPath, "path %s: invalid array index: %s", path, last)
		}
		if idx < 0 || idx >= len(parent.Content) {
			return wrapErr(ErrIndex, "path %s: array index out of bounds", path)
		}
		parent.Content = append(parent.Content[:idx], parent.Content[idx+1:]...)
	} else {
		if parent.Kind != yaml.MappingNode {
			return wrapErr(ErrType, "path %s: expected map, got %s", path, kindName(parent))
		}
		found := false
		for i := 0; i+1 < len(parent.Content); i += 2 {
			if parent.Content[i].Value == last {
				parent.Content = append(parent.Content[:i], parent.Content[i+2:]...)
				found = true
				break
			}
		}
		if !found {
			return wrapErr(ErrNotFound, "path %s: key %s not found", path, last)
		}
	}
	return d.reserialize()
}

// DeleteAll removes every path matching the wildcard pattern (see GetAll)
// and returns the number of removed entries.
func (d *Document) DeleteAll(pattern string) (int, error) {
	root, err := d.nodeAt("")
	if err != nil {
		return 0, err
	}
	matches := make(map[string]interface{})
	if err := findMatchingPaths(root, pattern, "", matches); err != nil {
		return 0, err
	}
	if len(matches) == 0 {
		return 0, nil
	}
	// Delete deeper paths and higher indices first so that removing one
	// entry does not shift the paths of the others.
	paths := make([]string, 0, len(matches))
	for p := range matches {
		paths = append(paths, p)
	}
	sort.Sort(sort.Reverse(byDocumentOrder(paths)))
	for _, p := range paths {
		if err := d.deleteNode(p); err != nil {
			return 0, err
		}
	}
	return len(paths), d.reserialize()
}

// deleteNode is Delete without re-serialization, for bulk operations.
func (d *Document) deleteNode(path string) error {
	parentPath, last := parentAndLast(path)
	parent, err := d.nodeAt(parentPath)
	if err != nil {
		return err
	}
	if strings.HasPrefix(last, "[") {
		idx, err := strconv.Atoi(strings.Trim(last, "[]"))
		if err != nil || parent.Kind != yaml.SequenceNode || idx < 0 || idx >= len(parent.Content) {
			return wrapErr(ErrNotFound, "path %s: not found", path)
		}
		parent.Content = append(parent.Content[:idx], parent.Content[idx+1:]...)
		return nil
	}
	if parent.Kind != yaml.MappingNode {
		return wrapErr(ErrNotFound, "path %s: not found", path)
	}
	for i := 0; i+1 < len(parent.Content); i += 2 {
		if parent.Content[i].Value == last {
			parent.Content = append(parent.Content[:i], parent.Content[i+2:]...)
			return nil
		}
	}
	return wrapErr(ErrNotFound, "path %s: key %s not found", path, last)
}

// byDocumentOrder sorts paths so that a path sorts before its own children
// and array indices compare numerically.
type byDocumentOrder []string

func (p byDocumentOrder) Len() int      { return len(p) }
func (p byDocumentOrder) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p byDocumentOrder) Less(i, j int) bool {
	a, b := pathSegments(p[i]), pathSegments(p[j])
	for k := 0; k < len(a) && k < len(b); k++ {
		if a[k] == b[k] {
			continue
		}
		ai, aErr := strconv.Atoi(strings.Trim(a[k], "[]"))
		bi, bErr := strconv.Atoi(strings.Trim(b[k], "[]"))
		if aErr == nil && bErr == nil && strings.HasPrefix(a[k], "[") && strings.HasPrefix(b[k], "[") {
			return ai < bi
		}
		return a[k] < b[k]
	}
	return len(a) < len(b)
}

// pathSegments splits "a.b[2].c" into ["a", "b", "[2]", "c"].
func pathSegments(path string) []string {
	return splitPath(path)
}

// reserialize re-encodes the document after a structural change so that
// d.raw reflects the current tree.
func (d *Document) reserialize() error {
	content, err := d.ToBytes()
	if err != nil {
		return err
	}
	d.raw = string(content)
	return nil
}

func kindName(n *yaml.Node) string {
	switch n.Kind {
	case yaml.MappingNode:
		return "map"
	case yaml.SequenceNode:
		return "array"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	case yaml.DocumentNode:
		return "document"
	}
	return "unknown"
}
