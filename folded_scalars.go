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

// preserveFoldedScalars restores the original line layout of folded (>)
// scalars, which the encoder re-folds onto a single line.
func preserveFoldedScalars(newContent, original string, info *FormattingInfo) string {
	if len(info.ScalarStyles) == 0 {
		return newContent
	}
	originalLines := strings.Split(original, "\n")
	foldedInfo := make(map[string]foldedScalarInfo)

	w := newLineWalker()
	for i, line := range originalLines {
		li := w.next(line)
		if li.skip || li.key == "" || info.ScalarStyles[li.keyPath] != yaml.FoldedStyle {
			continue
		}
		if !strings.HasPrefix(stripInlineComment(li.value), ">") {
			continue
		}
		indent := keyColumn(li)
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
			if getLineIndentation(nextLine) <= indent {
				break
			}
			foldedLines = append(foldedLines, nextLine)
			if !foundContent {
				contentIndent = leadingWhitespace(nextLine)
				foundContent = true
			}
			if !strings.HasPrefix(strings.TrimSpace(nextLine), "-") {
				listItems = false
			}
		}
		for len(foldedLines) > 0 && foldedLines[len(foldedLines)-1] == "" {
			foldedLines = foldedLines[:len(foldedLines)-1]
		}
		if !foundContent {
			contentIndent = strings.Repeat(" ", indent+info.IndentSize)
			listItems = false
		}
		foldedInfo[li.keyPath] = foldedScalarInfo{
			key:           li.key,
			keyIndent:     indent,
			contentIndent: contentIndent,
			indicator:     extractFoldedIndicator(li.value),
			originalLines: foldedLines,
			listItems:     listItems,
		}
	}
	if len(foldedInfo) == 0 {
		return newContent
	}

	return replaceFoldedScalars(newContent, foldedInfo)
}

// replaceFoldedScalars rewrites every folded scalar in content whose path
// is in foldedInfo using the original layout.
func replaceFoldedScalars(content string, foldedInfo map[string]foldedScalarInfo) string {
	lines := strings.Split(content, "\n")
	var out []string
	w := newLineWalker()
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		li := w.next(line)
		if li.skip || li.key == "" {
			out = append(out, line)
			continue
		}
		info, ok := foldedInfo[li.keyPath]
		if !ok {
			out = append(out, line)
			continue
		}
		vs := keyValueStart(line, li)
		if vs < 0 {
			out = append(out, line)
			continue
		}
		indent := keyColumn(li)
		valuePart := stripInlineComment(strings.TrimSpace(line[vs:]))
		hasFoldedIndicator := strings.HasPrefix(valuePart, ">")

		// Collect the value the encoder produced.
		var newValueLines []string
		end := i
		if hasFoldedIndicator {
			for j := i + 1; j < len(lines); j++ {
				if strings.TrimSpace(lines[j]) == "" {
					newValueLines = append(newValueLines, "")
					end = j
					continue
				}
				if getLineIndentation(lines[j]) <= indent {
					break
				}
				newValueLines = append(newValueLines, lines[j])
				end = j
			}
			// The encoder may leave blank lines after the block; drop them
			// from the value but keep them consumed.
			for len(newValueLines) > 0 && newValueLines[len(newValueLines)-1] == "" {
				newValueLines = newValueLines[:len(newValueLines)-1]
			}
			// Feed the skipped block to the walker so its state stays in sync.
			for j := i + 1; j <= end; j++ {
				w.next(lines[j])
			}
		} else if valuePart != "" && valuePart != `""` && valuePart != "''" {
			if valuePart[0] == '"' || valuePart[0] == '\'' {
				// The encoder chose a quoted string (e.g. trailing spaces
				// that cannot be folded). Only restore the original block
				// if it denotes exactly the same value.
				var decoded string
				if err := yaml.Unmarshal([]byte(valuePart), &decoded); err != nil || decoded != foldedBlockValue(info) {
					out = append(out, line)
					continue
				}
				out = append(out, line[:vs]+" "+foldedIndicatorOrDefault(info))
				out = append(out, info.originalLines...)
				continue
			} else {
				newValueLines = append(newValueLines, valuePart)
			}
		}

		replacement := formatFoldedScalarLines(newValueLines, info)
		if len(info.originalLines) > 0 {
			originalValue := foldScalarValueFromLines(info.originalLines)
			newValue := foldScalarValueFromLines(newValueLines)
			if originalValue == newValue || normalizeFoldedScalarValue(originalValue) == normalizeFoldedScalarValue(newValue) {
				replacement = info.originalLines
			}
		}
		out = append(out, line[:vs]+" "+foldedIndicatorOrDefault(info))
		out = append(out, replacement...)
		i = end
	}
	return strings.Join(out, "\n")
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

func extractFoldedIndicator(value string) string {
	valuePart := strings.TrimSpace(value)
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
	if idx := inlineCommentIndex(value); idx > 0 {
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

func foldedIndicatorOrDefault(info foldedScalarInfo) string {
	if info.indicator == "" {
		return ">"
	}
	return info.indicator
}

// foldedBlockValue decodes the original folded block (indicator plus
// content lines) with yaml.v3 to obtain its exact value.
func foldedBlockValue(info foldedScalarInfo) string {
	indicator := info.indicator
	if indicator == "" {
		indicator = ">"
	}
	var decoded string
	src := indicator + "\n" + strings.Join(info.originalLines, "\n") + "\n"
	if err := yaml.Unmarshal([]byte(src), &decoded); err != nil {
		return "\x00undecodable"
	}
	return decoded
}

func normalizeFoldedScalarValue(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	return strings.Join(fields, " ")
}
