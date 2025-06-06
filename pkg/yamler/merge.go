package yamler

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Merge merges another Document into this one, preserving the formatting of this document
// and adding/updating values from the other document
func (d *Document) Merge(other *Document) error {
	if other == nil {
		return fmt.Errorf("other document is nil")
	}

	otherRoot, err := other.mappingRoot()
	if err != nil {
		return fmt.Errorf("other document has invalid root: %w", err)
	}

	thisRoot, err := d.mappingRoot()
	if err != nil {
		return fmt.Errorf("this document has invalid root: %w", err)
	}

	err = mergeNodes(thisRoot, otherRoot)
	if err != nil {
		return err
	}

	// Update the raw content
	content, err := d.ToBytes()
	if err != nil {
		return err
	}
	d.raw = string(content)

	return nil
}

// MergeAt merges another Document at the specified path in this document
func (d *Document) MergeAt(path string, other *Document) error {
	if other == nil {
		return fmt.Errorf("other document is nil")
	}

	otherRoot, err := other.mappingRoot()
	if err != nil {
		return fmt.Errorf("other document has invalid root: %w", err)
	}

	thisRoot, err := d.mappingRoot()
	if err != nil {
		return fmt.Errorf("this document has invalid root: %w", err)
	}

	// Get or create the target node
	targetNode, err := getOrCreateNode(thisRoot, path)
	if err != nil {
		return fmt.Errorf("failed to get/create target node at path %s: %w", path, err)
	}

	// Convert target to mapping if it's not already
	if targetNode.Kind != yaml.MappingNode {
		targetNode.Kind = yaml.MappingNode
		targetNode.Tag = "!!map"
		targetNode.Content = make([]*yaml.Node, 0)
		targetNode.Value = ""
	}

	err = mergeNodes(targetNode, otherRoot)
	if err != nil {
		return err
	}

	// Update the raw content
	content, err := d.ToBytes()
	if err != nil {
		return err
	}
	d.raw = string(content)

	return nil
}

// mergeNodes merges the content of source node into target node
func mergeNodes(target, source *yaml.Node) error {
	if source == nil {
		return nil
	}

	switch source.Kind {
	case yaml.MappingNode:
		return mergeMappingNodes(target, source)
	case yaml.SequenceNode:
		return mergeSequenceNodes(target, source)
	case yaml.ScalarNode:
		return mergeScalarNodes(target, source)
	default:
		return fmt.Errorf("unsupported node kind for merging: %v", source.Kind)
	}
}

// mergeMappingNodes merges mapping nodes
func mergeMappingNodes(target, source *yaml.Node) error {
	// Ensure target is a mapping node
	if target.Kind != yaml.MappingNode {
		target.Kind = yaml.MappingNode
		target.Tag = "!!map"
		target.Content = make([]*yaml.Node, 0)
		target.Value = ""
	}

	// Merge each key-value pair from source
	for i := 0; i < len(source.Content); i += 2 {
		sourceKey := source.Content[i]
		sourceValue := source.Content[i+1]

		if err := mergeKeyValuePair(target, sourceKey, sourceValue); err != nil {
			return err
		}
	}
	return nil
}

// mergeKeyValuePair merges a single key-value pair into target
func mergeKeyValuePair(target, sourceKey, sourceValue *yaml.Node) error {
	// Find if key exists in target
	for j := 0; j < len(target.Content); j += 2 {
		targetKey := target.Content[j]
		if targetKey.Value == sourceKey.Value {
			// Key exists, merge the values
			targetValue := target.Content[j+1]
			return mergeNodes(targetValue, sourceValue)
		}
	}

	// Key doesn't exist, add it
	keyNode := &yaml.Node{
		Kind:        sourceKey.Kind,
		Tag:         sourceKey.Tag,
		Value:       sourceKey.Value,
		HeadComment: sourceKey.HeadComment,
		LineComment: sourceKey.LineComment,
		FootComment: sourceKey.FootComment,
	}

	valueNode, err := cloneNode(sourceValue)
	if err != nil {
		return err
	}

	target.Content = append(target.Content, keyNode, valueNode)
	return nil
}

// mergeSequenceNodes merges sequence nodes
func mergeSequenceNodes(target, source *yaml.Node) error {
	// For sequences, we replace the entire content but preserve original style if target is already a sequence
	originalStyle := target.Style
	if target.Kind == yaml.SequenceNode && originalStyle != 0 {
		// Preserve the original style (e.g., flow style)
		target.Style = originalStyle
	} else {
		// Use source style
		target.Style = source.Style
	}

	target.Kind = yaml.SequenceNode
	target.Tag = "!!seq"
	target.Content = make([]*yaml.Node, 0, len(source.Content))
	target.Value = ""

	for _, item := range source.Content {
		clonedItem, err := cloneNode(item)
		if err != nil {
			return err
		}
		target.Content = append(target.Content, clonedItem)
	}
	return nil
}

// mergeScalarNodes merges scalar nodes
func mergeScalarNodes(target, source *yaml.Node) error {
	// For scalars, replace the value but preserve comments
	origHeadComment := target.HeadComment
	origLineComment := target.LineComment
	origFootComment := target.FootComment

	target.Kind = source.Kind
	target.Tag = source.Tag
	target.Value = source.Value

	// Preserve original comments if source doesn't have them
	target.HeadComment = preserveComment(source.HeadComment, origHeadComment)
	target.LineComment = preserveComment(source.LineComment, origLineComment)
	target.FootComment = preserveComment(source.FootComment, origFootComment)

	return nil
}

// preserveComment returns source comment if not empty, otherwise returns original
func preserveComment(sourceComment, originalComment string) string {
	if sourceComment == "" && originalComment != "" {
		return originalComment
	}
	return sourceComment
}

// cloneNode creates a deep copy of a YAML node
func cloneNode(node *yaml.Node) (*yaml.Node, error) {
	if node == nil {
		return nil, nil
	}

	clone := &yaml.Node{
		Kind:        node.Kind,
		Style:       node.Style,
		Tag:         node.Tag,
		Value:       node.Value,
		Anchor:      node.Anchor,
		Alias:       node.Alias,
		HeadComment: node.HeadComment,
		LineComment: node.LineComment,
		FootComment: node.FootComment,
		Line:        node.Line,
		Column:      node.Column,
	}

	if len(node.Content) > 0 {
		clone.Content = make([]*yaml.Node, 0, len(node.Content))
		for _, child := range node.Content {
			clonedChild, err := cloneNode(child)
			if err != nil {
				return nil, err
			}
			clone.Content = append(clone.Content, clonedChild)
		}
	}

	return clone, nil
}

// getOrCreateNode gets or creates a node at the specified path
func getOrCreateNode(root *yaml.Node, path string) (*yaml.Node, error) {
	if path == "" {
		return root, nil
	}

	parts := splitPath(path)
	current := root

	for i, part := range parts {
		if isArrayIndex(part) {
			idx, err := parseArrayIndex(part)
			if err != nil {
				return nil, err
			}

			if current.Kind != yaml.SequenceNode {
				current.Kind = yaml.SequenceNode
				current.Tag = "!!seq"
				current.Content = make([]*yaml.Node, 0)
			}

			// Extend array if needed
			for len(current.Content) <= idx {
				current.Content = append(current.Content, &yaml.Node{
					Kind: yaml.MappingNode,
					Tag:  "!!map",
				})
			}

			current = current.Content[idx]
		} else {
			// Handle mapping nodes
			if current.Kind == yaml.ScalarNode {
				current.Kind = yaml.MappingNode
				current.Tag = "!!map"
				current.Value = ""
				current.Content = make([]*yaml.Node, 0)
			}
			if current.Kind != yaml.MappingNode {
				current.Kind = yaml.MappingNode
				current.Tag = "!!map"
				current.Content = make([]*yaml.Node, 0)
			}

			found := false
			for j := 0; j < len(current.Content); j += 2 {
				if current.Content[j].Value == part {
					current = current.Content[j+1]
					found = true
					break
				}
			}

			if !found {
				// Create new key-value pair
				keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: part}
				var valueNode *yaml.Node

				// If this is the last part, create the target node
				if i == len(parts)-1 {
					valueNode = &yaml.Node{
						Kind:    yaml.MappingNode,
						Tag:     "!!map",
						Content: make([]*yaml.Node, 0),
					}
				} else {
					valueNode = &yaml.Node{
						Kind:    yaml.MappingNode,
						Tag:     "!!map",
						Content: make([]*yaml.Node, 0),
					}
				}

				current.Content = append(current.Content, keyNode, valueNode)
				current = valueNode
			}
		}
	}

	return current, nil
}
