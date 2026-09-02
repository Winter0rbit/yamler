// Preservation of folded (>) block scalar formatting.

package yamler

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type foldedScalarInfo struct {
	key           string
	keyIndent     int
	contentIndent string
	indicator     string
	originalLines []string
	listItems     bool
}

// preserveFoldedScalars preserves original formatting of folded scalars
func preserveFoldedScalars(newContent, original string, info *FormattingInfo) string {
	// Find folded scalars in original content and preserve their formatting
	originalLines := strings.Split(original, "\n")

	foldedInfo := make(map[string]foldedScalarInfo)

	// Extract original folded scalar formatting info
	for i, line := range originalLines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, ">") && strings.Contains(trimmed, ":") {
			if idx := strings.Index(trimmed, ":"); idx > 0 {
				key := strings.TrimSpace(trimmed[:idx])
				if info.ScalarStyles[key] == yaml.FoldedStyle {
					indent := getLineIndentation(line)
					contentIndent := ""
					listItems := true
					foundContent := false

					var foldedLines []string
					for j := i + 1; j < len(originalLines); j++ {
						nextLine := originalLines[j]
						if strings.TrimSpace(nextLine) == "" {
							foldedLines = append(foldedLines, "")
							continue
						}

						nextIndent := getLineIndentation(nextLine)
						if nextIndent > indent {
							foldedLines = append(foldedLines, nextLine)
							if !foundContent {
								contentIndent = leadingWhitespace(nextLine)
								foundContent = true
							}
							if !strings.HasPrefix(strings.TrimSpace(nextLine), "-") {
								listItems = false
							}
						} else {
							break
						}
					}

					if !foundContent {
						contentIndent = strings.Repeat(" ", indent+info.IndentSize)
						listItems = false
					}

					indicator := extractFoldedIndicator(line)
					foldedInfo[key] = foldedScalarInfo{
						key:           key,
						keyIndent:     indent,
						contentIndent: contentIndent,
						indicator:     indicator,
						originalLines: foldedLines,
						listItems:     listItems,
					}
				}
			}
		}
	}

	// Replace folded scalars in new content with preserved formatting
	for key, originalInfo := range foldedInfo {
		newContent = replaceFoldedScalar(newContent, key, originalInfo)
	}

	return newContent
}

// replaceFoldedScalar replaces folded scalar content in new YAML with original formatting
func replaceFoldedScalar(content, key string, info foldedScalarInfo) string {
	lines := strings.Split(content, "\n")

	// Find the key line in new content
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+":") {
			colonPos := strings.Index(line, ":")
			if colonPos < 0 {
				continue
			}

			indent := getLineIndentation(line)
			lineIndent := leadingWhitespace(line)
			valuePart := strings.TrimSpace(line[colonPos+1:])
			valuePart = stripInlineComment(valuePart)
			hasFoldedIndicator := strings.HasPrefix(valuePart, ">")

			// Remove old folded content
			endIdx := i + 1
			var newValueLines []string
			if hasFoldedIndicator {
				for j := i + 1; j < len(lines); j++ {
					if strings.TrimSpace(lines[j]) == "" {
						newValueLines = append(newValueLines, "")
						endIdx = j + 1
						continue
					}
					lineIndent := getLineIndentation(lines[j])
					if lineIndent > indent {
						newValueLines = append(newValueLines, lines[j])
						endIdx = j + 1
					} else {
						break
					}
				}
			} else if valuePart != "" {
				newValueLines = append(newValueLines, valuePart)
			}

			originalValue := foldScalarValueFromLines(info.originalLines)
			newValue := foldScalarValueFromLines(newValueLines)
			replacementLines := formatFoldedScalarLines(newValueLines, info)
			if len(info.originalLines) > 0 {
				if originalValue == newValue || normalizeFoldedScalarValue(originalValue) == normalizeFoldedScalarValue(newValue) {
					replacementLines = info.originalLines
				}
			}

			indicator := info.indicator
			if indicator == "" {
				indicator = ">"
			}
			lines[i] = lineIndent + key + ": " + indicator

			// Insert replacement content
			newLines := make([]string, 0, len(lines)-endIdx+i+1+len(replacementLines))
			newLines = append(newLines, lines[:i+1]...)
			newLines = append(newLines, replacementLines...)
			newLines = append(newLines, lines[endIdx:]...)

			return strings.Join(newLines, "\n")
		}
	}

	return content
}

func formatFoldedScalarLines(valueLines []string, info foldedScalarInfo) []string {
	if len(valueLines) == 0 {
		return nil
	}

	indentPrefix := info.contentIndent
	reindent := func(contentLine string) string {
		if strings.TrimSpace(contentLine) == "" {
			return ""
		}
		return indentPrefix + strings.TrimLeft(contentLine, " \t")
	}

	if len(valueLines) == 1 {
		trimmedValue := strings.TrimSpace(valueLines[0])
		if trimmedValue == "" {
			return []string{""}
		}
		if info.listItems {
			items := splitFoldedListItems(trimmedValue)
			if len(items) > 1 {
				result := make([]string, 0, len(items))
				for _, item := range items {
					result = append(result, indentPrefix+item)
				}
				return result
			}
		}
		return []string{indentPrefix + trimmedValue}
	}

	result := make([]string, 0, len(valueLines))
	for _, line := range valueLines {
		result = append(result, reindent(line))
	}
	return result
}

func splitFoldedListItems(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	var items []string
	var current strings.Builder
	for i := 0; i < len(value); i++ {
		ch := value[i]
		if ch == '-' && (i == 0 || value[i-1] == ' ' || value[i-1] == '\t') {
			if current.Len() > 0 {
				item := strings.TrimSpace(current.String())
				if item != "" {
					items = append(items, item)
				}
				current.Reset()
			}
			current.WriteByte('-')
			continue
		}
		current.WriteByte(ch)
	}

	if current.Len() > 0 {
		item := strings.TrimSpace(current.String())
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func extractFoldedIndicator(line string) string {
	colonPos := strings.Index(line, ":")
	if colonPos < 0 {
		return ">"
	}
	valuePart := strings.TrimSpace(line[colonPos+1:])
	if valuePart == "" {
		return ">"
	}
	fields := strings.Fields(valuePart)
	if len(fields) == 0 {
		return ">"
	}
	indicator := fields[0]
	if !strings.HasPrefix(indicator, ">") {
		return ">"
	}
	return indicator
}

func stripInlineComment(value string) string {
	if value == "" {
		return value
	}
	if idx := strings.Index(value, " #"); idx >= 0 {
		return strings.TrimSpace(value[:idx])
	}
	return value
}

func foldScalarValueFromLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	var result strings.Builder
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if result.Len() > 0 && !strings.HasSuffix(result.String(), "\n") {
				result.WriteString("\n")
			}
			continue
		}
		if result.Len() > 0 && !strings.HasSuffix(result.String(), "\n") {
			result.WriteString(" ")
		}
		result.WriteString(trimmed)
	}
	return result.String()
}

func normalizeFoldedScalarValue(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, " ")
}
