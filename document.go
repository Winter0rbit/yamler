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
	root                      *yaml.Node
	raw                       string
	arrayRoot                 bool
	trailingNewlines          int
	preserveDocumentSeparator bool // Whether to preserve document separators for array root documents
}

// mappingRoot returns the root MappingNode of the document
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

	// Count trailing newlines
	trailingNewlines := 0
	for i := len(content) - 1; i >= 0; i-- {
		if content[i] == '\n' {
			trailingNewlines++
		} else if content[i] == '\r' {
			// Skip carriage returns
			continue
		} else {
			break
		}
	}

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(content), &node); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	doc := &Document{
		root:             &node,
		raw:              content,
		trailingNewlines: trailingNewlines,
	}

	// Detect if this is an array document root
	if doc.isArrayRoot() {
		doc.arrayRoot = true
	}

	return doc, nil
}

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
	parts := strings.Split(path, ".")
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
	parts := strings.Split(path, ".")
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

	// Always encode with standard 2-space indentation first
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2) // Always use 2 spaces for encoding

	// Preserve original node styles before encoding
	if d.raw != "" {
		preserveNodeStyles(d.root, d.raw)
		// Apply zero-indent arrays formatting to nodes before encoding
		info := detectFormattingInfo(d.raw)
		applyZeroIndentToNodes(d.root, info, "")
	}

	if err := encoder.Encode(d.root); err != nil {
		return nil, err
	}
	encoder.Close()

	result := buf.Bytes()

	// If we have raw content, apply formatting preservation
	if d.raw != "" {
		// Detect original formatting characteristics
		indentInfo := detectFormattingInfo(d.raw)

		// Post-process to maintain original style characteristics
		result = preserveOriginalFormatting(result, d.raw, indentInfo, d.preserveDocumentSeparator)
	}

	// Remove any trailing newlines that might have been added by the encoder
	for len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	// Add the correct number of trailing newlines
	// YAML files should end with at least one newline per convention
	if d.trailingNewlines > 0 {
		result = append(result, bytes.Repeat([]byte("\n"), d.trailingNewlines)...)
	} else {
		// If no trailing newlines were detected, add one (YAML convention)
		result = append(result, '\n')
	}

	return result, nil
}

// FormattingInfo holds information about the original YAML formatting
type FormattingInfo struct {
	IndentSize       int
	UseTabs          bool
	EmptyLines       map[string]int        // Number of empty lines before each key
	FlowStyles       map[string]bool       // Nodes that should remain in flow style
	ScalarStyles     map[string]yaml.Style // Preserve literal/folded scalars
	MultilineFlow    map[string]bool       // Multiline flow objects
	ZeroIndentArrays map[string]bool       // Arrays that start without additional indentation
	HasDocumentStart bool                  // Whether the original had "---"
	HasDocumentEnd   bool                  // Whether the original had "..."
}

// detectFormattingInfo analyzes the raw YAML to determine formatting characteristics
func detectFormattingInfo(raw string) *FormattingInfo {
	lines := strings.Split(raw, "\n")
	info := &FormattingInfo{
		IndentSize:       2,
		UseTabs:          false,
		EmptyLines:       make(map[string]int),
		FlowStyles:       make(map[string]bool),
		ScalarStyles:     make(map[string]yaml.Style),
		MultilineFlow:    make(map[string]bool),
		ZeroIndentArrays: make(map[string]bool),
		HasDocumentStart: false,
		HasDocumentEnd:   false,
	}

	// Collect all indentation levels
	var indentLevels []int
	indentCounts := make(map[int]int)

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Detect indentation type and size
		leadingSpaces := 0
		leadingTabs := 0
		for _, r := range line {
			if r == ' ' {
				leadingSpaces++
			} else if r == '\t' {
				leadingTabs++
				info.UseTabs = true
			} else {
				break
			}
		}

		// Collect indentation levels
		if leadingSpaces > 0 && !info.UseTabs {
			indentLevels = append(indentLevels, leadingSpaces)
			indentCounts[leadingSpaces]++
		} else if leadingTabs > 0 {
			info.IndentSize = 4 // Standard tab equivalent
		}

		trimmed := strings.TrimSpace(line)

		// Detect document separators
		if trimmed == "---" {
			info.HasDocumentStart = true
		}
		if trimmed == "..." {
			info.HasDocumentEnd = true
		}

		// Detect flow styles
		if strings.Contains(trimmed, "{") || strings.Contains(trimmed, "[") {
			if strings.Contains(trimmed, ":") {
				if idx := strings.Index(trimmed, ":"); idx > 0 {
					key := strings.TrimSpace(trimmed[:idx])
					info.FlowStyles[key] = true

					// Check for multiline flow
					if strings.HasSuffix(trimmed, "{") || strings.HasSuffix(trimmed, "[") {
						info.MultilineFlow[key] = true
					}
				}
			}
		}

		// Detect scalar styles (literal | or folded >)
		if strings.Contains(trimmed, "|") || strings.Contains(trimmed, ">") {
			if strings.Contains(trimmed, ":") {
				if idx := strings.Index(trimmed, ":"); idx > 0 {
					key := strings.TrimSpace(trimmed[:idx])
					if strings.Contains(trimmed, "|") {
						info.ScalarStyles[key] = yaml.LiteralStyle
					} else if strings.Contains(trimmed, ">") {
						info.ScalarStyles[key] = yaml.FoldedStyle
					}
				}
			}
		}

		// Detect zero-indent arrays (Kubernetes/GitHub Actions style)
		// Look for lines that are array elements immediately following a key
		if strings.HasPrefix(trimmed, "- ") && i > 0 {
			// Check previous non-empty line for a key
			for j := i - 1; j >= 0; j-- {
				prevLine := strings.TrimSpace(lines[j])
				if prevLine == "" {
					continue // Skip empty lines
				}

				if strings.Contains(prevLine, ":") && !strings.Contains(prevLine, "- ") {
					if idx := strings.Index(prevLine, ":"); idx > 0 {
						key := strings.TrimSpace(prevLine[:idx])
						// Check if the array element has no additional indentation relative to the key
						prevIndent := getLineIndentation(lines[j])
						currentIndent := getLineIndentation(line)
						if currentIndent == prevIndent {
							info.ZeroIndentArrays[key] = true
						}
					}
				}
				break // Only check the immediately previous non-empty line
			}
		}
	}

	// Find the most common indentation increment if not using tabs
	if !info.UseTabs && len(indentLevels) > 0 {
		// Find the GCD of all indentation levels to get the base increment
		baseIndent := findBaseIndentation(indentLevels)
		if baseIndent > 0 {
			info.IndentSize = baseIndent
		}
	}

	// Count empty lines before each key
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && strings.Contains(trimmed, ":") {
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				key := strings.TrimSpace(trimmed[:idx])

				// Count consecutive empty lines before this key
				emptyCount := 0
				for j := i - 1; j >= 0; j-- {
					if strings.TrimSpace(lines[j]) == "" {
						emptyCount++
					} else {
						break
					}
				}

				if emptyCount > 0 {
					info.EmptyLines[key] = emptyCount
				}
			}
		}
	}

	return info
}

// findBaseIndentation finds the greatest common divisor of indentation levels
func findBaseIndentation(levels []int) int {
	if len(levels) == 0 {
		return 2
	}

	// Find minimum non-zero level
	minLevel := levels[0]
	for _, level := range levels {
		if level > 0 && level < minLevel {
			minLevel = level
		}
	}

	// Check if minLevel is a valid base for all levels
	for _, level := range levels {
		if level%minLevel != 0 {
			// Try common indentation sizes
			for _, candidate := range []int{2, 3, 4, 6, 8} {
				valid := true
				for _, l := range levels {
					if l%candidate != 0 {
						valid = false
						break
					}
				}
				if valid {
					return candidate
				}
			}
			// Fallback to 2 if no pattern found
			return 2
		}
	}

	return minLevel
}

// preserveNodeStyles recursively preserves original node styles
func preserveNodeStyles(node *yaml.Node, raw string) {
	if node == nil {
		return
	}

	info := detectFormattingInfo(raw)
	preserveNodeStylesWithInfo(node, info, "")
}

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

// preserveOriginalFormatting applies original formatting characteristics to new content
func preserveOriginalFormatting(newContent []byte, original string, info *FormattingInfo, preserveDocumentSeparator bool) []byte {
	newStr := string(newContent)

	// Convert spaces to tabs if original used tabs
	if info.UseTabs {
		newStr = convertSpacesToTabs(newStr, info)
	} else if info.IndentSize != 2 {
		// Handle custom space indentation (4, 6, 8 spaces, etc.)
		newStr = convertToCustomIndentation(newStr, info.IndentSize)
	}

	// Apply empty line patterns
	newStr = applyEmptyLinePatterns(newStr, info)

	// Preserve multiline flow formatting
	newStr = preserveMultilineFlow(newStr, original, info)

	// Preserve folded scalar formatting
	newStr = preserveFoldedScalars(newStr, original, info)

	// Apply zero-indent array formatting
	newStr = applyZeroIndentArrays(newStr, info)

	// Restore document separators
	newStr = restoreDocumentSeparators(newStr, info, original, preserveDocumentSeparator)

	return []byte(newStr)
}

// convertSpacesToTabs converts spaces to tabs based on indent size
func convertSpacesToTabs(content string, info *FormattingInfo) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "  ") {
			// Convert leading spaces to tabs
			leadingSpaces := 0
			for _, r := range line {
				if r == ' ' {
					leadingSpaces++
				} else {
					break
				}
			}

			if leadingSpaces > 0 {
				tabs := strings.Repeat("\t", leadingSpaces/info.IndentSize)
				remainingSpaces := strings.Repeat(" ", leadingSpaces%info.IndentSize)
				lines[i] = tabs + remainingSpaces + strings.TrimLeft(line, " ")
			}
		}
	}
	return strings.Join(lines, "\n")
}

// convertToCustomIndentation converts 2-space indentation to custom indentation
func convertToCustomIndentation(content string, targetIndentSize int) string {
	if targetIndentSize == 2 {
		return content // No conversion needed
	}

	lines := strings.Split(content, "\n")
	converted := false

	for i, line := range lines {
		if strings.HasPrefix(line, " ") {
			// Count leading spaces
			leadingSpaces := 0
			for _, r := range line {
				if r == ' ' {
					leadingSpaces++
				} else {
					break
				}
			}

			// Only convert if it looks like 2-space indentation (multiples of 2, not already converted)
			if leadingSpaces > 0 && leadingSpaces%2 == 0 && leadingSpaces < targetIndentSize*10 {
				// Convert 2-space levels to target indent size
				indentLevel := leadingSpaces / 2
				newIndent := strings.Repeat(" ", indentLevel*targetIndentSize)
				lines[i] = newIndent + strings.TrimLeft(line, " ")
				converted = true
			}
		}
	}

	// If no conversion happened and we have very large indents, it might already be converted
	if !converted {
		return content
	}

	return strings.Join(lines, "\n")
}

// preserveFoldedScalars preserves original formatting of folded scalars
func preserveFoldedScalars(newContent, original string, info *FormattingInfo) string {
	// Find folded scalars in original content and preserve their exact formatting
	originalLines := strings.Split(original, "\n")

	// Map of key -> original folded content
	foldedContent := make(map[string][]string)

	// Extract original folded scalar content
	for i, line := range originalLines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, ">") && strings.Contains(trimmed, ":") {
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				key := strings.TrimSpace(trimmed[:idx])
				if info.ScalarStyles[key] == yaml.FoldedStyle {
					// Capture the folded content
					var foldedLines []string
					indent := getLineIndentation(line)

					// Start from next line and capture indented content
					for j := i + 1; j < len(originalLines); j++ {
						nextLine := originalLines[j]
						if strings.TrimSpace(nextLine) == "" {
							foldedLines = append(foldedLines, "")
							continue
						}

						nextIndent := getLineIndentation(nextLine)
						if nextIndent > indent {
							// This line belongs to the folded scalar
							foldedLines = append(foldedLines, nextLine)
						} else {
							// End of folded scalar
							break
						}
					}
					foldedContent[key] = foldedLines
				}
			}
		}
	}

	// Replace folded scalars in new content with original formatting
	for key, originalContent := range foldedContent {
		newContent = replaceFoldedScalar(newContent, key, originalContent)
	}

	return newContent
}

// applyZeroIndentArrays applies zero-indent formatting to arrays that should have it
func applyZeroIndentArrays(content string, info *FormattingInfo) string {
	if len(info.ZeroIndentArrays) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for keys that should have zero-indent arrays
		if strings.Contains(trimmed, ":") && !strings.Contains(trimmed, "- ") {
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				key := strings.TrimSpace(trimmed[:idx])

				if info.ZeroIndentArrays[key] {
					// Found a zero-indent array, adjust following array elements
					keyIndent := getLineIndentation(line)

					// Process following lines that are array elements
					for j := i + 1; j < len(lines); j++ {
						nextLine := lines[j]
						nextTrimmed := strings.TrimSpace(nextLine)

						if nextTrimmed == "" {
							continue // Skip empty lines
						}

						if strings.HasPrefix(nextTrimmed, "- ") {
							// This is an array element
							nextIndent := getLineIndentation(nextLine)

							// If it has extra indentation, remove it to match key level
							if nextIndent > keyIndent {
								// Remove extra indentation to match key level
								newIndent := strings.Repeat(" ", keyIndent)
								lines[j] = newIndent + nextTrimmed
							}
						} else {
							// Non-array element, check if it belongs to the array element
							nextIndent := getLineIndentation(nextLine)
							if nextIndent > keyIndent {
								// This might be a nested element of the array item
								// Adjust its indentation relative to the array element
								baseArrayIndent := keyIndent
								expectedElementIndent := baseArrayIndent + info.IndentSize
								if nextIndent > expectedElementIndent {
									// Reduce indentation
									reduction := info.IndentSize
									newIndent := nextIndent - reduction
									if newIndent < expectedElementIndent {
										newIndent = expectedElementIndent
									}
									lines[j] = strings.Repeat(" ", newIndent) + nextTrimmed
								}
							} else {
								// Not part of this array anymore
								break
							}
						}
					}
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// restoreDocumentSeparators adds back document separators if they were in the original
func restoreDocumentSeparators(content string, info *FormattingInfo, originalContent string, preserveDocumentSeparator bool) string {
	// Check if the original content actually starts with ---
	originallyHadDocumentStart := strings.HasPrefix(strings.TrimSpace(originalContent), "---")
	originallyHadDocumentEnd := strings.HasSuffix(strings.TrimSpace(originalContent), "...")

	// Don't add separators if preservation is disabled or they weren't in original
	if !preserveDocumentSeparator || (!originallyHadDocumentStart && !originallyHadDocumentEnd) {
		return content
	}

	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines)+2)

	// Add document start separator only if original had it and preservation is enabled
	if originallyHadDocumentStart {
		result = append(result, "---")
	}

	// Add content, but remove trailing empty lines if we're adding document end separator
	if originallyHadDocumentEnd {
		// Remove trailing empty lines
		for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
			lines = lines[:len(lines)-1]
		}
	}
	result = append(result, lines...)

	// Add document end separator only if original had it and preservation is enabled
	if originallyHadDocumentEnd {
		result = append(result, "...")
	}

	return strings.Join(result, "\n")
}

// getLineIndentation returns the number of leading spaces in a line
func getLineIndentation(line string) int {
	count := 0
	for _, r := range line {
		if r == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}

// replaceFoldedScalar replaces folded scalar content in new YAML with original formatting
func replaceFoldedScalar(content, key string, originalLines []string) string {
	lines := strings.Split(content, "\n")

	// Find the key line in new content
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, ">") && strings.HasPrefix(trimmed, key+":") {
			// Found the folded scalar key, replace subsequent content
			indent := getLineIndentation(line)

			// Remove old folded content
			endIdx := i + 1
			for j := i + 1; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) == "" {
					endIdx = j + 1
					continue
				}
				lineIndent := getLineIndentation(lines[j])
				if lineIndent > indent {
					endIdx = j + 1
				} else {
					break
				}
			}

			// Insert original content
			newLines := make([]string, 0, len(lines)-endIdx+i+1+len(originalLines))
			newLines = append(newLines, lines[:i+1]...)
			newLines = append(newLines, originalLines...)
			newLines = append(newLines, lines[endIdx:]...)

			return strings.Join(newLines, "\n")
		}
	}

	return content
}

// applyEmptyLinePatterns adds empty lines before specified keys
func applyEmptyLinePatterns(content string, info *FormattingInfo) string {
	lines := strings.Split(content, "\n")
	var result []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && strings.Contains(trimmed, ":") {
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				key := strings.TrimSpace(trimmed[:idx])
				emptyLinesCount := info.EmptyLines[key]
				if emptyLinesCount > 0 && i > 0 && strings.TrimSpace(lines[i-1]) != "" {
					// Add the specified number of empty lines
					for j := 0; j < emptyLinesCount; j++ {
						result = append(result, "")
					}
				}
			}
		}
		result = append(result, line)
	}

	return strings.Join(result, "\n")
}

// preserveMultilineFlow preserves multiline flow object formatting
func preserveMultilineFlow(newContent, original string, info *FormattingInfo) string {
	// Extract multiline flow objects from original and restore them in new content
	originalLines := strings.Split(original, "\n")

	// Map of key -> original multiline flow content
	multilineFlowContent := make(map[string][]string)

	// Find multiline flow objects in original
	for i, line := range originalLines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, ":") {
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				key := strings.TrimSpace(trimmed[:idx])

				// Check if this key has multiline flow formatting
				if info.MultilineFlow[key] {
					// Capture the multiline flow content
					var flowLines []string
					indent := getLineIndentation(line)

					// Add the key line itself
					flowLines = append(flowLines, line)

					// Capture subsequent lines that belong to this flow object
					for j := i + 1; j < len(originalLines); j++ {
						nextLine := originalLines[j]
						nextTrimmed := strings.TrimSpace(nextLine)

						if nextTrimmed == "" {
							flowLines = append(flowLines, nextLine)
							continue
						}

						nextIndent := getLineIndentation(nextLine)

						// If it's more indented or ends the flow object, include it
						if nextIndent > indent || strings.HasSuffix(nextTrimmed, "}") || strings.HasSuffix(nextTrimmed, "]") {
							flowLines = append(flowLines, nextLine)

							// If it ends the flow object, stop
							if strings.HasSuffix(nextTrimmed, "}") || strings.HasSuffix(nextTrimmed, "]") {
								break
							}
						} else {
							// End of flow object
							break
						}
					}

					multilineFlowContent[key] = flowLines
				}
			}
		}
	}

	// Replace multiline flow objects in new content with original formatting
	for key, originalFlowLines := range multilineFlowContent {
		newContent = replaceMultilineFlow(newContent, key, originalFlowLines)
	}

	return newContent
}

// replaceMultilineFlow replaces multiline flow object in new content with original formatting
func replaceMultilineFlow(content, key string, originalFlowLines []string) string {
	lines := strings.Split(content, "\n")

	// Find the key line in new content
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, ":") && strings.HasPrefix(trimmed, key+":") {
			// Check if this became a single-line flow object
			if strings.Contains(trimmed, "{") && strings.Contains(trimmed, "}") {
				// Replace single line with multiline original
				newLines := make([]string, 0, len(lines)-1+len(originalFlowLines))
				newLines = append(newLines, lines[:i]...)
				newLines = append(newLines, originalFlowLines...)
				newLines = append(newLines, lines[i+1:]...)

				return strings.Join(newLines, "\n")
			}
		}
	}

	return content
}

// detectIndentation analyzes the raw YAML to determine the indentation level
func detectIndentation(raw string) int {
	info := detectFormattingInfo(raw)
	return info.IndentSize
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
