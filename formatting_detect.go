// Detection of formatting information (indentation, flow styles, comment
// alignment, blank lines, ...) from the raw YAML text.

package yamler

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// detectFormattingInfoOptimized builds the formatting snapshot of a raw YAML
// text. Every hint is keyed by the path of the line it was found on (see
// lineWalker), so keys with the same name at different places do not
// interfere.
func detectFormattingInfoOptimized(raw string) *FormattingInfo {
	info := &FormattingInfo{
		IndentSize:       2,
		EmptyLines:       make(map[string]int),
		FlowStyles:       make(map[string]bool),
		ScalarStyles:     make(map[string]yaml.Style),
		MultilineFlow:    make(map[string]bool),
		ZeroIndentArrays: make(map[string]bool),
		CommentAlignment: make(map[string]int),
		AlignmentMode:    CommentAlignmentRelative,
		ArrayStyles:      make(map[string]*ArrayStyle),
		KeyIndents:       make(map[string]int),
		FlowObjectStyles: make(map[string]string),
		CommentIndents:   make(map[string]int),
	}

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		detectLineMarkers(line, info)
	}

	indentLevels := detectLineFormatting(lines, info, make([]int, 0, 32))

	// Find the most common indentation increment if not using tabs
	if !info.UseTabs && len(indentLevels) > 0 {
		if baseIndent := findBaseIndentationOptimized(indentLevels); baseIndent > 0 {
			info.IndentSize = baseIndent
		}
	}

	if len(info.CommentAlignment) > 0 {
		info.CommentSpacing = findCommonCommentAlignment(info.CommentAlignment)
	}

	return info
}

// collectMultilineFlowObject collects a complete multiline flow object starting from startLine
func collectMultilineFlowObject(lines []string, startLine int, firstValue string, openBrace, closeBrace rune) string {
	var result strings.Builder
	depth := 0

	for i := startLine; i < len(lines); i++ {
		line := lines[i]

		if i == startLine {
			// For the first line, only take the value part after the key
			result.WriteString(firstValue)
			line = firstValue
		} else {
			// For subsequent lines, take the trimmed content
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				result.WriteString("\n")
				result.WriteString(line)
			}
		}

		// Count brackets AFTER adding the line to result, ignoring comments
		// and quoted strings
		counted := strings.TrimSpace(line)
		if strings.HasPrefix(counted, "#") {
			continue
		}
		depth += flowBracketBalance(stripInlineComment(counted))
		if depth <= 0 {
			return result.String()
		}
	}

	// If we get here, we didn't find a complete object
	return ""
}

// detectLineMarkers records tab usage and document markers.
func detectLineMarkers(line string, info *FormattingInfo) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return
	}
	if strings.HasPrefix(line, "\t") {
		info.UseTabs = true
		info.IndentSize = 4
	}
	switch trimmed {
	case "---":
		info.HasDocumentStart = true
	case "...":
		info.HasDocumentEnd = true
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

	// Ensure result is reasonable (between 2 and 8)
	if result < 2 {
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

// detectLineFormatting walks the document once and records, keyed by path:
// exact key and item columns (KeyIndents: "a.b" for keys, "a.b[]" for the
// items of a sequence), zero-indent sequences, blank lines before a line
// (EmptyLines, keyed with indices, or by comment text for comment lines),
// inline comment spacing (CommentAlignment, with indices), flow styles,
// flow object text (FlowObjectStyles), flow array styles (ArrayStyles) and
// literal/folded scalar styles. It returns the indentation levels of the
// structural lines for indentation-size detection.
func detectLineFormatting(lines []string, info *FormattingInfo, indentLevels []int) []int {
	w := newLineWalker()
	var lastKey *lineInfo
	emptyBefore := 0
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			emptyBefore++
			continue
		}
		li := w.next(line)
		if li.skip {
			if strings.HasPrefix(trimmed, "#") {
				if emptyBefore > 0 {
					info.EmptyLines[trimmed] = emptyBefore
				}
				if prev, seen := info.CommentIndents[trimmed]; seen && prev != li.indent {
					info.CommentIndents[trimmed] = -1
				} else if !seen {
					info.CommentIndents[trimmed] = li.indent
				}
			}
			emptyBefore = 0
			continue
		}
		if li.indent > 0 && !strings.HasPrefix(line, "\t") {
			indentLevels = append(indentLevels, li.indent)
		}
		if emptyBefore > 0 && (li.key != "" || li.isItem) {
			info.EmptyLines[li.idxPath] = emptyBefore
		}
		emptyBefore = 0

		// Paths ignore indices, so all items of a sequence share a record;
		// the first occurrence wins when items are laid out differently.
		if li.isItem {
			if _, seen := info.KeyIndents[li.path+"[]"]; !seen {
				info.KeyIndents[li.path+"[]"] = li.indent
			}
			if lastKey != nil && lastKey.keyPath == li.path && keyColumn(*lastKey) == li.indent {
				info.ZeroIndentArrays[li.path] = true
			}
		}

		if c := inlineCommentIndex(line); c > 0 {
			before := line[:c]
			if spaces := len(before) - len(strings.TrimRight(before, " ")); spaces > 0 && strings.TrimSpace(before) != "" {
				info.CommentAlignment[li.idxPath] = spaces
			}
		}

		if li.key == "" {
			lastKey = nil
			continue
		}
		if _, seen := info.KeyIndents[li.keyPath]; !seen {
			info.KeyIndents[li.keyPath] = li.keyCol
		}
		cp := li
		lastKey = &cp

		value := stripInlineComment(li.value)
		if value == "" {
			continue
		}
		switch value[0] {
		case '|':
			info.ScalarStyles[li.keyPath] = yaml.LiteralStyle
		case '>':
			info.ScalarStyles[li.keyPath] = yaml.FoldedStyle
		case '{', '[':
			if inlineCommentIndex(li.value) < 0 {
				// Flow values with comments inside are left to the encoder.
				detectFlowValue(lines, i, value, li.keyPath, info)
			}
		}
	}
	return indentLevels
}

// detectFlowValue records the style of a flow mapping or sequence value.
func detectFlowValue(lines []string, i int, value, keyPath string, info *FormattingInfo) {
	info.FlowStyles[keyPath] = true
	open, closing := value[0], byte('}')
	if open == '[' {
		closing = ']'
	}
	if strings.IndexByte(value, closing) >= 0 {
		// Complete single-line flow value.
		info.FlowObjectStyles[keyPath] = value
	} else {
		obj := collectMultilineFlowObject(lines, i, value, rune(open), rune(closing))
		if obj == "" || strings.Contains(obj, "#") {
			// Unterminated, or with comments inside: left to the encoder.
			delete(info.FlowStyles, keyPath)
			return
		}
		info.MultilineFlow[keyPath] = true
		info.FlowObjectStyles[keyPath] = obj
	}
	if open != '[' {
		return
	}
	style := &ArrayStyle{IsFlow: true}
	if strings.HasSuffix(value, "]") {
		inner := value[1 : len(value)-1]
		if strings.Contains(inner, " , ") || (strings.HasPrefix(inner, " ") && strings.HasSuffix(inner, " ")) {
			style.HasSpaces = true
		} else if !strings.Contains(inner, " ") {
			style.IsCompact = true
		}
	} else {
		style.IsMultiline = true
	}
	info.ArrayStyles[keyPath] = style
}
