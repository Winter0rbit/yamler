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
			// Split on [ and ]
			subparts := strings.FieldsFunc(part, func(r rune) bool {
				return r == '[' || r == ']'
			})
			result = append(result, subparts...)
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
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: val,
		}, nil
	case int:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!int",
			Value: fmt.Sprintf("%d", val),
		}, nil
	case int64:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!int",
			Value: fmt.Sprintf("%d", val),
		}, nil
	case float64:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!float",
			Value: fmt.Sprintf("%g", val),
		}, nil
	case bool:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!bool",
			Value: fmt.Sprintf("%t", val),
		}, nil
	case []interface{}:
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
	case []string:
		node := &yaml.Node{
			Kind: yaml.SequenceNode,
			Tag:  "!!seq",
		}
		for _, item := range val {
			node.Content = append(node.Content, &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: item,
			})
		}
		return node, nil
	case []int64:
		node := &yaml.Node{
			Kind: yaml.SequenceNode,
			Tag:  "!!seq",
		}
		for _, item := range val {
			node.Content = append(node.Content, &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!int",
				Value: fmt.Sprintf("%d", item),
			})
		}
		return node, nil
	case []float64:
		node := &yaml.Node{
			Kind: yaml.SequenceNode,
			Tag:  "!!seq",
		}
		for _, item := range val {
			node.Content = append(node.Content, &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!float",
				Value: fmt.Sprintf("%g", item),
			})
		}
		return node, nil
	case []bool:
		node := &yaml.Node{
			Kind: yaml.SequenceNode,
			Tag:  "!!seq",
		}
		for _, item := range val {
			node.Content = append(node.Content, &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!bool",
				Value: fmt.Sprintf("%t", item),
			})
		}
		return node, nil
	case map[string]interface{}:
		node := &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
		}
		// Sort keys to ensure consistent output
		keys := make([]string, 0, len(val))
		for key := range val {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := val[key]
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: key,
			}
			valueNode, err := interfaceToNode(value)
			if err != nil {
				return nil, err
			}
			node.Content = append(node.Content, keyNode, valueNode)
		}
		return node, nil
	case []map[string]interface{}:
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
	case nil:
		return &yaml.Node{
			Kind: yaml.ScalarNode,
			Tag:  "!!null",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}
