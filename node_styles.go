// Adjustments applied to the yaml.v3 node tree before encoding so that the
// encoder itself reproduces as much of the original style as possible.

package yamler

import (
	"gopkg.in/yaml.v3"
)

// preserveNodeStylesWithInfo recursively preserves node styles using formatting info
func preserveNodeStylesWithInfo(node *yaml.Node, info *FormattingInfo, path string) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		// Check if this mapping should be in flow style
		if info.FlowStyles[path] {
			if info.MultilineFlow[path] {
				// Keep multiline flow formatting
				node.Style = yaml.FlowStyle
			} else {
				node.Style = yaml.FlowStyle
			}
		}

		// Process children
		for i := 0; i < len(node.Content); i += 2 {
			if i+1 < len(node.Content) {
				key := node.Content[i].Value
				newPath := path
				if newPath == "" {
					newPath = key
				} else {
					newPath = path + "." + key
				}

				// Preserve scalar styles
				if style, exists := info.ScalarStyles[key]; exists {
					node.Content[i+1].Style = style
				}

				preserveNodeStylesWithInfo(node.Content[i+1], info, newPath)
			}
		}

	case yaml.SequenceNode:
		// Check if sequence should be in flow style
		if info.FlowStyles[path] {
			node.Style = yaml.FlowStyle
		}

		// Process children
		for _, child := range node.Content {
			preserveNodeStylesWithInfo(child, info, path)
		}
	}
}

// applyZeroIndentToNodes applies zero-indent formatting to nodes before encoding
func applyZeroIndentToNodes(node *yaml.Node, info *FormattingInfo, path string) {
	if node == nil || len(info.ZeroIndentArrays) == 0 {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		// Process document children
		for _, child := range node.Content {
			applyZeroIndentToNodes(child, info, path)
		}

	case yaml.MappingNode:
		// Process mapping children
		for i := 0; i < len(node.Content); i += 2 {
			if i+1 < len(node.Content) {
				key := node.Content[i].Value
				newPath := path
				if newPath == "" {
					newPath = key
				} else {
					newPath = path + "." + key
				}

				// Check if this key should have zero-indent arrays
				if info.ZeroIndentArrays[key] && node.Content[i+1].Kind == yaml.SequenceNode {
					// Mark this sequence for special indentation handling
					// We'll use a custom tag to identify it during post-processing
					node.Content[i+1].Tag = "!!seq"
					node.Content[i+1].Style = 0 // Block style
				}

				applyZeroIndentToNodes(node.Content[i+1], info, newPath)
			}
		}

	case yaml.SequenceNode:
		// Process sequence children
		for _, child := range node.Content {
			applyZeroIndentToNodes(child, info, path)
		}
	}
}
