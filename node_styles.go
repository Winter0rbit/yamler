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

// stripMergeTags clears the explicit "!!merge" tag that the yaml.v3 decoder
// attaches to "<<" keys. The v3.0.1 encoder would otherwise render them as
// "!!merge <<: *anchor" instead of the plain "<<: *anchor" form.
func stripMergeTags(node *yaml.Node) {
	if node == nil {
		return
	}
	if node.Kind == yaml.ScalarNode && node.Value == "<<" && node.Tag == "!!merge" {
		node.Tag = ""
	}
	for _, child := range node.Content {
		stripMergeTags(child)
	}
}
