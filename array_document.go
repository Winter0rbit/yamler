// Support for documents whose root node is a sequence (e.g. Ansible playbooks).

package yamler

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// isArrayRoot checks if the document root is an array
func (d *Document) isArrayRoot() bool {
	if d.root == nil || len(d.root.Content) == 0 {
		return false
	}
	return d.root.Content[0].Kind == yaml.SequenceNode
}

// arrayRoot returns the root SequenceNode of the document for array documents
func (d *Document) sequenceRoot() (*yaml.Node, error) {
	if d.root == nil || len(d.root.Content) == 0 {
		return nil, fmt.Errorf("empty document root")
	}
	root := d.root.Content[0]
	if root.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("root is not a sequence node")
	}
	return root, nil
}

// SetArrayElement sets a value in an array document at the specified index and path
func (d *Document) SetArrayElement(index int, path string, value interface{}) error {
	// Do not preserve document separators for array element operations
	d.preserveDocumentSeparator = false

	if !d.isArrayRoot() {
		return fmt.Errorf("document root is not an array")
	}

	root, err := d.sequenceRoot()
	if err != nil {
		return err
	}

	if index < 0 || index >= len(root.Content) {
		return fmt.Errorf("array index %d out of bounds (length: %d)", index, len(root.Content))
	}

	element := root.Content[index]
	if path == "" {
		// Replace entire element
		newNode, err := interfaceToNode(value)
		if err != nil {
			return err
		}
		root.Content[index] = newNode
		return nil
	}

	// Set value in the element (assuming it's a mapping)
	if element.Kind != yaml.MappingNode {
		return fmt.Errorf("array element at index %d is not a mapping", index)
	}

	return d.setValueInNode(element, path, value)
}

// GetArrayDocumentElement gets a value from an array document at the specified index and path
func (d *Document) GetArrayDocumentElement(index int, path string) (interface{}, error) {
	if !d.isArrayRoot() {
		return nil, fmt.Errorf("document root is not an array")
	}

	root, err := d.sequenceRoot()
	if err != nil {
		return nil, err
	}

	if index < 0 || index >= len(root.Content) {
		return nil, fmt.Errorf("array index %d out of bounds (length: %d)", index, len(root.Content))
	}

	element := root.Content[index]
	if path == "" {
		// Return entire element
		return nodeToInterface(element)
	}

	// Get value from the element
	return d.getValueFromNode(element, path)
}

// AddArrayElement adds a new element to an array document
func (d *Document) AddArrayElement(value interface{}) error {
	// Do not preserve document separators for array element operations
	d.preserveDocumentSeparator = false

	if !d.isArrayRoot() {
		return fmt.Errorf("document root is not an array")
	}

	root, err := d.sequenceRoot()
	if err != nil {
		return err
	}

	newNode, err := interfaceToNode(value)
	if err != nil {
		return err
	}

	root.Content = append(root.Content, newNode)
	return nil
}

// setValueInNode sets a value in a specific node using a path
func (d *Document) setValueInNode(node *yaml.Node, path string, value interface{}) error {
	// This is a simplified version - could be extended to use the full Set logic
	parts := parsePath(path)
	current := node

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - set the value
			return d.setDirectValue(current, part, value)
		}

		// Navigate to the next level
		found := false
		for j := 0; j < len(current.Content); j += 2 {
			if current.Content[j].Value == part {
				current = current.Content[j+1]
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("path not found: %s", path)
		}
	}

	return nil
}

// getValueFromNode gets a value from a specific node using a path
func (d *Document) getValueFromNode(node *yaml.Node, path string) (interface{}, error) {
	parts := parsePath(path)
	current := node

	for _, part := range parts {
		if current.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("cannot navigate path in non-mapping node")
		}

		// Find the key
		found := false
		for j := 0; j < len(current.Content); j += 2 {
			if current.Content[j].Value == part {
				current = current.Content[j+1]
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("key not found: %s", part)
		}
	}

	return nodeToInterface(current)
}

// setDirectValue sets a direct value in a mapping node
func (d *Document) setDirectValue(node *yaml.Node, key string, value interface{}) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("cannot set value in non-mapping node")
	}

	// Find existing key
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			// Update existing value
			newNode, err := interfaceToNode(value)
			if err != nil {
				return err
			}
			node.Content[i+1] = newNode
			return nil
		}
	}

	// Add new key-value pair
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: key,
	}
	valueNode, err := interfaceToNode(value)
	if err != nil {
		return err
	}

	node.Content = append(node.Content, keyNode, valueNode)
	return nil
}
