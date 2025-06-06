package yamler

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// mappingRoot возвращает корневой MappingNode документа
func (d *Document) mappingRoot() (*yaml.Node, error) {
	if d.root == nil || len(d.root.Content) == 0 {
		return nil, fmt.Errorf("empty document root")
	}
	root := d.root.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("root is not a mapping node")
	}
	return root, nil
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

// Save writes the YAML document to a file while preserving formatting
func (d *Document) Save(filename string) error {
	content, err := d.ToBytes()
	if err != nil {
		return err
	}

	return os.WriteFile(filename, content, 0644)
}

// ToBytes converts the document to bytes while preserving formatting
func (d *Document) ToBytes() ([]byte, error) {
	if d.root == nil || len(d.root.Content) == 0 {
		return []byte{}, nil
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(4) // Set 4-space indentation

	if err := encoder.Encode(d.root); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// OrderedMap preserves key order for YAML marshaling
type OrderedMap struct {
	Keys   []string
	Values map[string]interface{}
}

// MarshalYAML implements yaml.Marshaler interface
func (om OrderedMap) MarshalYAML() (interface{}, error) {
	result := make(map[string]interface{})
	for _, key := range om.Keys {
		result[key] = om.Values[key]
	}
	return result, nil
}

// nodeToOrderedInterface converts a YAML node to a Go interface{} preserving order
func nodeToOrderedInterface(node *yaml.Node) (interface{}, error) {
	switch node.Kind {
	case yaml.ScalarNode:
		return scalarToInterface(node)
	case yaml.SequenceNode:
		var result []interface{}
		for _, item := range node.Content {
			value, err := nodeToOrderedInterface(item)
			if err != nil {
				return nil, err
			}
			result = append(result, value)
		}
		return result, nil
	case yaml.MappingNode:
		om := OrderedMap{
			Keys:   make([]string, 0),
			Values: make(map[string]interface{}),
		}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i].Value
			value, err := nodeToOrderedInterface(node.Content[i+1])
			if err != nil {
				return nil, err
			}
			om.Keys = append(om.Keys, key)
			om.Values[key] = value
		}
		return om, nil
	default:
		return nil, fmt.Errorf("unsupported node kind: %v", node.Kind)
	}
}

// convertSequencesToBlockStyle recursively converts sequence nodes to block style
func convertSequencesToBlockStyle(node *yaml.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.SequenceNode:
		// Create a completely new sequence node to avoid any flow style history
		oldContent := node.Content
		node.Content = make([]*yaml.Node, len(oldContent))
		for i, child := range oldContent {
			// Create new nodes for each element
			newChild := &yaml.Node{
				Kind:        child.Kind,
				Style:       0, // Force block style
				Tag:         child.Tag,
				Value:       child.Value,
				Anchor:      child.Anchor,
				Alias:       child.Alias,
				HeadComment: child.HeadComment,
				LineComment: child.LineComment,
				FootComment: child.FootComment,
				Line:        child.Line,
				Column:      child.Column,
			}
			if child.Content != nil {
				newChild.Content = make([]*yaml.Node, len(child.Content))
				copy(newChild.Content, child.Content)
			}
			node.Content[i] = newChild
			convertSequencesToBlockStyle(newChild)
		}
		node.Style = 0
		node.Tag = "!!seq"
	case yaml.MappingNode:
		// Process children only
		for _, child := range node.Content {
			convertSequencesToBlockStyle(child)
		}
	}
}

// forceBlockStyle recursively forces block style for all sequence and mapping nodes
func forceBlockStyle(node *yaml.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.SequenceNode:
		// Clear any flow style flags and ensure block style
		node.Style = 0
		node.Tag = "!!seq"
		// Process all children
		for _, child := range node.Content {
			forceBlockStyle(child)
		}
	case yaml.MappingNode:
		// Clear any flow style flags and ensure block style
		node.Style = 0
		node.Tag = "!!map"
		// Process all children
		for _, child := range node.Content {
			forceBlockStyle(child)
		}
	case yaml.ScalarNode:
		// For scalars, ensure no flow styling
		if node.Style == yaml.FlowStyle {
			node.Style = 0
		}
	}
}

// convertFlowToBlock recursively converts flow-style sequences to block style by setting line breaks
func convertFlowToBlock(node *yaml.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.SequenceNode:
		// Force each element to be on a separate line by clearing any flow markers
		node.Style = 0
		node.Tag = "!!seq"
		for _, elem := range node.Content {
			convertFlowToBlock(elem)
		}
	case yaml.MappingNode:
		node.Style = 0
		node.Tag = "!!map"
		for _, elem := range node.Content {
			convertFlowToBlock(elem)
		}
	default:
		// For scalar nodes, ensure they don't have flow styling
		if node.Kind == yaml.ScalarNode {
			node.Style = 0
		}
	}
}

// copyNodeStyles recursively copies styles and comments from source to target node
func copyNodeStyles(source, target *yaml.Node) {
	if source == nil || target == nil {
		return
	}

	// Copy comments
	target.HeadComment = source.HeadComment
	target.LineComment = source.LineComment
	target.FootComment = source.FootComment

	// For mapping nodes, copy styles for both keys and values
	if source.Kind == yaml.MappingNode && target.Kind == yaml.MappingNode {
		for i := 0; i < len(source.Content) && i < len(target.Content); i += 2 {
			if i+1 < len(source.Content) && i+1 < len(target.Content) {
				// Copy key style
				target.Content[i].Style = source.Content[i].Style
				target.Content[i].HeadComment = source.Content[i].HeadComment
				target.Content[i].LineComment = source.Content[i].LineComment
				target.Content[i].FootComment = source.Content[i].FootComment

				// Copy value style
				target.Content[i+1].Style = source.Content[i+1].Style
				target.Content[i+1].HeadComment = source.Content[i+1].HeadComment
				target.Content[i+1].LineComment = source.Content[i+1].LineComment
				target.Content[i+1].FootComment = source.Content[i+1].FootComment

				// Recursively copy styles for nested nodes
				copyNodeStyles(source.Content[i+1], target.Content[i+1])
			}
		}
	}

	// For sequence nodes, copy styles for all elements
	if source.Kind == yaml.SequenceNode && target.Kind == yaml.SequenceNode {
		for i := 0; i < len(source.Content) && i < len(target.Content); i++ {
			target.Content[i].Style = source.Content[i].Style
			target.Content[i].HeadComment = source.Content[i].HeadComment
			target.Content[i].LineComment = source.Content[i].LineComment
			target.Content[i].FootComment = source.Content[i].FootComment

			// Recursively copy styles for nested nodes
			copyNodeStyles(source.Content[i], target.Content[i])
		}
	}
}

// String returns the YAML document as a string
func (d *Document) String() (string, error) {
	bytes, err := d.ToBytes()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
