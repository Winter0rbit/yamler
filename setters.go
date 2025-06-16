package yamler

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Set sets a value at the specified path
func (d *Document) Set(path string, value interface{}) error {
	// Set document separator preservation flag for Set() operations
	d.preserveDocumentSeparator = true

	// Handle both mapping and array root documents
	if d.isArrayRoot() {
		// Store the current preserveDocumentSeparator setting
		preserveFlag := d.preserveDocumentSeparator

		// For array root documents, prepend [0] to the path if it doesn't start with array index
		if path != "" && !strings.HasPrefix(path, "[") {
			path = "[0]." + path
		}
		err := d.SetArrayElement(0, path[4:], value) // Remove "[0]." prefix

		// Restore the preserveDocumentSeparator setting after SetArrayElement
		d.preserveDocumentSeparator = preserveFlag
		return err
	}

	root, err := d.mappingRoot()
	if err != nil {
		return err
	}
	parts := splitPath(path)
	if len(parts) == 0 {
		// Empty path — replace entire root
		valueNode, err := interfaceToNode(value)
		if err != nil {
			return err
		}
		root.Content = valueNode.Content
		content, err := d.ToBytes()
		if err != nil {
			return err
		}
		d.raw = string(content)
		return nil
	}

	parent, key, err := getOrCreateParentNode(root, parts)
	if err != nil {
		return err
	}

	valueNode, err := interfaceToNode(value)
	if err != nil {
		return err
	}

	// Convert scalar parent to mapping if needed
	if parent.Kind == yaml.ScalarNode {
		parent.Kind = yaml.MappingNode
		parent.Tag = "!!map"
		parent.Value = ""
		parent.Content = make([]*yaml.Node, 0)
	}

	if parent.Kind == yaml.MappingNode {
		found := false
		for i := 0; i < len(parent.Content); i += 2 {
			if parent.Content[i].Value == key {
				valueNode.HeadComment = parent.Content[i+1].HeadComment
				valueNode.LineComment = parent.Content[i+1].LineComment
				valueNode.FootComment = parent.Content[i+1].FootComment
				parent.Content[i+1] = valueNode
				found = true
				break
			}
		}
		if !found {
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: key,
			}
			parent.Content = append(parent.Content, keyNode, valueNode)
		}
	} else if parent.Kind == yaml.SequenceNode {
		idx, err := parseArrayIndex(key)
		if err != nil {
			return err
		}
		if idx < 0 || idx >= len(parent.Content) {
			return fmt.Errorf("array index out of bounds: %d", idx)
		}
		valueNode.HeadComment = parent.Content[idx].HeadComment
		valueNode.LineComment = parent.Content[idx].LineComment
		valueNode.FootComment = parent.Content[idx].FootComment
		parent.Content[idx] = valueNode
	} else {
		return fmt.Errorf("parent node is not mapping or sequence")
	}

	content, err := d.ToBytes()
	if err != nil {
		return err
	}
	d.raw = string(content)
	return nil
}

// getOrCreateParentNode returns the parent node and key for replacement/addition
func getOrCreateParentNode(root *yaml.Node, parts []string) (*yaml.Node, string, error) {
	current := root
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		// Array
		if isArrayIndex(part) {
			idx, err := parseArrayIndex(part)
			if err != nil {
				return nil, "", err
			}
			if current.Kind != yaml.SequenceNode {
				current.Kind = yaml.SequenceNode
				current.Content = make([]*yaml.Node, 0)
			}
			for len(current.Content) <= idx {
				current.Content = append(current.Content, &yaml.Node{
					Kind: yaml.MappingNode,
					Tag:  "!!map",
				})
			}
			current = current.Content[idx]
			continue
		}
		// Map
		if current.Kind == yaml.ScalarNode {
			// If scalar but map expected — replace with map
			current.Kind = yaml.MappingNode
			current.Tag = "!!map"
			current.Value = ""
			current.Content = make([]*yaml.Node, 0)
		}
		if current.Kind != yaml.MappingNode {
			current.Kind = yaml.MappingNode
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
			newNode := &yaml.Node{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
			}
			current.Content = append(current.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: part},
				newNode,
			)
			current = newNode
		}
	}
	return current, parts[len(parts)-1], nil
}

// isArrayIndex checks if a path part is an array index
func isArrayIndex(part string) bool {
	return len(part) >= 2 && part[0] == '[' && part[len(part)-1] == ']'
}

// parseArrayIndex extracts the index from a path part
func parseArrayIndex(part string) (int, error) {
	indexStr := part[1 : len(part)-1]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return 0, fmt.Errorf("invalid array index: %s", part)
	}
	if index < 0 {
		return 0, fmt.Errorf("negative array index: %d", index)
	}
	return index, nil
}

// SetString sets a string value in the YAML document
func (d *Document) SetString(path string, value string) error {
	return d.Set(path, value)
}

// SetInt sets an integer value in the YAML document
func (d *Document) SetInt(path string, value int64) error {
	return d.Set(path, value)
}

// SetFloat sets a float value in the YAML document
func (d *Document) SetFloat(path string, value float64) error {
	return d.Set(path, value)
}

// SetBool sets a boolean value in the YAML document
func (d *Document) SetBool(path string, value bool) error {
	return d.Set(path, value)
}

// SetStringSlice sets a string slice in the YAML document
func (d *Document) SetStringSlice(path string, value []string) error {
	return d.Set(path, value)
}

// SetIntSlice sets an integer slice in the YAML document
func (d *Document) SetIntSlice(path string, value []int64) error {
	return d.Set(path, value)
}

// SetFloatSlice sets a float slice in the YAML document
func (d *Document) SetFloatSlice(path string, value []float64) error {
	return d.Set(path, value)
}

// SetBoolSlice sets a boolean slice in the YAML document
func (d *Document) SetBoolSlice(path string, value []bool) error {
	return d.Set(path, value)
}

// SetMapSlice sets a slice of maps in the YAML document
func (d *Document) SetMapSlice(path string, value []map[string]interface{}) error {
	return d.Set(path, value)
}
