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

// keyValueStart returns the index in line just after the ":" of its key.
func keyValueStart(line string, li lineInfo) int {
	trimmed := strings.TrimLeft(line, " ")
	offset := len(line) - len(trimmed)
	if li.isItem {
		trimmed = strings.TrimLeft(trimmed[1:], " ")
		offset = len(line) - len(trimmed)
	}
	// Skip a quoted key.
	i := 0
	if trimmed != "" && (trimmed[0] == '"' || trimmed[0] == '\'') {
		q := trimmed[0]
		for i = 1; i < len(trimmed) && trimmed[i] != q; i++ {
		}
	}
	for ; i < len(trimmed); i++ {
		if trimmed[i] == ':' && (i+1 == len(trimmed) || trimmed[i+1] == ' ' || trimmed[i+1] == '\t') {
			return offset + i + 1
		}
	}
	return -1
}

// applyArrayStyles restores the original style of flow arrays that the
// encoder has rewritten on a single line with default spacing.
func applyArrayStyles(content string, info *FormattingInfo) string {
	if len(info.ArrayStyles) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	w := newLineWalker()
	for i, line := range lines {
		li := w.next(line)
		if li.skip || li.key == "" {
			continue
		}
		style, ok := info.ArrayStyles[li.keyPath]
		if !ok || !style.IsFlow {
			continue
		}
		vs := keyValueStart(line, li)
		if vs < 0 {
			continue
		}
		value := line[vs:]
		start := strings.Index(value, "[")
		end := strings.LastIndex(value, "]")
		if start < 0 || end <= start {
			continue
		}

		if originalFlow, ok := info.FlowObjectStyles[li.keyPath]; ok && style.IsMultiline && strings.Contains(originalFlow, "\n") {
			// Multi-line flow array collapsed by the encoder: restore the
			// original layout, updating elements if they changed.
			current := extractCurrentArrayElements(value)
			original := extractCurrentArrayElements(originalFlow)
			if !equalStringSlices(current, original) {
				lines[i] = line[:vs] + " " + updateMultilineFlowArrayWithElements(originalFlow, current)
			} else {
				lines[i] = line[:vs] + " " + originalFlow
			}
			continue
		}

		elements := strings.Split(value[start+1:end], ",")
		for j, elem := range elements {
			elements[j] = strings.TrimSpace(elem)
		}
		var inner string
		switch {
		case style.HasSpaces:
			for j := range elements {
				elements[j] = " " + elements[j] + " "
			}
			inner = strings.Join(elements, ",")
		case style.IsCompact:
			inner = strings.Join(elements, ",")
		default:
			inner = strings.Join(elements, ", ")
		}
		if len(elements) == 1 && elements[0] == "" {
			inner = ""
		}
		lines[i] = line[:vs] + value[:start+1] + inner + value[end:]
	}

	return strings.Join(lines, "\n")
}

// applyFlowObjectStyles restores the original text of flow mappings
// ({key: value}), including multi-line ones, updating values that changed.
func applyFlowObjectStyles(content string, info *FormattingInfo) string {
	if len(info.FlowObjectStyles) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	w := newLineWalker()
	for i, line := range lines {
		li := w.next(line)
		if li.skip || li.key == "" {
			continue
		}
		originalStyle, ok := info.FlowObjectStyles[li.keyPath]
		if !ok || !strings.HasPrefix(originalStyle, "{") {
			continue
		}
		vs := keyValueStart(line, li)
		if vs < 0 {
			continue
		}
		currentValue := strings.TrimSpace(line[vs:])
		if !strings.HasPrefix(currentValue, "{") {
			continue
		}
		currentValues := extractFlowObjectValues(currentValue)
		if strings.Contains(originalStyle, "\n") {
			// Collapsed multi-line flow object: put the new values into the
			// original layout.
			if len(currentValues) > 0 {
				lines[i] = line[:vs] + " " + updateFlowObjectWithNewValues(originalStyle, currentValues)
			}
			continue
		}
		if currentValue != originalStyle {
			lines[i] = line[:vs] + " " + updateFlowObjectWithNewValues(originalStyle, currentValues)
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
