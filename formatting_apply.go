// Post-processing pipeline that restores original formatting on top of the
// yaml.v3 encoder output. preserveOriginalFormatting is the entry point.

package yamler

import (
	"fmt"
	"strings"
)

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
