// Post-processing pipeline that restores original formatting on top of the
// yaml.v3 encoder output. preserveOriginalFormatting is the entry point.

package yamler

import (
	"strings"
)

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

	// Preserve multiline flow formatting
	// TODO: Fix path resolution issue - temporarily disabled
	// newStr = preserveMultilineFlow(newStr, original, info)

	// Apply array styles
	newStr = applyArrayStyles(newStr, info)

	// Apply flow object styles to preserve spacing
	newStr = applyFlowObjectStyles(newStr, info)

	// Apply zero-indent array formatting before exact indentations so that
	// keys nested in zero-indent items are measured at their final column
	newStr = applyZeroIndentArrays(newStr, info)

	// Apply exact key indentations
	newStr = applyExactIndentations(newStr, info)

	// Apply empty line patterns (after indentation to avoid conflicts)
	newStr = applyEmptyLinePatterns(newStr, info)

	// Preserve folded scalar formatting
	newStr = preserveFoldedScalars(newStr, original, info)

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

// applyExactIndentations moves every key and sequence item that existed in
// the original document back to its original column. When a line is moved,
// the block nested under it moves with it, so lines added by modifications
// keep their position relative to their parent.
func applyExactIndentations(content string, info *FormattingInfo) string {
	if len(info.KeyIndents) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	w := newLineWalker()
	for i := 0; i < len(lines); i++ {
		li := w.next(lines[i])
		if li.skip {
			continue
		}
		var want int
		var ok bool
		switch {
		case li.isItem:
			want, ok = info.KeyIndents[li.path+"[]"]
		case li.key != "":
			want, ok = info.KeyIndents[li.keyPath]
		}
		if !ok || want == li.indent {
			continue
		}
		shiftBlock(lines, i, li.indent, want-li.indent)
	}

	return strings.Join(lines, "\n")
}

// shiftBlock changes the indentation of lines[start] and of every following
// line that is nested under it (indented deeper than blockIndent) by delta
// columns. Blank lines are left untouched.
func shiftBlock(lines []string, start, blockIndent, delta int) {
	for j := start; j < len(lines); j++ {
		trimmed := strings.TrimSpace(lines[j])
		if trimmed == "" {
			continue
		}
		indent := getLineIndentation(lines[j])
		if j > start && indent <= blockIndent {
			break
		}
		newIndent := indent + delta
		if newIndent < 0 {
			newIndent = 0
		}
		lines[j] = strings.Repeat(" ", newIndent) + strings.TrimLeft(lines[j], " ")
	}
}

// restoreDocumentSeparators adds back document separators if they were in the original
func restoreDocumentSeparators(content string, info *FormattingInfo, originalContent string, preserveDocumentSeparator bool) string {
	// Check if the original content actually starts with ---
	originallyHadDocumentStart := strings.HasPrefix(strings.TrimSpace(originalContent), "---")
	originallyHadDocumentEnd := lastLineIs(originalContent, "...")

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

// applyZeroIndentArrays re-indents block sequences that were written in
// zero-indent style in the original document. The encoder always indents
// items relative to the parent key, so each such sequence is shifted left by
// the difference between the item column and the key column.
func applyZeroIndentArrays(content string, info *FormattingInfo) string {
	if len(info.ZeroIndentArrays) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	w := newLineWalker()
	for i := 0; i < len(lines); i++ {
		li := w.next(lines[i])
		if li.skip || li.key == "" || !info.ZeroIndentArrays[li.keyPath] {
			continue
		}
		// Find the first item of the sequence that follows this key.
		j := i + 1
		for j < len(lines) && (strings.TrimSpace(lines[j]) == "" || strings.HasPrefix(strings.TrimSpace(lines[j]), "#")) {
			j++
		}
		if j >= len(lines) {
			break
		}
		first := strings.TrimSpace(lines[j])
		itemIndent := getLineIndentation(lines[j])
		if (first != "-" && !strings.HasPrefix(first, "- ")) || itemIndent <= li.indent {
			continue
		}
		shift := itemIndent - li.indent
		// Shift the whole sequence block (items and their nested content).
		for ; j < len(lines); j++ {
			trimmed := strings.TrimSpace(lines[j])
			if trimmed == "" {
				continue
			}
			if getLineIndentation(lines[j]) < itemIndent {
				break
			}
			lines[j] = lines[j][shift:]
		}
	}

	return strings.Join(lines, "\n")
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

func leadingWhitespace(line string) string {
	for i, r := range line {
		if r != ' ' && r != '\t' {
			return line[:i]
		}
	}
	return line
}

// lastLineIs reports whether the last non-empty line of content is exactly marker.
func lastLineIs(content, marker string) bool {
	content = strings.TrimRight(content, " \t\r\n")
	if i := strings.LastIndexByte(content, '\n'); i >= 0 {
		content = content[i+1:]
	}
	return strings.TrimSpace(content) == marker
}
