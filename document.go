package yamler

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Pool for reusing byte buffers to reduce allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// Pool for reusing string builders
var stringBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// Cache for parsed paths to avoid repeated string splitting
var pathCache = sync.Map{} // string -> []string

// parsePath splits a path and caches the result
func parsePath(path string) []string {
	if cached, ok := pathCache.Load(path); ok {
		return cached.([]string)
	}

	parts := strings.Split(path, ".")
	pathCache.Store(path, parts)
	return parts
}

// Document represents a YAML document with preserved formatting
type Document struct {
	root                      *yaml.Node
	raw                       string
	arrayRoot                 bool
	trailingNewlines          int
	preserveDocumentSeparator bool // Whether to preserve document separators for array root documents
	// Performance optimization: cache formatting info
	formattingCache *FormattingInfo
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

	// Initialize formatting cache if we have raw content
	if content != "" {
		doc.formattingCache = detectFormattingInfoOptimized(content)
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

	// Get buffer from pool to reduce allocations
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2) // Always use 2 spaces for encoding

	// Preserve original node styles before encoding
	if d.raw != "" {
		// Use cached formatting info or detect if not cached
		var info *FormattingInfo
		if d.formattingCache != nil {
			info = d.formattingCache
		} else {
			info = detectFormattingInfoOptimized(d.raw)
			d.formattingCache = info // Cache for future use
		}

		preserveNodeStylesWithInfo(d.root, info, "")
		// Apply zero-indent arrays formatting to nodes before encoding
		applyZeroIndentToNodes(d.root, info, "")
	}

	if err := encoder.Encode(d.root); err != nil {
		return nil, err
	}
	encoder.Close()

	// Make a copy of the buffer contents
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	// If we have raw content, apply formatting preservation
	if d.raw != "" {
		// Use cached formatting info (already computed above)
		var indentInfo *FormattingInfo
		if d.formattingCache != nil {
			indentInfo = d.formattingCache
		} else {
			indentInfo = detectFormattingInfoOptimized(d.raw)
			d.formattingCache = indentInfo // Cache for future use
		}

		// Post-process to maintain original style characteristics
		result = preserveOriginalFormatting(result, d.raw, indentInfo, d.preserveDocumentSeparator)
	}

	// Remove any trailing newlines that might have been added by the encoder
	for len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	// Add the correct number of trailing newlines (preserve original exactly)
	if d.trailingNewlines > 0 {
		// Pre-allocate with exact size needed
		finalResult := make([]byte, len(result)+d.trailingNewlines)
		copy(finalResult, result)
		for i := len(result); i < len(finalResult); i++ {
			finalResult[i] = '\n'
		}
		return finalResult, nil
	} else {
		// If no trailing newlines were detected, don't add any (preserve original)
		return result, nil
	}
}

// CommentAlignmentMode defines how inline comments should be aligned
type CommentAlignmentMode int

const (
	// CommentAlignmentRelative preserves original spacing between value and comment
	CommentAlignmentRelative CommentAlignmentMode = iota
	// CommentAlignmentAbsolute aligns all comments to the same column
	CommentAlignmentAbsolute
	// CommentAlignmentDisabled disables comment alignment processing
	CommentAlignmentDisabled
)

// ArrayStyle represents different array formatting styles
type ArrayStyle struct {
	IsFlow      bool // true for [1,2,3], false for block style
	IsMultiline bool // true for multiline flow arrays
	HasSpaces   bool // true for [ 1 , 2 , 3 ] (spaces around elements)
	IsCompact   bool // true for [1,2,3] (no spaces)
	Indentation int  // custom indentation level
}

// FormattingInfo holds information about the original YAML formatting
type FormattingInfo struct {
	IndentSize       int
	UseTabs          bool
	EmptyLines       map[string]int         // Number of empty lines before each key
	FlowStyles       map[string]bool        // Nodes that should remain in flow style
	ScalarStyles     map[string]yaml.Style  // Preserve literal/folded scalars
	MultilineFlow    map[string]bool        // Multiline flow objects
	ZeroIndentArrays map[string]bool        // Arrays that start without additional indentation
	HasDocumentStart bool                   // Whether the original had "---"
	HasDocumentEnd   bool                   // Whether the original had "..."
	CommentAlignment map[string]int         // Spacing or column position for inline comments
	CommentSpacing   int                    // Common spacing for comment alignment
	AlignmentMode    CommentAlignmentMode   // How to align comments
	ArrayStyles      map[string]*ArrayStyle // Array formatting styles
	KeyIndents       map[string]int         // Exact indentation for each key
	FlowObjectStyles map[string]string      // Original flow object strings to preserve exact formatting
}

// detectFormattingInfoOptimized is an optimized version with fewer allocations
func detectFormattingInfoOptimized(raw string) *FormattingInfo {
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
		CommentAlignment: make(map[string]int),
		CommentSpacing:   0,
		AlignmentMode:    CommentAlignmentRelative, // Default to relative alignment
		ArrayStyles:      make(map[string]*ArrayStyle),
		KeyIndents:       make(map[string]int),
		FlowObjectStyles: make(map[string]string),
	}

	// Pre-allocate slices with reasonable capacity
	indentLevels := make([]int, 0, 32)

	// Process lines to detect formatting patterns
	lines := strings.Split(raw, "\n")
	emptyLineCount := 0

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			emptyLineCount++
		} else {
			processLineOptimized(line, i, emptyLineCount, info, &indentLevels)
			emptyLineCount = 0
		}
	}

	// Find the most common indentation increment if not using tabs
	if !info.UseTabs && len(indentLevels) > 0 {
		baseIndent := findBaseIndentationOptimized(indentLevels)
		if baseIndent > 0 {
			info.IndentSize = baseIndent
		}
	}

	// Calculate common comment alignment
	if len(info.CommentAlignment) > 0 {
		info.CommentSpacing = findCommonCommentAlignment(info.CommentAlignment)
	}

	// Filter KeyIndents to keep only non-standard indentations
	filteredKeyIndents := make(map[string]int)
	for key, indent := range info.KeyIndents {
		// Check if this indent is a standard multiple of IndentSize
		if indent == 0 {
			// Root level keys - only keep if they have non-zero indent (custom)
			continue
		} else if info.IndentSize > 0 && indent%info.IndentSize == 0 {
			// Standard indentation - but check if it's actually expected for this nesting level
			expectedLevel := indent / info.IndentSize

			// Only filter out if it's a reasonable and expected standard indentation
			// For custom indentations like 6 spaces at level 1, we should keep it
			if expectedLevel == 1 && indent == info.IndentSize {
				// Standard first level indentation (2, 4, etc.) - filter out
				continue
			} else if expectedLevel == 2 && indent == info.IndentSize*2 {
				// Standard second level indentation (4, 8, etc.) - filter out
				continue
			}
			// For all other cases (like 6 spaces), keep as custom indentation
		}
		// Keep custom indentations
		filteredKeyIndents[key] = indent
	}
	info.KeyIndents = filteredKeyIndents

	// Detect multiline flow objects after processing all lines
	detectMultilineFlowObjects(lines, info)

	return info
}

// detectMultilineFlowObjects finds and stores multiline flow objects like:
//
//	resources: {
//	  cpu: 256,
//	  memory: 256}
func detectMultilineFlowObjects(lines []string, info *FormattingInfo) {
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Look for lines that end with { or [
		if strings.Contains(trimmed, ":") && (strings.HasSuffix(trimmed, "{") || strings.HasSuffix(trimmed, "[")) {
			colonPos := strings.Index(trimmed, ":")
			if colonPos > 0 {
				key := strings.TrimSpace(trimmed[:colonPos])
				value := strings.TrimSpace(trimmed[colonPos+1:])

				// Check if this starts a multiline flow object
				if strings.HasSuffix(value, "{") && !strings.Contains(value, "}") {
					// Found start of multiline flow object, now find the end
					flowObject := collectMultilineFlowObject(lines, i, '{', '}')
					if flowObject != "" {
						info.FlowObjectStyles[key] = flowObject
					}
				} else if strings.HasSuffix(value, "[") && !strings.Contains(value, "]") {
					// Found start of multiline flow array, now find the end
					flowObject := collectMultilineFlowObject(lines, i, '[', ']')
					if flowObject != "" {
						info.FlowObjectStyles[key] = flowObject
					}
				}
			}
		}
	}
}

// collectMultilineFlowObject collects a complete multiline flow object starting from startLine
func collectMultilineFlowObject(lines []string, startLine int, openBrace, closeBrace rune) string {
	var result strings.Builder
	depth := 0

	for i := startLine; i < len(lines); i++ {
		line := lines[i]

		if i == startLine {
			// For the first line, only take the value part after the colon
			if colonPos := strings.Index(line, ":"); colonPos >= 0 {
				value := strings.TrimSpace(line[colonPos+1:])
				result.WriteString(value)
			}
		} else {
			// For subsequent lines, take the trimmed content
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				result.WriteString("\n")
				result.WriteString(line)
			}
		}

		// Count braces AFTER adding the line to result
		for _, r := range line {
			if r == openBrace {
				depth++
			} else if r == closeBrace {
				depth--
				// Check if we've closed all braces after processing this character
				if depth == 0 {
					return result.String()
				}
			}
		}
	}

	// If we get here, we didn't find a complete object
	return ""
}

// processLineOptimized processes a single line efficiently
func processLineOptimized(line string, lineNum, emptyLinesBefore int, info *FormattingInfo, indentLevels *[]int) {
	if len(line) == 0 {
		return
	}

	// Count leading whitespace in one pass
	leadingSpaces := 0
	leadingTabs := 0
	contentStart := 0

	for i, r := range line {
		if r == ' ' {
			leadingSpaces++
		} else if r == '\t' {
			leadingTabs++
			info.UseTabs = true
		} else {
			contentStart = i
			break
		}
	}

	// Skip empty lines
	if contentStart >= len(line) {
		return
	}

	content := line[contentStart:]

	// Collect indentation levels
	if leadingSpaces > 0 && !info.UseTabs {
		*indentLevels = append(*indentLevels, leadingSpaces)
	} else if leadingTabs > 0 {
		info.IndentSize = 4
	}

	// Quick checks for common patterns
	if len(content) >= 3 {
		if content == "---" {
			info.HasDocumentStart = true
			return
		}
		if content == "..." {
			info.HasDocumentEnd = true
			return
		}
	}

	// Handle standalone comments (like "# Application settings")
	if strings.HasPrefix(content, "#") {
		// Use the comment text as a key for empty line tracking
		commentKey := strings.TrimSpace(content)
		if emptyLinesBefore > 0 {
			info.EmptyLines[commentKey] = emptyLinesBefore
		}
		return
	}

	// Handle array elements (lines starting with "- ")
	if strings.HasPrefix(content, "- ") {
		// Store indentation for array elements using a special key format
		arrayElementKey := fmt.Sprintf("__array_element_%d__", leadingSpaces)
		info.KeyIndents[arrayElementKey] = leadingSpaces
		return
	}

	// Find colon position efficiently
	colonPos := -1
	for i, r := range content {
		if r == ':' {
			colonPos = i
			break
		}
	}

	if colonPos <= 0 {
		return
	}

	// Extract key efficiently
	key := strings.TrimSpace(content[:colonPos])
	if key == "" {
		return
	}

	// Store empty lines count
	if emptyLinesBefore > 0 {
		info.EmptyLines[key] = emptyLinesBefore
	}

	// Store exact indentation for this key only if it's non-standard
	// We need to check against the detected indent size after processing all lines
	// For now, store all indents and filter later
	info.KeyIndents[key] = leadingSpaces

	// Check for flow styles, scalar styles, and comments in one pass
	valueStart := colonPos + 1
	if valueStart < len(content) {
		value := content[valueStart:]

		// Check for inline comments
		if commentPos := strings.Index(value, "#"); commentPos >= 0 {
			// For relative alignment, calculate spacing between value and comment
			valueBeforeComment := value[:commentPos]
			// Count trailing spaces in the value part
			spacesBeforeComment := len(valueBeforeComment) - len(strings.TrimRight(valueBeforeComment, " "))
			if spacesBeforeComment > 0 {
				info.CommentAlignment[key] = spacesBeforeComment
			}
		}

		// Check for flow styles
		if strings.ContainsAny(value, "{[") {
			// Mark as FlowStyles regardless of indentation level to preserve nested flow objects
			info.FlowStyles[key] = true

			// Store the original flow object string to preserve exact formatting
			trimmedValue := strings.TrimSpace(value)

			// Check if this is a complete single-line flow object
			if (strings.Contains(trimmedValue, "{") && strings.Contains(trimmedValue, "}")) ||
				(strings.Contains(trimmedValue, "[") && strings.Contains(trimmedValue, "]")) {
				// This is a single-line flow object, save the exact format
				info.FlowObjectStyles[key] = trimmedValue
			}

			// Only mark as MultilineFlow if the line actually ends with { or [
			// AND doesn't contain the closing bracket/brace on the same line
			if strings.HasSuffix(trimmedValue, "{") && !strings.Contains(trimmedValue, "}") {
				info.MultilineFlow[key] = true
			} else if strings.HasSuffix(trimmedValue, "[") && !strings.Contains(trimmedValue, "]") {
				info.MultilineFlow[key] = true
			}
		}

		// Detect array styles
		if strings.Contains(value, "[") {
			// This is a flow array, analyze its style
			arrayStyle := &ArrayStyle{
				IsFlow:      true,
				IsMultiline: false,
				Indentation: leadingSpaces,
			}

			trimmedValue := strings.TrimSpace(value)
			if strings.HasPrefix(trimmedValue, "[") && strings.HasSuffix(trimmedValue, "]") {
				// Single line flow array
				arrayContent := trimmedValue[1 : len(trimmedValue)-1]

				// Check for spaces around elements
				if strings.Contains(arrayContent, " , ") ||
					(strings.HasPrefix(arrayContent, " ") && strings.HasSuffix(arrayContent, " ")) {
					arrayStyle.HasSpaces = true
				} else if !strings.Contains(arrayContent, " ") {
					arrayStyle.IsCompact = true
				}
			} else if strings.HasSuffix(trimmedValue, "[") {
				// Multiline flow array
				arrayStyle.IsMultiline = true
				info.MultilineFlow[key] = true
			}

			info.ArrayStyles[key] = arrayStyle
		}

		// Check for scalar styles
		trimmedValue := strings.TrimSpace(value)
		if len(trimmedValue) > 0 {
			switch trimmedValue[0] {
			case '|':
				info.ScalarStyles[key] = yaml.LiteralStyle
			case '>':
				info.ScalarStyles[key] = yaml.FoldedStyle
			}
		}
	}
}

// findBaseIndentationOptimized finds the most appropriate base indentation
func findBaseIndentationOptimized(levels []int) int {
	if len(levels) == 0 {
		return 2
	}

	// Filter out zero levels and find the minimum non-zero level
	nonZeroLevels := make([]int, 0, len(levels))
	for _, level := range levels {
		if level > 0 {
			nonZeroLevels = append(nonZeroLevels, level)
		}
	}

	if len(nonZeroLevels) == 0 {
		return 2
	}

	// Find the minimum indentation level (this is often the base level)
	minLevel := nonZeroLevels[0]
	for _, level := range nonZeroLevels {
		if level < minLevel {
			minLevel = level
		}
	}

	// If minimum level is reasonable (2, 4, 6, 8), use it directly
	if minLevel >= 2 && minLevel <= 8 {
		// Check if this level works well with other levels
		consistentLevels := 0
		totalLevels := len(nonZeroLevels)

		for _, level := range nonZeroLevels {
			if level%minLevel == 0 {
				consistentLevels++
			}
		}

		// If most levels are multiples of minLevel, use it
		if float64(consistentLevels)/float64(totalLevels) >= 0.7 {
			return minLevel
		}
	}

	// Fallback to GCD approach for complex cases
	result := nonZeroLevels[0]
	for i := 1; i < len(nonZeroLevels); i++ {
		result = gcd(result, nonZeroLevels[i])
		if result == 1 {
			break
		}
	}

	// Ensure result is reasonable (between 1 and 8)
	if result < 1 {
		result = 2
	} else if result > 8 {
		result = 8
	}

	return result
}

// gcd calculates the greatest common divisor
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// findCommonCommentAlignment finds the most common comment alignment column
func findCommonCommentAlignment(alignments map[string]int) int {
	if len(alignments) == 0 {
		return 0
	}

	// Count frequency of each alignment position
	counts := make(map[int]int)
	for _, pos := range alignments {
		counts[pos]++
	}

	// Find the most common alignment
	maxCount := 0
	commonAlignment := 0
	for pos, count := range counts {
		if count > maxCount {
			maxCount = count
			commonAlignment = pos
		}
	}

	return commonAlignment
}

// preserveNodeStyles recursively preserves original node styles
func preserveNodeStyles(node *yaml.Node, raw string) {
	if node == nil {
		return
	}

	info := detectFormattingInfoOptimized(raw)
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
	} else if info.IndentSize != 2 && len(info.KeyIndents) == 0 {
		// Handle custom space indentation (4, 6, 8 spaces, etc.) only if we don't have exact indents
		newStr = convertToCustomIndentation(newStr, info.IndentSize)
	}

	// Preserve multiline flow formatting
	// TODO: Fix path resolution issue - temporarily disabled
	// newStr = preserveMultilineFlow(newStr, original, info)

	// Apply array styles
	newStr = applyArrayStyles(newStr, info)

	// Apply flow object styles to preserve spacing
	newStr = applyFlowObjectStyles(newStr, info)

	// Apply exact key indentations
	newStr = applyExactIndentations(newStr, info)

	// Apply empty line patterns (after indentation to avoid conflicts)
	newStr = applyEmptyLinePatterns(newStr, info)

	// Preserve folded scalar formatting
	newStr = preserveFoldedScalars(newStr, original, info)

	// Apply zero-indent array formatting
	newStr = applyZeroIndentArrays(newStr, info)

	// Align inline comments
	newStr = alignInlineComments(newStr, info)

	// Restore document separators
	newStr = restoreDocumentSeparators(newStr, info, original, preserveDocumentSeparator)

	// Final cleanup: ensure empty lines are truly empty (no indentation)
	// but preserve original empty lines with indentation
	newStr = cleanupEmptyLines(newStr, original)

	return []byte(newStr)
}

// cleanupEmptyLines removes indentation from empty lines, but preserves original empty lines with indentation
func cleanupEmptyLines(content, original string) string {
	lines := strings.Split(content, "\n")
	originalLines := strings.Split(original, "\n")

	// Create a map of original empty lines with their indentation
	originalEmptyLines := make(map[int]string)
	for i, line := range originalLines {
		if strings.TrimSpace(line) == "" && line != "" {
			// This is an original empty line with indentation
			originalEmptyLines[i] = line
		}
	}

	for i, line := range lines {
		// If line contains only whitespace
		if strings.TrimSpace(line) == "" {
			// Check if this corresponds to an original empty line with indentation
			if originalLine, exists := originalEmptyLines[i]; exists {
				// Preserve the original indentation
				lines[i] = originalLine
			} else {
				// Make it truly empty (this was likely added by applyEmptyLinePatterns)
				lines[i] = ""
			}
		}
	}
	return strings.Join(lines, "\n")
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

// alignInlineComments aligns inline comments according to the specified mode
func alignInlineComments(content string, info *FormattingInfo) string {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments-only lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Look for lines with inline comments
		if commentPos := strings.Index(line, "#"); commentPos >= 0 {
			// Extract the part before the comment
			beforeComment := line[:commentPos]
			comment := line[commentPos:]

			// Check if this line has a key that should be aligned
			if colonPos := strings.Index(beforeComment, ":"); colonPos >= 0 {
				key := strings.TrimSpace(beforeComment[:colonPos])

				switch info.AlignmentMode {
				case CommentAlignmentDisabled:
					// Remove comments entirely
					lines[i] = strings.TrimRight(beforeComment, " ")

				case CommentAlignmentRelative:
					// Relative alignment: preserve original spacing
					if alignmentValue, exists := info.CommentAlignment[key]; exists && alignmentValue > 0 {
						// Extract the value part after colon
						valueStart := colonPos + 1
						if valueStart < len(beforeComment) {
							valuePart := beforeComment[valueStart:]
							trimmedValue := strings.TrimSpace(valuePart)

							// Reconstruct with exact spacing
							keyPart := beforeComment[:colonPos] // Just up to colon
							if len(trimmedValue) > 0 {
								padding := strings.Repeat(" ", alignmentValue)
								lines[i] = keyPart + ": " + trimmedValue + padding + comment
							}
						}
					}

				case CommentAlignmentAbsolute:
					// Absolute alignment: align to specific column
					targetColumn := info.CommentSpacing
					if targetColumn > 0 {
						// Remove trailing spaces from before comment
						beforeComment = strings.TrimRight(beforeComment, " ")

						// Add spaces to reach target column
						spacesNeeded := targetColumn - len(beforeComment)
						if spacesNeeded > 0 {
							padding := strings.Repeat(" ", spacesNeeded)
							lines[i] = beforeComment + padding + comment
						} else {
							// If we can't fit, use at least one space
							lines[i] = beforeComment + " " + comment
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

// applyEmptyLinePatterns adds empty lines before specified keys and comments
func applyEmptyLinePatterns(content string, info *FormattingInfo) string {
	lines := strings.Split(content, "\n")
	var result []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		var key string

		// Handle both key-value pairs and standalone comments
		if trimmed != "" {
			if strings.HasPrefix(trimmed, "#") {
				// Standalone comment - use trimmed version for matching
				key = trimmed
			} else if strings.Contains(trimmed, ":") {
				// Key-value pair
				if idx := strings.Index(trimmed, ":"); idx > 0 {
					key = strings.TrimSpace(trimmed[:idx])
				}
			}

			// Apply empty lines if needed
			if key != "" {
				emptyLinesCount := info.EmptyLines[key]
				if emptyLinesCount > 0 && i > 0 && strings.TrimSpace(lines[i-1]) != "" {
					// Add the specified number of empty lines (truly empty, no indentation)
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
// isInlineObject checks if a multiline flow object is actually an inline object
func isInlineObject(lines []string, startIndex int) bool {
	if startIndex >= len(lines) {
		return false
	}

	line := lines[startIndex]
	trimmed := strings.TrimSpace(line)

	// Count opening and closing brackets/braces in the entire object
	openBraces := strings.Count(trimmed, "{")
	closeBraces := strings.Count(trimmed, "}")
	openBrackets := strings.Count(trimmed, "[")
	closeBrackets := strings.Count(trimmed, "]")

	// Check a few lines ahead to see if this is a compact inline object
	maxLinesAhead := 10
	for i := startIndex + 1; i < len(lines) && i < startIndex+maxLinesAhead; i++ {
		nextLine := strings.TrimSpace(lines[i])
		if nextLine == "" {
			continue
		}

		openBraces += strings.Count(nextLine, "{")
		closeBraces += strings.Count(nextLine, "}")
		openBrackets += strings.Count(nextLine, "[")
		closeBrackets += strings.Count(nextLine, "]")

		// If we've balanced all brackets/braces within a few lines, it's likely inline
		if openBraces == closeBraces && openBrackets == closeBrackets {
			// Additional check: if the total line count is small (< 6 lines), it's inline
			return (i - startIndex) < 6
		}

		// If we encounter a line that starts a new key at the same level (not indented more than the original), stop
		nextIndent := len(nextLine) - len(strings.TrimLeft(nextLine, " \t"))
		originalIndent := len(line) - len(strings.TrimLeft(line, " \t"))
		if strings.Contains(nextLine, ":") && nextIndent <= originalIndent {
			break
		}
	}

	return false
}

func preserveMultilineFlow(newContent, original string, info *FormattingInfo) string {
	if len(info.MultilineFlow) == 0 {
		return newContent
	}

	// Simple approach: extract original multiline flow blocks and replace them
	// only if the content semantically matches but formatting differs
	originalLines := strings.Split(original, "\n")

	// Map of key -> original multiline flow block
	originalFlowBlocks := extractMultilineFlowBlocks(originalLines, info)

	// Process each multiline flow key
	for key, originalBlock := range originalFlowBlocks {
		newContent = replaceMultilineFlowBlock(newContent, key, originalBlock)
	}

	return newContent
}

// extractMultilineFlowBlocks extracts multiline flow blocks from original content
func extractMultilineFlowBlocks(lines []string, info *FormattingInfo) map[string][]string {
	blocks := make(map[string][]string)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, ":") {
			continue
		}

		colonIdx := strings.Index(trimmed, ":")
		if colonIdx <= 0 {
			continue
		}

		key := strings.TrimSpace(trimmed[:colonIdx])

		// Only process keys marked as multiline flow
		if !info.MultilineFlow[key] {
			continue
		}

		// Extract the complete multiline flow block
		block := extractFlowBlock(lines, i)
		if len(block) > 1 { // Only multiline blocks
			blocks[key] = block
		}
	}

	return blocks
}

// extractFlowBlock extracts a complete flow block starting from the given line
func extractFlowBlock(lines []string, startIdx int) []string {
	if startIdx >= len(lines) {
		return nil
	}

	var block []string
	line := lines[startIdx]
	trimmed := strings.TrimSpace(line)
	baseIndent := getLineIndentation(line)

	// Add the key line
	block = append(block, line)

	// Count brackets/braces to track flow object boundaries
	bracketCount := 0
	braceCount := 0

	// Count opening brackets/braces in the key line
	for _, r := range trimmed {
		switch r {
		case '[':
			bracketCount++
		case ']':
			bracketCount--
		case '{':
			braceCount++
		case '}':
			braceCount--
		}
	}

	// If the key line already closes the flow object, it's a single line
	if bracketCount == 0 && braceCount == 0 {
		return block
	}

	// Continue reading lines until we close all brackets/braces
	for i := startIdx + 1; i < len(lines); i++ {
		currentLine := lines[i]
		currentTrimmed := strings.TrimSpace(currentLine)
		currentIndent := getLineIndentation(currentLine)

		// Always include empty lines
		if currentTrimmed == "" {
			block = append(block, currentLine)
			continue
		}

		// Check if we should include this line BEFORE counting brackets/braces
		shouldInclude := currentIndent > baseIndent

		// Count brackets/braces in current line
		for _, r := range currentTrimmed {
			switch r {
			case '[':
				bracketCount++
			case ']':
				bracketCount--
				shouldInclude = true // Always include closing brackets
			case '{':
				braceCount++
			case '}':
				braceCount--
				shouldInclude = true // Always include closing braces
			}
		}

		// Include the line if it's indented or contains closing brackets/braces
		// or if we still have open brackets/braces
		if shouldInclude || bracketCount > 0 || braceCount > 0 {
			block = append(block, currentLine)
		} else {
			// We've reached a line that doesn't belong to the flow object
			break
		}

		// If we've closed all brackets/braces, we're done
		if bracketCount == 0 && braceCount == 0 {
			break
		}
	}

	return block
}

// replaceMultilineFlowBlock replaces a multiline flow block in new content with original formatting
func replaceMultilineFlowBlock(content, key string, originalBlock []string) string {
	if len(originalBlock) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")

	// Track document structure to build full paths for keys
	pathStack := make([]string, 0)
	indentStack := make([]int, 0)

	// Find the key in new content
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, ":") {
			continue
		}

		colonIdx := strings.Index(trimmed, ":")
		if colonIdx <= 0 {
			continue
		}

		lineKey := strings.TrimSpace(trimmed[:colonIdx])
		currentIndent := getLineIndentation(line)

		// Update path stack based on indentation
		for len(indentStack) > 0 && indentStack[len(indentStack)-1] >= currentIndent {
			pathStack = pathStack[:len(pathStack)-1]
			indentStack = indentStack[:len(indentStack)-1]
		}

		// Build full path
		var fullPath string
		if len(pathStack) > 0 {
			fullPath = strings.Join(pathStack, ".") + "." + lineKey
		} else {
			fullPath = lineKey
		}

		// Check if this matches the target key (simple key or full path)
		if lineKey != key && fullPath != key {
			// Add current key to path stack for nested elements
			pathStack = append(pathStack, lineKey)
			indentStack = append(indentStack, currentIndent)
			continue
		}

		// Found the key - now check if we should replace it
		newBlock := extractFlowBlock(lines, i)

		// Compare semantic content
		if shouldReplaceFlowBlock(newBlock, originalBlock) {
			// Calculate correct end index for replacement
			endIdx := i + len(newBlock)
			if endIdx > len(lines) {
				endIdx = len(lines)
			}

			// Check if content changed - if so, update the original block
			newContent := extractSemanticContent(newBlock)
			originalContent := extractSemanticContent(originalBlock)

			var blockToUse []string
			if newContent != originalContent {
				// Content changed - update the multiline block with new content
				blockToUse = updateMultilineFlowBlock(originalBlock, newBlock)
			} else {
				// Content same - use original formatting
				blockToUse = originalBlock
			}

			// Replace the new block with the updated/original block
			newLines := make([]string, 0, len(lines)-len(newBlock)+len(blockToUse))
			newLines = append(newLines, lines[:i]...)
			newLines = append(newLines, blockToUse...)
			if endIdx < len(lines) {
				newLines = append(newLines, lines[endIdx:]...)
			}

			return strings.Join(newLines, "\n")
		}

		break // Found the key, no need to continue
	}

	return content
}

// shouldReplaceFlowBlock determines if we should replace a new block with original formatting
func shouldReplaceFlowBlock(newBlock, originalBlock []string) bool {
	if len(newBlock) == 0 || len(originalBlock) == 0 {
		return false
	}

	// Extract semantic content from both blocks and compare
	newContent := extractSemanticContent(newBlock)
	originalContent := extractSemanticContent(originalBlock)

	// If semantic content is the same, use original formatting
	if newContent == originalContent {
		return true
	}

	// If new block is single line and original is multiline,
	// but content is different, we need to update the multiline format
	if len(newBlock) == 1 && len(originalBlock) > 1 {
		return true
	}

	return false
}

// extractSemanticContent extracts the semantic content from a flow block (ignoring formatting)
func extractSemanticContent(block []string) string {
	var content strings.Builder

	for _, line := range block {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Remove all whitespace to compare semantic content only
		compressed := strings.ReplaceAll(trimmed, " ", "")
		content.WriteString(compressed)
	}

	return content.String()
}

// updateMultilineFlowBlock updates a multiline flow block with new content
func updateMultilineFlowBlock(originalBlock, newBlock []string) []string {
	if len(newBlock) != 1 || len(originalBlock) == 0 {
		// Only handle single-line new block updating multiline original
		return originalBlock
	}

	// Extract new elements from the single-line block
	newLine := newBlock[0]
	newElements := extractArrayElementsFromLine(newLine)
	if len(newElements) == 0 {
		return originalBlock
	}

	// Update the multiline block with new elements
	return updateMultilineFlowWithElements(originalBlock, newElements)
}

// extractArrayElementsFromLine extracts array elements from a single line like "items: [a, b, c]"
func extractArrayElementsFromLine(line string) []string {
	// Find the array part
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		return nil
	}

	arrayPart := strings.TrimSpace(line[colonIdx+1:])
	if !strings.HasPrefix(arrayPart, "[") || !strings.HasSuffix(arrayPart, "]") {
		return nil
	}

	// Extract content between brackets
	content := arrayPart[1 : len(arrayPart)-1]
	if strings.TrimSpace(content) == "" {
		return []string{}
	}

	// Split by comma and trim each element
	elements := strings.Split(content, ",")
	for i, elem := range elements {
		elements[i] = strings.TrimSpace(elem)
	}

	return elements
}

// updateMultilineFlowWithElements updates a multiline flow block with new elements
func updateMultilineFlowWithElements(originalBlock []string, newElements []string) []string {
	if len(originalBlock) == 0 || len(newElements) == 0 {
		return originalBlock
	}

	// Create a copy of the original block
	updatedBlock := make([]string, len(originalBlock))
	copy(updatedBlock, originalBlock)

	// Find the last content line (before closing bracket)
	lastContentIdx := -1
	for i := len(updatedBlock) - 1; i >= 0; i-- {
		line := strings.TrimSpace(updatedBlock[i])
		if line != "" && line != "]" && line != "}" {
			lastContentIdx = i
			break
		}
	}

	if lastContentIdx == -1 {
		return originalBlock
	}

	// Extract current elements from the original block
	originalElements := extractElementsFromMultilineBlock(originalBlock)

	// If new elements are different, update the block
	if !equalStringSlices(originalElements, newElements) {
		// Get the indentation pattern from existing elements
		baseIndent := getLineIndentation(originalBlock[0])
		elementIndent := ""
		if len(originalBlock) > 1 {
			elementIndent = strings.Repeat(" ", getLineIndentation(originalBlock[1])-baseIndent)
		}

		// Rebuild the block
		var newBlock []string

		// Add the opening line
		newBlock = append(newBlock, originalBlock[0])

		// Add all elements
		for i, elem := range newElements {
			var line string
			if i == len(newElements)-1 {
				// Last element - no comma
				line = strings.Repeat(" ", baseIndent) + elementIndent + elem
			} else {
				// Not last element - add comma
				line = strings.Repeat(" ", baseIndent) + elementIndent + elem + ","
			}
			newBlock = append(newBlock, line)
		}

		// Add the closing bracket with same indentation as opening
		newBlock = append(newBlock, strings.Repeat(" ", baseIndent)+"]")

		return newBlock
	}

	return originalBlock
}

// extractElementsFromMultilineBlock extracts elements from a multiline flow block
func extractElementsFromMultilineBlock(block []string) []string {
	var elements []string

	for _, line := range block {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "{") ||
			trimmed == "]" || trimmed == "}" {
			continue
		}

		// Remove trailing comma
		elem := strings.TrimRight(trimmed, ",")
		if elem != "" {
			elements = append(elements, elem)
		}
	}

	return elements
}

// Legacy functions - keeping for now to avoid breaking other parts

// replaceMultilineFlow replaces multiline flow object in new content with original formatting
func replaceMultilineFlow(content, key string, originalFlowLines []string) string {
	lines := strings.Split(content, "\n")

	// Find the key line in new content
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, ":") && strings.HasPrefix(trimmed, key+":") {
			// Check if this became a single-line flow object or compact flow
			if (strings.Contains(trimmed, "{") && strings.Contains(trimmed, "}")) ||
				(strings.Contains(trimmed, "[") && strings.Contains(trimmed, "]")) {

				// Extract the new array content from the single line
				newArrayContent := extractArrayContent(trimmed)
				originalArrayContent := extractArrayContentFromLines(originalFlowLines)

				// If content changed, update the original with new elements
				if !equalStringSlices(newArrayContent, originalArrayContent) {
					updatedFlowLines := updateMultilineFlowWithNewContent(originalFlowLines, newArrayContent)

					// Replace single line with updated multiline original
					newLines := make([]string, 0, len(lines)-1+len(updatedFlowLines))
					newLines = append(newLines, lines[:i]...)
					newLines = append(newLines, updatedFlowLines...)
					newLines = append(newLines, lines[i+1:]...)

					return strings.Join(newLines, "\n")
				} else {
					// No changes, use original formatting
					newLines := make([]string, 0, len(lines)-1+len(originalFlowLines))
					newLines = append(newLines, lines[:i]...)
					newLines = append(newLines, originalFlowLines...)
					newLines = append(newLines, lines[i+1:]...)

					return strings.Join(newLines, "\n")
				}
			} else {
				// For multiline flow objects that stayed multiline,
				// use simple string comparison to check if they changed

				// Find the end of the current flow object in new content
				endLine := i
				bracketCount := 0
				braceCount := 0

				for j := i; j < len(lines); j++ {
					currentLine := lines[j]
					for _, r := range currentLine {
						if r == '[' {
							bracketCount++
						} else if r == ']' {
							bracketCount--
						} else if r == '{' {
							braceCount++
						} else if r == '}' {
							braceCount--
						}
					}

					if bracketCount == 0 && braceCount == 0 && j > i {
						endLine = j
						break
					}
				}

				// Compare the multiline flow sections as strings
				newFlowSection := strings.Join(lines[i:endLine+1], "\n")
				originalFlowSection := strings.Join(originalFlowLines, "\n")

				// Only replace if the sections are actually different
				if strings.TrimSpace(newFlowSection) != strings.TrimSpace(originalFlowSection) {
					// Content changed, use updated version but try to preserve formatting
					return content // For now, keep new content as-is
				} else {
					// Content unchanged, use original formatting
					newLines := make([]string, 0, len(lines)-(endLine-i)+len(originalFlowLines))
					newLines = append(newLines, lines[:i]...)
					newLines = append(newLines, originalFlowLines...)
					newLines = append(newLines, lines[endLine+1:]...)

					return strings.Join(newLines, "\n")
				}
			}
		}
	}

	return content
}

// extractArrayContent extracts array elements from a single line
func extractArrayContent(line string) []string {
	start := strings.Index(line, "[")
	end := strings.LastIndex(line, "]")
	if start >= 0 && end > start {
		content := line[start+1 : end]
		if strings.TrimSpace(content) == "" {
			return []string{}
		}
		elements := strings.Split(content, ",")
		for i, elem := range elements {
			elements[i] = strings.TrimSpace(elem)
		}
		return elements
	}
	return []string{}
}

// extractArrayContentFromLines extracts array elements from multiline flow
func extractArrayContentFromLines(lines []string) []string {
	var elements []string
	inArray := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "[") {
			inArray = true
			// Check if there's content after [
			if idx := strings.Index(trimmed, "["); idx >= 0 && idx+1 < len(trimmed) {
				after := strings.TrimSpace(trimmed[idx+1:])
				if after != "" && !strings.HasPrefix(after, "]") {
					// Remove trailing comma if present
					after = strings.TrimRight(after, ",")
					if after != "" {
						elements = append(elements, after)
					}
				}
			}
		} else if strings.Contains(trimmed, "]") {
			// Last element before ]
			if idx := strings.Index(trimmed, "]"); idx > 0 {
				before := strings.TrimSpace(trimmed[:idx])
				before = strings.TrimRight(before, ",")
				if before != "" {
					elements = append(elements, before)
				}
			}
			break
		} else if inArray && trimmed != "" {
			// Middle elements
			elem := strings.TrimRight(trimmed, ",")
			if elem != "" {
				elements = append(elements, elem)
			}
		}
	}

	return elements
}

// updateMultilineFlowWithNewContent updates original multiline flow with new elements
func updateMultilineFlowWithNewContent(originalLines []string, newElements []string) []string {
	if len(newElements) == 0 {
		return originalLines
	}

	var result []string
	elementIndex := 0
	inArray := false

	for _, line := range originalLines {
		trimmed := strings.TrimSpace(line)

		if strings.Contains(trimmed, "[") {
			inArray = true
			result = append(result, line)

			// Check if there's content after [
			if idx := strings.Index(trimmed, "["); idx >= 0 && idx+1 < len(trimmed) {
				after := strings.TrimSpace(trimmed[idx+1:])
				if after != "" && !strings.HasPrefix(after, "]") {
					elementIndex++
				}
			}
		} else if strings.Contains(trimmed, "]") {
			// Add any remaining new elements before closing
			// Find the proper indentation from previous elements
			elementIndent := 2 // default
			if len(result) > 1 {
				// Look for the last element's indentation
				for i := len(result) - 1; i >= 0; i-- {
					if strings.TrimSpace(result[i]) != "" && !strings.Contains(result[i], "[") {
						elementIndent = getLineIndentation(result[i])
						break
					}
				}
			}

			for elementIndex < len(newElements) {
				newLine := strings.Repeat(" ", elementIndent) + newElements[elementIndex]
				// Don't add comma to the last element
				if elementIndex < len(newElements)-1 {
					newLine += ","
				}
				result = append(result, newLine)
				elementIndex++
			}
			result = append(result, line)
			break
		} else if inArray && trimmed != "" {
			// Replace with new element if available
			if elementIndex < len(newElements) {
				indent := getLineIndentation(line)
				newLine := strings.Repeat(" ", indent) + newElements[elementIndex]
				// Add comma if there are more elements to come
				if elementIndex < len(newElements)-1 {
					newLine += ","
				}
				result = append(result, newLine)
				elementIndex++
			}
		} else {
			result = append(result, line)
		}
	}

	return result
}

// equalStringSlices compares two string slices for equality
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// isInsideInlineObject checks if a line is inside an inline object
func isInsideInlineObject(lines []string, lineIndex int) bool {
	if lineIndex >= len(lines) {
		return false
	}

	currentLine := lines[lineIndex]
	currentIndent := len(currentLine) - len(strings.TrimLeft(currentLine, " \t"))

	// Look backwards to find a potential inline object start
	for i := lineIndex - 1; i >= 0; i-- {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		lineIndent := len(line) - len(strings.TrimLeft(line, " \t"))

		// If we find a line with less indentation that contains a key, check if it's an inline object
		if lineIndent < currentIndent && strings.Contains(trimmed, ":") {
			// Check if this line starts an inline object
			if strings.Contains(trimmed, "{") && !strings.Contains(trimmed, "}") {
				// This could be the start of an inline object, check if it closes
				return isInlineObject(lines, i)
			}
			// If we find a regular key at a lower indent level, we're not inside an inline object
			return false
		}
	}

	return false
}

// applyArrayStyles applies array formatting styles to the content
func applyArrayStyles(content string, info *FormattingInfo) string {
	if len(info.ArrayStyles) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, ":") {
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				key := strings.TrimSpace(trimmed[:idx])

				// Check if this key has a specific array style
				// But skip keys that are inside inline objects
				if style, exists := info.ArrayStyles[key]; exists && style.IsFlow && !isInsideInlineObject(lines, i) {
					value := line[idx+1:]

					// Find the array content
					if strings.Contains(value, "[") && strings.Contains(value, "]") {
						// Extract the array content
						start := strings.Index(value, "[")
						end := strings.LastIndex(value, "]")
						if start >= 0 && end > start {
							arrayContent := value[start+1 : end]

							// Apply the specific style
							var newArrayContent string
							if style.HasSpaces {
								// Add spaces around elements: [1,2,3] -> [ 1 , 2 , 3 ]
								elements := strings.Split(arrayContent, ",")
								for j, elem := range elements {
									elements[j] = " " + strings.TrimSpace(elem) + " "
								}
								newArrayContent = strings.Join(elements, ",")
							} else if style.IsCompact {
								// Remove all spaces: [ 1 , 2 , 3 ] -> [1,2,3]
								elements := strings.Split(arrayContent, ",")
								for j, elem := range elements {
									elements[j] = strings.TrimSpace(elem)
								}
								newArrayContent = strings.Join(elements, ",")
							} else {
								// Default formatting
								elements := strings.Split(arrayContent, ",")
								for j, elem := range elements {
									elements[j] = strings.TrimSpace(elem)
								}
								newArrayContent = strings.Join(elements, ", ")
							}

							// Rebuild the line
							newValue := value[:start+1] + newArrayContent + value[end:]
							lines[i] = line[:idx+1] + newValue
						}
					}
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// applyFlowObjectStyles applies original flow object formatting to preserve exact spacing
func applyFlowObjectStyles(content string, info *FormattingInfo) string {
	if len(info.FlowObjectStyles) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	// Track document structure to build full paths for keys
	pathStack := make([]string, 0)
	indentStack := make([]int, 0)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, ":") {
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				key := strings.TrimSpace(trimmed[:idx])
				currentIndent := getLineIndentation(line)

				// Update path stack based on indentation
				for len(indentStack) > 0 && indentStack[len(indentStack)-1] >= currentIndent {
					pathStack = pathStack[:len(pathStack)-1]
					indentStack = indentStack[:len(indentStack)-1]
				}

				// Build full path
				var fullPath string
				if len(pathStack) > 0 {
					fullPath = strings.Join(pathStack, ".") + "." + key
				} else {
					fullPath = key
				}

				// Check both the simple key and the full path for flow object styles
				var originalStyle string
				var exists bool
				if originalStyle, exists = info.FlowObjectStyles[key]; !exists {
					// If simple key doesn't exist, try the full path
					originalStyle, exists = info.FlowObjectStyles[fullPath]
				}

				if exists {
					// Extract current value part after the colon
					valueStart := strings.Index(line, ":") + 1
					if valueStart < len(line) {
						currentValue := strings.TrimSpace(line[valueStart:])

						// Only process if this is a flow object starting with {
						if strings.HasPrefix(currentValue, "{") {
							// Check if the current value is a single-line collapsed version
							// like "{cpu: 512, memory: 512}" vs original multiline format
							if !strings.Contains(currentValue, "\n") && strings.Contains(originalStyle, "\n") {
								// This is a collapsed flow object, we need to extract new values
								// and apply them to the original multiline format
								newValues := extractFlowObjectValues(currentValue)
								if len(newValues) > 0 {
									// Update the original style with new values
									updatedStyle := updateFlowObjectWithNewValues(originalStyle, newValues)

									// Replace the value part with updated multiline style
									newLine := line[:valueStart] + " " + updatedStyle
									lines[i] = newLine
								}
							} else if !strings.Contains(currentValue, "\n") && !strings.Contains(originalStyle, "\n") {
								// Both are single-line - apply original formatting with updated values
								currentValues := extractFlowObjectValues(currentValue)
								originalValues := extractFlowObjectValues(originalStyle)

								// Only apply if values actually changed
								valuesChanged := false
								for k, newVal := range currentValues {
									if origVal, ok := originalValues[k]; !ok || origVal != newVal {
										valuesChanged = true
										break
									}
								}

								if valuesChanged {
									updatedStyle := updateFlowObjectWithNewValues(originalStyle, currentValues)
									newLine := line[:valueStart] + " " + updatedStyle
									lines[i] = newLine
								}
							}
						}
					}
				}

				// Add current key to path stack for nested elements
				if strings.Contains(line, ":") && !strings.HasSuffix(strings.TrimSpace(line), "}") {
					pathStack = append(pathStack, key)
					indentStack = append(indentStack, currentIndent)
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// updateFlowObjectWithNewValues updates values in original flow object with new values map
func updateFlowObjectWithNewValues(originalStyle string, newValues map[string]string) string {
	if len(newValues) == 0 {
		return originalStyle
	}

	result := originalStyle
	originalValues := extractFlowObjectValues(originalStyle)

	// Update each value that exists in the new values
	for key, newValue := range newValues {
		if originalValue, exists := originalValues[key]; exists && originalValue != newValue {
			// Replace the old value with new value while preserving surrounding formatting
			result = replaceValueInFlowObject(result, key, originalValue, newValue)
		}
	}

	return result
}

// extractFlowObjectValues extracts key-value pairs from flow object string
func extractFlowObjectValues(flowStr string) map[string]string {
	values := make(map[string]string)

	// Remove outer braces/brackets
	inner := flowStr
	if strings.HasPrefix(inner, "{") && strings.HasSuffix(inner, "}") {
		inner = inner[1 : len(inner)-1]
	} else if strings.HasPrefix(inner, "[") && strings.HasSuffix(inner, "]") {
		inner = inner[1 : len(inner)-1]
	}

	// Split by comma but be careful about nested structures
	parts := splitFlowObjectParts(inner)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, ":") {
			idx := strings.Index(part, ":")
			key := strings.TrimSpace(part[:idx])
			value := strings.TrimSpace(part[idx+1:])
			values[key] = value
		}
	}

	return values
}

// splitFlowObjectParts splits flow object content by commas, respecting nested structures
func splitFlowObjectParts(content string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for _, r := range content {
		switch r {
		case '{', '[':
			depth++
			current.WriteRune(r)
		case '}', ']':
			depth--
			current.WriteRune(r)
		case ',':
			if depth == 0 {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(r)
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// replaceValueInFlowObject replaces a specific value in flow object while preserving formatting
func replaceValueInFlowObject(flowStr, key, oldValue, newValue string) string {
	// Use simple string replacement pattern: "key: oldValue" -> "key: newValue"
	// This preserves all the surrounding formatting
	oldPattern := key + ": " + oldValue
	newPattern := key + ": " + newValue

	result := strings.Replace(flowStr, oldPattern, newPattern, 1)

	// If that didn't work, try without space after colon
	if result == flowStr {
		oldPattern = key + ":" + oldValue
		newPattern = key + ":" + newValue
		result = strings.Replace(flowStr, oldPattern, newPattern, 1)
	}

	return result
}

// applyExactIndentations applies exact indentations for keys that had custom indents
func applyExactIndentations(content string, info *FormattingInfo) string {
	if len(info.KeyIndents) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle array elements
		if strings.HasPrefix(trimmed, "- ") {
			currentIndent := getLineIndentation(line)
			// Look for a matching array element indentation
			arrayElementKey := fmt.Sprintf("__array_element_%d__", currentIndent)
			if exactIndent, exists := info.KeyIndents[arrayElementKey]; exists {
				if currentIndent != exactIndent {
					// Replace the line with correct indentation
					newLine := strings.Repeat(" ", exactIndent) + trimmed
					lines[i] = newLine
				}
			} else {
				// Try to find any array element indentation pattern
				for key, exactIndent := range info.KeyIndents {
					if strings.HasPrefix(key, "__array_element_") {
						// Use this indentation for array elements
						newLine := strings.Repeat(" ", exactIndent) + trimmed
						lines[i] = newLine
						break
					}
				}
			}
		} else if strings.Contains(trimmed, ":") && !strings.HasPrefix(trimmed, "#") {
			// Handle regular keys
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				key := strings.TrimSpace(trimmed[:idx])

				// Check if this key has a specific indentation
				if exactIndent, exists := info.KeyIndents[key]; exists {
					currentIndent := getLineIndentation(line)

					// Only apply indentation if the current indentation is different from expected
					// and it's not a case where we're trying to add indentation to a root key
					if currentIndent != exactIndent {
						// Special case: if exactIndent is 0 and currentIndent is also 0, don't change
						if exactIndent == 0 && currentIndent == 0 {
							continue
						}
						// Replace the line with correct indentation
						newLine := strings.Repeat(" ", exactIndent) + trimmed
						lines[i] = newLine
					}
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// detectIndentation analyzes the raw YAML to determine the indentation level
func detectIndentation(raw string) int {
	info := detectFormattingInfoOptimized(raw)
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
// SetCommentAlignment configures how inline comments should be aligned
func (d *Document) SetCommentAlignment(mode CommentAlignmentMode) {
	if d.formattingCache == nil {
		d.formattingCache = detectFormattingInfoOptimized(d.raw)
	}
	d.formattingCache.AlignmentMode = mode
}

// SetAbsoluteCommentAlignment aligns all comments to the specified column
func (d *Document) SetAbsoluteCommentAlignment(column int) {
	if d.formattingCache == nil {
		d.formattingCache = detectFormattingInfoOptimized(d.raw)
	}
	d.formattingCache.AlignmentMode = CommentAlignmentAbsolute
	d.formattingCache.CommentSpacing = column
}

// EnableRelativeCommentAlignment preserves original spacing between values and comments
func (d *Document) EnableRelativeCommentAlignment() {
	if d.formattingCache == nil {
		d.formattingCache = detectFormattingInfoOptimized(d.raw)
	}
	d.formattingCache.AlignmentMode = CommentAlignmentRelative
}

// DisableCommentAlignment disables all comment alignment processing
func (d *Document) DisableCommentAlignment() {
	if d.formattingCache == nil {
		d.formattingCache = detectFormattingInfoOptimized(d.raw)
	}
	d.formattingCache.AlignmentMode = CommentAlignmentDisabled
}

func (d *Document) String() (string, error) {
	bytes, err := d.ToBytes()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
