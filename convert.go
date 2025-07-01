package yamler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// splitPath splits a path into parts, handling array indices
func splitPath(path string) []string {
	if path == "" {
		return nil
	}

	parts := strings.Split(path, ".")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		// Handle array indices
		if strings.Contains(part, "[") {
			// Find the array name and index
			idx := strings.Index(part, "[")
			if idx > 0 {
				arrayName := part[:idx]
				arrayIndex := part[idx:]
				result = append(result, arrayName, arrayIndex)
			} else {
				result = append(result, part)
			}
		} else {
			result = append(result, part)
		}
	}

	return result
}

// nodeToInterface converts a YAML node to a Go interface{}
func nodeToInterface(node *yaml.Node) (interface{}, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		return scalarToInterface(node)
	case yaml.SequenceNode:
		var result []interface{}
		for _, item := range node.Content {
			value, err := nodeToInterface(item)
			if err != nil {
				return nil, err
			}
			result = append(result, value)
		}
		return result, nil
	case yaml.MappingNode:
		result := make(map[string]interface{})
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			value, err := nodeToInterface(node.Content[i+1])
			if err != nil {
				return nil, err
			}
			result[key] = value
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported node kind: %v", node.Kind)
	}
}

// scalarToInterface converts a scalar YAML node to a Go interface{}
func scalarToInterface(node *yaml.Node) (interface{}, error) {
	switch node.Tag {
	case "!!str":
		return node.Value, nil
	case "!!int":
		return strconv.ParseInt(node.Value, 10, 64)
	case "!!float":
		return strconv.ParseFloat(node.Value, 64)
	case "!!bool":
		return strconv.ParseBool(node.Value)
	case "!!null":
		return nil, nil
	default:
		return node.Value, nil
	}
}

// interfaceToNode converts a Go interface{} to a YAML node
func interfaceToNode(v interface{}) (*yaml.Node, error) {
	switch val := v.(type) {
	case string:
		return createScalarNode("!!str", val), nil
	case int:
		return createScalarNode("!!int", fmt.Sprintf("%d", val)), nil
	case int64:
		return createScalarNode("!!int", fmt.Sprintf("%d", val)), nil
	case float64:
		return createScalarNode("!!float", fmt.Sprintf("%g", val)), nil
	case bool:
		return createScalarNode("!!bool", fmt.Sprintf("%t", val)), nil
	case []interface{}:
		return createGenericSliceNode(val)
	case []string:
		return createStringSliceNode(val)
	case []int64:
		return createInt64SliceNode(val)
	case []float64:
		return createFloat64SliceNode(val)
	case []bool:
		return createBoolSliceNode(val)
	case map[string]interface{}:
		return createMapNode(val)
	case []map[string]interface{}:
		return createMapSliceNode(val)
	case nil:
		return createNullNode(), nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

// createScalarNode creates a scalar YAML node
func createScalarNode(tag, value string) *yaml.Node {
	return &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   tag,
		Value: value,
	}
}

// createNullNode creates a null YAML node
func createNullNode() *yaml.Node {
	return &yaml.Node{
		Kind: yaml.ScalarNode,
		Tag:  "!!null",
	}
}

// createGenericSliceNode creates a sequence node from []interface{}
func createGenericSliceNode(val []interface{}) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
		Tag:  "!!seq",
	}
	for _, item := range val {
		itemNode, err := interfaceToNode(item)
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, itemNode)
	}
	return node, nil
}

// createStringSliceNode creates a sequence node from []string
func createStringSliceNode(val []string) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
		Tag:  "!!seq",
	}
	for _, item := range val {
		node.Content = append(node.Content, createScalarNode("!!str", item))
	}
	return node, nil
}

// createInt64SliceNode creates a sequence node from []int64
func createInt64SliceNode(val []int64) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
		Tag:  "!!seq",
	}
	for _, item := range val {
		node.Content = append(node.Content, createScalarNode("!!int", fmt.Sprintf("%d", item)))
	}
	return node, nil
}

// createFloat64SliceNode creates a sequence node from []float64
func createFloat64SliceNode(val []float64) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
		Tag:  "!!seq",
	}
	for _, item := range val {
		node.Content = append(node.Content, createScalarNode("!!float", fmt.Sprintf("%g", item)))
	}
	return node, nil
}

// createBoolSliceNode creates a sequence node from []bool
func createBoolSliceNode(val []bool) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
		Tag:  "!!seq",
	}
	for _, item := range val {
		node.Content = append(node.Content, createScalarNode("!!bool", fmt.Sprintf("%t", item)))
	}
	return node, nil
}

// createMapNode creates a mapping node from map[string]interface{}
func createMapNode(val map[string]interface{}) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}
	// Sort keys to ensure consistent output for new objects
	// Order preservation for existing objects is handled elsewhere
	keys := make([]string, 0, len(val))
	for key := range val {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := val[key]
		keyNode := createScalarNode("!!str", key)
		valueNode, err := interfaceToNode(value)
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, keyNode, valueNode)
	}
	return node, nil
}

// createMapSliceNode creates a sequence node from []map[string]interface{}
func createMapSliceNode(val []map[string]interface{}) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.SequenceNode,
		Tag:  "!!seq",
	}
	for _, item := range val {
		itemNode, err := interfaceToNode(item)
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, itemNode)
	}
	return node, nil
}
