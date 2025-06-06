package yamler

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Document represents a YAML document with preserved formatting
type Document struct {
	root *yaml.Node
	raw  string
}

// LoadFile loads a YAML file and preserves its formatting
func LoadFile(filename string) (*Document, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return LoadBytes(content)
}

// LoadBytes loads a YAML document from a byte slice and preserves its formatting
func LoadBytes(content []byte) (*Document, error) {
	return Load(string(content))
}

// Load parses a YAML string and preserves its formatting
func Load(content string) (*Document, error) {
	if content == "" {
		// Create empty document
		return &Document{
			root: &yaml.Node{
				Kind: yaml.DocumentNode,
				Content: []*yaml.Node{
					{
						Kind: yaml.MappingNode,
						Tag:  "!!map",
					},
				},
			},
		}, nil
	}

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(content), &node); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &Document{
		root: &node,
		raw:  content,
	}, nil
}

// Get returns a value at the specified path
func (d *Document) Get(path string) (interface{}, error) {
	if path == "" {
		return nodeToInterface(d.root)
	}

	parts := strings.Split(path, ".")
	node := d.root

	// Skip the document node
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}

	for _, part := range parts {
		found := false
		if node.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("invalid path: %s is not a mapping", part)
		}

		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == part {
				node = node.Content[i+1]
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("key not found: %s", part)
		}
	}

	return nodeToInterface(node)
}

// Set sets a value at the specified path while preserving formatting
func (d *Document) Set(path string, value interface{}) error {
	if path == "" {
		newNode, err := interfaceToNode(value)
		if err != nil {
			return err
		}
		// Preserve style from root
		if d.root != nil {
			newNode.Style = d.root.Style
		}
		d.root = newNode
		return nil
	}

	// Initialize root if it's nil
	if d.root == nil {
		d.root = &yaml.Node{
			Kind: yaml.DocumentNode,
			Content: []*yaml.Node{
				{
					Kind: yaml.MappingNode,
					Tag:  "!!map",
				},
			},
		}
	}

	parts := strings.Split(path, ".")
	node := d.root

	// Skip the document node
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) == 0 {
			node.Content = append(node.Content, &yaml.Node{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
			})
		}
		node = node.Content[0]
	}

	current := node
	for i, part := range parts {
		if current.Kind != yaml.MappingNode {
			return fmt.Errorf("invalid path: %s is not a mapping", strings.Join(parts[:i+1], "."))
		}

		if i == len(parts)-1 {
			// Last part - set the value
			found := false
			for i := 0; i < len(current.Content); i += 2 {
				if current.Content[i].Value == part {
					// Preserve style and formatting from the original value
					style := current.Content[i+1].Style
					footComment := current.Content[i+1].FootComment
					headComment := current.Content[i+1].HeadComment
					lineComment := current.Content[i+1].LineComment

					newNode, err := interfaceToNode(value)
					if err != nil {
						return err
					}
					newNode.Style = style
					newNode.FootComment = footComment
					newNode.HeadComment = headComment
					newNode.LineComment = lineComment
					current.Content[i+1] = newNode
					found = true
					break
				}
			}

			if !found {
				// Add new key-value pair
				keyNode := &yaml.Node{
					Kind:  yaml.ScalarNode,
					Value: part,
					Tag:   "!!str",
				}
				valueNode, err := interfaceToNode(value)
				if err != nil {
					return err
				}
				current.Content = append(current.Content, keyNode, valueNode)
			}
			return nil
		}

		// Create or find the next level
		found := false
		var next *yaml.Node
		for i := 0; i < len(current.Content); i += 2 {
			if current.Content[i].Value == part {
				next = current.Content[i+1]
				found = true
				break
			}
		}

		if !found {
			// Create new nested structure
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: part,
				Tag:   "!!str",
			}
			next = &yaml.Node{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
			}
			current.Content = append(current.Content, keyNode, next)
		}
		current = next
	}

	return nil
}

// Save writes the YAML document to a file while preserving formatting
func (d *Document) Save(filename string) error {
	content, err := d.ToBytes()
	if err != nil {
		return err
	}

	return os.WriteFile(filename, content, 0644)
}

// ToBytes returns the YAML document as a byte slice while preserving formatting
func (d *Document) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	defer encoder.Close()
	encoder.SetIndent(2)

	if err := encoder.Encode(d.root); err != nil {
		return nil, fmt.Errorf("failed to encode YAML: %w", err)
	}

	return buf.Bytes(), nil
}

// String returns the YAML document as a string while preserving formatting
func (d *Document) String() (string, error) {
	content, err := d.ToBytes()
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// helper functions to convert between yaml.Node and interface{}
func nodeToInterface(node *yaml.Node) (interface{}, error) {
	if node == nil {
		return nil, nil
	}

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		return nodeToInterface(node.Content[0])
	case yaml.AliasNode:
		return nodeToInterface(node.Alias)
	case yaml.ScalarNode:
		return parseScalar(node)
	case yaml.SequenceNode:
		result := make([]interface{}, len(node.Content))
		for i, v := range node.Content {
			var err error
			result[i], err = nodeToInterface(v)
			if err != nil {
				return nil, err
			}
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

// parseScalar converts YAML scalar node to appropriate Go type
func parseScalar(node *yaml.Node) (interface{}, error) {
	switch node.Tag {
	case "!!null":
		return nil, nil
	case "!!bool":
		return node.Value == "true", nil
	case "!!int":
		return node.Value, nil
	case "!!float":
		return node.Value, nil
	case "!!timestamp":
		return node.Value, nil
	case "!!str":
		return node.Value, nil
	default:
		// Try to infer the type
		switch node.Value {
		case "true":
			return true, nil
		case "false":
			return false, nil
		case "null", "":
			return nil, nil
		default:
			return node.Value, nil
		}
	}
}

func interfaceToNode(v interface{}) (*yaml.Node, error) {
	if v == nil {
		return &yaml.Node{
			Kind: yaml.ScalarNode,
			Tag:  "!!null",
		}, nil
	}

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
			Value: fmt.Sprintf("%v", val),
		}, nil
	case []interface{}:
		content := make([]*yaml.Node, len(val))
		for i, item := range val {
			node, err := interfaceToNode(item)
			if err != nil {
				return nil, err
			}
			content[i] = node
		}
		return &yaml.Node{
			Kind:    yaml.SequenceNode,
			Tag:     "!!seq",
			Content: content,
		}, nil
	case map[string]interface{}:
		content := make([]*yaml.Node, 0, len(val)*2)
		for k, v := range val {
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: k,
			}
			valueNode, err := interfaceToNode(v)
			if err != nil {
				return nil, err
			}
			content = append(content, keyNode, valueNode)
		}
		return &yaml.Node{
			Kind:    yaml.MappingNode,
			Tag:     "!!map",
			Content: content,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}
