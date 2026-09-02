// A small line-oriented walker that assigns a key path to every line of a
// block-style YAML text. It is used by the formatting post-processors that
// need to correlate lines of the encoder output with lines of the original
// source without a full parse.

package yamler

import "strings"

// lineInfo describes one line of YAML text as seen by the walker.
type lineInfo struct {
	indent int    // number of leading spaces
	isItem bool   // line starts with "- " (or is a bare "-")
	key    string // mapping key on this line, "" if none
	// path is the dotted path of the mapping key on this line, or, for a
	// sequence item, the path of the sequence it belongs to. Array indices
	// are not part of the path, so every item of a sequence shares the
	// path of its parent key.
	path string
	// keyPath is the path of the mapping key on this line. It differs from
	// path only for "- key: value" item lines.
	keyPath string
	// skip is true for lines the walker does not interpret: blank lines,
	// comments, document markers and block-scalar contents.
	skip bool
}

type pathFrame struct {
	indent int
	key    string // "" for a sequence-item frame
}

// lineWalker computes lineInfo for successive lines.
type lineWalker struct {
	stack         []pathFrame
	blockScalarAt int // indentation of the key owning an open block scalar, -1 if none
}

func newLineWalker() *lineWalker {
	return &lineWalker{blockScalarAt: -1}
}

func (w *lineWalker) pathOf(stack []pathFrame) string {
	var parts []string
	for _, f := range stack {
		if f.key != "" {
			parts = append(parts, f.key)
		}
	}
	return strings.Join(parts, ".")
}

// next consumes one line and returns its description.
func (w *lineWalker) next(line string) lineInfo {
	indent := getLineIndentation(line)
	trimmed := strings.TrimSpace(line)

	if w.blockScalarAt >= 0 {
		if trimmed == "" || indent > w.blockScalarAt {
			return lineInfo{indent: indent, skip: true}
		}
		w.blockScalarAt = -1
	}

	if trimmed == "" || strings.HasPrefix(trimmed, "#") || trimmed == "---" || trimmed == "..." ||
		strings.HasPrefix(trimmed, "--- ") || strings.HasPrefix(trimmed, "%") {
		return lineInfo{indent: indent, skip: true}
	}

	info := lineInfo{indent: indent}
	rest := trimmed
	if rest == "-" || strings.HasPrefix(rest, "- ") {
		info.isItem = true
		// Close previous items of the same sequence and anything nested deeper.
		for len(w.stack) > 0 {
			top := w.stack[len(w.stack)-1]
			if top.indent > indent || (top.indent == indent && top.key == "") {
				w.stack = w.stack[:len(w.stack)-1]
				continue
			}
			break
		}
		info.path = w.pathOf(w.stack)
		w.stack = append(w.stack, pathFrame{indent: indent})
		if rest == "-" {
			return info
		}
		// The rest of the line belongs to a mapping nested in the item and
		// sits two columns to the right of the dash.
		rest = strings.TrimSpace(rest[2:])
		indent += 2
		if rest == "" || strings.HasPrefix(rest, "#") {
			return info
		}
	}

	key, value, ok := splitKeyValue(rest)
	if !ok {
		return info
	}

	for len(w.stack) > 0 && w.stack[len(w.stack)-1].indent >= indent {
		w.stack = w.stack[:len(w.stack)-1]
	}
	w.stack = append(w.stack, pathFrame{indent: indent, key: key})
	info.key = key
	info.keyPath = w.pathOf(w.stack)
	if !info.isItem {
		info.path = info.keyPath
	}

	if isBlockScalarIndicator(value) {
		w.blockScalarAt = indent
	}
	return info
}

// splitKeyValue splits "key: value" into its parts. It understands quoted
// keys and ignores colons that are not followed by whitespace or EOL.
func splitKeyValue(s string) (key, value string, ok bool) {
	if s == "" {
		return "", "", false
	}
	if s[0] == '"' || s[0] == '\'' {
		q := s[0]
		end := -1
		for i := 1; i < len(s); i++ {
			if s[i] == q {
				if q == '\'' && i+1 < len(s) && s[i+1] == '\'' {
					i++
					continue
				}
				if q == '"' && s[i-1] == '\\' {
					continue
				}
				end = i
				break
			}
		}
		if end < 0 || end+1 >= len(s) || s[end+1] != ':' {
			return "", "", false
		}
		if end+2 < len(s) && s[end+2] != ' ' && s[end+2] != '\t' {
			return "", "", false
		}
		return s[1:end], strings.TrimSpace(s[end+2:]), true
	}
	if s[0] == '[' || s[0] == '{' || s[0] == '&' || s[0] == '*' || s[0] == '!' || s[0] == '|' || s[0] == '>' {
		return "", "", false
	}
	for i := 0; i < len(s); i++ {
		if s[i] == ':' && (i+1 == len(s) || s[i+1] == ' ' || s[i+1] == '\t') {
			return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:]), true
		}
		if s[i] == '#' && i > 0 && (s[i-1] == ' ' || s[i-1] == '\t') {
			return "", "", false
		}
	}
	return "", "", false
}

// isBlockScalarIndicator reports whether a value starts a literal (|) or
// folded (>) block scalar, allowing chomping/indentation indicators and a
// trailing comment.
func isBlockScalarIndicator(value string) bool {
	value = stripInlineComment(value)
	if value == "" || (value[0] != '|' && value[0] != '>') {
		return false
	}
	return strings.Trim(value[1:], "+-0123456789") == ""
}
