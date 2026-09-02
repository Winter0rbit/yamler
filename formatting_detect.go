// Detection of formatting information (indentation, flow styles, comment
// alignment, blank lines, ...) from the raw YAML text.

package yamler

import (
	"strings"

	"gopkg.in/yaml.v3"
)

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

	// Process lines to detect formatting patterns
	lines := strings.Split(raw, "\n")
	emptyLineCount := 0

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			emptyLineCount++
		} else {
			processLineOptimized(line, i, emptyLineCount, info)
			emptyLineCount = 0
		}
	}

	// Structural lines only (no block scalar contents or comments) decide
	// the indentation step.
	indentLevels := detectLineIndents(lines, info, make([]int, 0, 32))

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
func processLineOptimized(line string, lineNum, emptyLinesBefore int, info *FormattingInfo) {
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

	if leadingTabs > 0 {
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

	// Array elements (lines starting with "- ") are handled by detectLineIndents
	if strings.HasPrefix(content, "- ") {
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

// detectIndentation analyzes the raw YAML to determine the indentation level
func detectIndentation(raw string) int {
	info := detectFormattingInfoOptimized(raw)
	return info.IndentSize
}

// detectLineIndents records the exact column of every mapping key and of
// every sequence item, keyed by path (see lineWalker). Items are stored
// under "<sequence path>[]". Sequences whose items start at the same column
// as their parent key (kubectl / GitHub Actions style) are additionally
// recorded in ZeroIndentArrays.
func detectLineIndents(lines []string, info *FormattingInfo, indentLevels []int) []int {
	w := newLineWalker()
	var lastKey *lineInfo
	for _, line := range lines {
		li := w.next(line)
		if li.skip {
			continue
		}
		if li.indent > 0 && !strings.HasPrefix(line, "\t") {
			indentLevels = append(indentLevels, li.indent)
		}
		if li.isItem {
			info.KeyIndents[li.path+"[]"] = li.indent
			if lastKey != nil && lastKey.keyPath == li.path && lastKey.indent == li.indent {
				info.ZeroIndentArrays[li.path] = true
			}
		}
		if li.key != "" {
			keyIndent := li.indent
			if li.isItem {
				keyIndent += 2
			}
			info.KeyIndents[li.keyPath] = keyIndent
			cp := li
			lastKey = &cp
		} else {
			lastKey = nil
		}
	}
	return indentLevels
}
