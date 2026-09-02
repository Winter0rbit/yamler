// Preservation of flow-style ([a, b] / {k: v}) arrays and objects.

package yamler

import (
	"strings"
)

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

					// Check if we have original multiline flow format stored
					if originalFlow, exists := info.FlowObjectStyles[key]; exists && style.IsMultiline && strings.Contains(originalFlow, "\n") {
						// This is a multiline flow array that was collapsed to single line
						// But we need to check if the current content has changed and update accordingly
						currentArrayContent := extractCurrentArrayElements(value)
						originalArrayContent := extractCurrentArrayElements(originalFlow)

						if !equalStringSlices(currentArrayContent, originalArrayContent) {
							// Content has changed, update the multiline format with new elements
							updatedFlow := updateMultilineFlowArrayWithElements(originalFlow, currentArrayContent)
							lines[i] = line[:idx+1] + " " + updatedFlow
						} else {
							// Content is the same, restore original format
							lines[i] = line[:idx+1] + " " + originalFlow
						}
					} else if strings.Contains(value, "[") && strings.Contains(value, "]") {
						// Extract the array content for non-multiline arrays
						start := strings.Index(value, "[")
						end := strings.LastIndex(value, "]")
						if start >= 0 && end > start {
							arrayContent := value[start+1 : end]

							// Apply the specific style
							var newArrayContent string
							if style.IsMultiline && !strings.Contains(arrayContent, "\n") {
								// Convert single-line to multiline
								elements := strings.Split(arrayContent, ",")
								for j, elem := range elements {
									elements[j] = strings.TrimSpace(elem)
								}
								// Create multiline format
								indent := "  " // 2-space indent for array elements
								newArrayContent = "\n" + indent + strings.Join(elements, ",\n"+indent) + "\n"
							} else if style.HasSpaces {
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

								// Check if values changed
								valuesChanged := false
								for k, newVal := range currentValues {
									if origVal, ok := originalValues[k]; !ok || origVal != newVal {
										valuesChanged = true
										break
									}
								}

								// Always apply original formatting to preserve spaces, even if values didn't change
								// This handles cases where YAML encoder strips spaces but we want to preserve them
								if valuesChanged || currentValue != originalStyle {
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

// extractCurrentArrayElements extracts array elements from both single-line and multiline formats
func extractCurrentArrayElements(arrayStr string) []string {
	if arrayStr == "" {
		return nil
	}

	// Remove leading/trailing whitespace
	arrayStr = strings.TrimSpace(arrayStr)

	// Check if it's a flow array
	if !strings.HasPrefix(arrayStr, "[") || !strings.HasSuffix(arrayStr, "]") {
		return nil
	}

	// Extract content between brackets
	content := arrayStr[1 : len(arrayStr)-1]

	// Split by comma and clean up
	var elements []string
	parts := splitFlowObjectParts(content)

	for _, part := range parts {
		element := strings.TrimSpace(part)
		if element != "" {
			elements = append(elements, element)
		}
	}

	return elements
}

// updateMultilineFlowArrayWithElements updates a multiline flow array with new elements
func updateMultilineFlowArrayWithElements(originalFlow string, newElements []string) string {
	if len(newElements) == 0 {
		return originalFlow
	}

	lines := strings.Split(originalFlow, "\n")
	if len(lines) == 0 {
		return originalFlow
	}

	// Find the indentation of array elements by looking at the first element line
	var elementIndent string
	for i := 1; i < len(lines)-1; i++ { // Skip first '[' and last ']' lines
		line := lines[i]
		if strings.TrimSpace(line) != "" {
			// Get the indentation from the first non-empty element line
			for j, r := range line {
				if r != ' ' && r != '\t' {
					elementIndent = line[:j]
					break
				}
			}
			break
		}
	}

	// If we couldn't detect indentation, use default 2 spaces
	if elementIndent == "" {
		elementIndent = "  "
	}

	// Build new multiline array
	result := "[\n"

	for i, element := range newElements {
		result += elementIndent + element
		if i < len(newElements)-1 {
			result += ","
		}
		result += "\n"
	}

	result += "]"

	return result
}
