package yamler

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// GetArrayLength returns the length of an array at the specified path
func (d *Document) GetArrayLength(path string) (int, error) {
	node, err := d.getNode(path)
	if err != nil {
		return 0, err
	}

	if node.Kind != yaml.SequenceNode {
		return 0, fmt.Errorf("path %s: expected sequence node", path)
	}

	return len(node.Content), nil
}

// getOrCreateArrayNode returns an array node at the specified path, creating it if necessary
func getOrCreateArrayNode(root *yaml.Node, path string) (*yaml.Node, error) {
	parts := splitPath(path)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	current := root
	for i := 0; i < len(parts); i++ {
		part := parts[i]
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
			for len(current.Content) <= idx {
				current.Content = append(current.Content, &yaml.Node{
					Kind:  yaml.MappingNode,
					Tag:   "!!map",
					Style: 0, // Block style
				})
			}
			current = current.Content[idx]
			continue
		}

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

			// If this is the last part, create an array
			if i == len(parts)-1 {
				valueNode = &yaml.Node{
					Kind:    yaml.SequenceNode,
					Tag:     "!!seq",
					Style:   yaml.FlowStyle, // Flow style for new arrays
					Content: make([]*yaml.Node, 0),
				}
			} else {
				valueNode = &yaml.Node{
					Kind:    yaml.MappingNode,
					Tag:     "!!map",
					Style:   0, // Block style
					Content: make([]*yaml.Node, 0),
				}
			}

			current.Content = append(current.Content, keyNode, valueNode)
			current = valueNode
		}
	}

	// Ensure the final node is an array
	if current.Kind != yaml.SequenceNode {
		current.Kind = yaml.SequenceNode
		current.Tag = "!!seq"
		current.Style = yaml.FlowStyle // Flow style for new arrays
		current.Content = make([]*yaml.Node, 0)
	}

	return current, nil
}

// AppendToArray appends a value to an array at the specified path
func (d *Document) AppendToArray(path string, value interface{}) error {
	root, err := d.mappingRoot()
	if err != nil {
		return err
	}

	// First check if path exists
	parts := splitPath(path)
	if len(parts) > 0 {
		// Check if all but the last part exist
		parentPath := strings.Join(parts[:len(parts)-1], ".")

		if parentPath != "" {
			// Check if parent exists
			_, err := d.getNode(parentPath)
			if err != nil {
				// Parent doesn't exist, create the array
				arrayNode, err := getOrCreateArrayNode(root, path)
				if err != nil {
					return err
				}

				valueNode, err := interfaceToNode(value)
				if err != nil {
					return err
				}

				arrayNode.Content = append(arrayNode.Content, valueNode)

				content, err := d.ToBytes()
				if err != nil {
					return err
				}
				d.raw = string(content)
				return nil
			}
		}

		// Check if the target path exists
		existingNode, err := d.getNode(path)
		if err != nil {
			// Path doesn't exist, create array
			arrayNode, err := getOrCreateArrayNode(root, path)
			if err != nil {
				return err
			}

			valueNode, err := interfaceToNode(value)
			if err != nil {
				return err
			}

			arrayNode.Content = append(arrayNode.Content, valueNode)

			content, err := d.ToBytes()
			if err != nil {
				return err
			}
			d.raw = string(content)
			return nil
		}

		// Path exists, check if it's an array
		if existingNode.Kind != yaml.SequenceNode {
			return fmt.Errorf("path %s: not an array", path)
		}

		// It's an array, append to it
		valueNode, err := interfaceToNode(value)
		if err != nil {
			return err
		}

		existingNode.Content = append(existingNode.Content, valueNode)

		content, err := d.ToBytes()
		if err != nil {
			return err
		}
		d.raw = string(content)
		return nil
	}

	// Fallback to original behavior for empty path
	arrayNode, err := getOrCreateArrayNode(root, path)
	if err != nil {
		return err
	}

	valueNode, err := interfaceToNode(value)
	if err != nil {
		return err
	}

	arrayNode.Content = append(arrayNode.Content, valueNode)

	content, err := d.ToBytes()
	if err != nil {
		return err
	}
	d.raw = string(content)
	return nil
}

// RemoveFromArray removes an element from an array at the specified path and index
func (d *Document) RemoveFromArray(path string, index int) error {
	root, err := d.mappingRoot()
	if err != nil {
		return err
	}
	arrayNode, err := getArrayNode(root, path)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(arrayNode.Content) {
		return fmt.Errorf("array index out of bounds: %d", index)
	}

	arrayNode.Content = append(arrayNode.Content[:index], arrayNode.Content[index+1:]...)

	content, err := d.ToBytes()
	if err != nil {
		return err
	}
	d.raw = string(content)
	return nil
}

// UpdateArrayElement updates an element in an array at the specified path and index
func (d *Document) UpdateArrayElement(path string, index int, value interface{}) error {
	root, err := d.mappingRoot()
	if err != nil {
		return err
	}
	arrayNode, err := getArrayNode(root, path)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(arrayNode.Content) {
		return fmt.Errorf("array index out of bounds: %d", index)
	}

	valueNode, err := interfaceToNode(value)
	if err != nil {
		return err
	}

	// Preserve original comments
	valueNode.HeadComment = arrayNode.Content[index].HeadComment
	valueNode.LineComment = arrayNode.Content[index].LineComment
	valueNode.FootComment = arrayNode.Content[index].FootComment

	// Set block style for the new value
	valueNode.Style = 0 // Block style

	arrayNode.Content[index] = valueNode

	content, err := d.ToBytes()
	if err != nil {
		return err
	}
	d.raw = string(content)
	return nil
}

// InsertIntoArray inserts a value into an array at the specified path and index
func (d *Document) InsertIntoArray(path string, index int, value interface{}) error {
	root, err := d.mappingRoot()
	if err != nil {
		return err
	}

	// First check if the path exists and is an array
	arrayNode, err := getArrayNode(root, path)
	if err != nil {
		return err
	}

	if index < 0 || index > len(arrayNode.Content) {
		return fmt.Errorf("array index out of bounds: %d", index)
	}

	valueNode, err := interfaceToNode(value)
	if err != nil {
		return err
	}

	// Set block style for the new value
	valueNode.Style = 0 // Block style

	arrayNode.Content = append(arrayNode.Content[:index], append([]*yaml.Node{valueNode}, arrayNode.Content[index:]...)...)

	content, err := d.ToBytes()
	if err != nil {
		return err
	}
	d.raw = string(content)
	return nil
}

// GetArrayElement returns an element from an array at the specified path and index
func (d *Document) GetArrayElement(path string, index int) (interface{}, error) {
	node, err := d.getNode(path)
	if err != nil {
		return nil, err
	}

	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("path %s: expected sequence node", path)
	}

	if index < 0 || index >= len(node.Content) {
		return nil, fmt.Errorf("path %s: index %d out of bounds", path, index)
	}

	return nodeToInterface(node.Content[index])
}

// GetTypedArrayElement returns a typed element from an array at the specified path and index
func (d *Document) GetTypedArrayElement(path string, index int, targetType string) (interface{}, error) {
	value, err := d.GetArrayElement(path, index)
	if err != nil {
		return nil, err
	}

	switch targetType {
	case "string":
		str, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("path %s[%d]: expected string, got %T", path, index, value)
		}
		return str, nil
	case "int":
		switch v := value.(type) {
		case int64:
			return v, nil
		case string:
			return strconv.ParseInt(v, 10, 64)
		default:
			return nil, fmt.Errorf("path %s[%d]: expected integer, got %T", path, index, value)
		}
	case "float":
		switch v := value.(type) {
		case float64:
			return v, nil
		case int64:
			return float64(v), nil
		case string:
			return strconv.ParseFloat(v, 64)
		default:
			return nil, fmt.Errorf("path %s[%d]: expected float, got %T", path, index, value)
		}
	case "bool":
		switch v := value.(type) {
		case bool:
			return v, nil
		case string:
			v = strings.ToLower(v)
			switch v {
			case "true", "yes", "1", "on":
				return true, nil
			case "false", "no", "0", "off":
				return false, nil
			default:
				return nil, fmt.Errorf("path %s[%d]: invalid boolean value: %s", path, index, v)
			}
		default:
			return nil, fmt.Errorf("path %s[%d]: expected boolean, got %T", path, index, value)
		}
	default:
		return nil, fmt.Errorf("path %s[%d]: unsupported type: %s", path, index, targetType)
	}
}

// getNode returns the YAML node at the specified path
func (d *Document) getNode(path string) (*yaml.Node, error) {
	root, err := d.mappingRoot()
	if err != nil {
		return nil, err
	}
	if path == "" {
		return root, nil
	}

	node := root
	parts := strings.Split(path, ".")
	for _, part := range parts {
		if node.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("path %s: expected mapping node", path)
		}
		found := false
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == part {
				node = node.Content[i+1]
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("path %s: key %s not found", path, part)
		}
	}
	return node, nil
}

// getArrayNode returns an array node at the specified path
func getArrayNode(root *yaml.Node, path string) (*yaml.Node, error) {
	parts := splitPath(path)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	current := root
	for i := 0; i < len(parts); i++ {
		part := parts[i]
		if isArrayIndex(part) {
			idx, err := parseArrayIndex(part)
			if err != nil {
				return nil, err
			}
			if current.Kind != yaml.SequenceNode {
				return nil, fmt.Errorf("path %s: not an array", path)
			}
			if idx < 0 || idx >= len(current.Content) {
				return nil, fmt.Errorf("array index out of bounds: %d", idx)
			}
			current = current.Content[idx]
			continue
		}
		if current.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("path %s: not a map", path)
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
			return nil, fmt.Errorf("path %s: key not found", path)
		}
	}

	if current.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("path %s: not an array", path)
	}
	return current, nil
}
