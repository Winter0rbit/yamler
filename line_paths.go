// A small line-oriented walker that assigns a key path to every line of a
// block-style YAML text. It is used by the formatting post-processors that
// need to correlate lines of the encoder output with lines of the original
// source without a full parse.

package yamler

import (
	"strconv"
	"strings"
)

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
	// idxPath is like keyPath but includes sequence indices ("a.b[2].c");
	// for an item line without an inline key it names the item ("a.b[2]").
	idxPath string
	// value is the text after "key:" (trimmed), "" if there is no key.
	value string
	// skip is true for lines the walker does not interpret: blank lines,
	// comments, document markers and block-scalar contents.
	skip bool
	// frames is the number of stack frames this line opened.
	frames int
	// keyCol is the column of the key on this line (right of any dashes).
	keyCol int
}

type pathFrame struct {
	indent int
	key    string // "" for a sequence-item frame
	index  int    // item index for a sequence-item frame
	items  int    // number of items seen under this frame
}

// lineWalker computes lineInfo for successive lines.
type lineWalker struct {
	stack         []pathFrame
	rootItems     int  // items seen in a root-level sequence
	blockScalarAt int  // indentation of the key owning an open block scalar, -1 if none
	flowDepth     int  // open [ or { brackets of a multi-line flow collection
	flowQuote     byte // quote character of a string left open at the end of the previous line
	flowToken     bool // whether the next flow line starts a new scalar
	openQuote     byte // quote character of a multi-line quoted scalar (block context)
}

func newLineWalker() *lineWalker {
	return &lineWalker{blockScalarAt: -1}
}

// pathOf renders the stack without indices: keys are joined with dots and
// every sequence level contributes "[]" ("spec.containers[].image").
func (w *lineWalker) pathOf(stack []pathFrame) string {
	var b strings.Builder
	for _, f := range stack {
		if f.key != "" {
			if b.Len() > 0 {
				b.WriteByte('.')
			}
			b.WriteString(f.key)
		} else {
			b.WriteString("[]")
		}
	}
	return b.String()
}

// idxPathOf renders the stack with sequence indices: "a.b[2].c".
func (w *lineWalker) idxPathOf(stack []pathFrame) string {
	var b strings.Builder
	for _, f := range stack {
		if f.key != "" {
			if b.Len() > 0 {
				b.WriteByte('.')
			}
			b.WriteString(f.key)
		} else {
			b.WriteString("[")
			b.WriteString(strconv.Itoa(f.index))
			b.WriteString("]")
		}
	}
	return b.String()
}

// clone returns an independent copy of the walker state.
func (w *lineWalker) clone() *lineWalker {
	c := *w
	c.stack = append([]pathFrame(nil), w.stack...)
	return &c
}

// pushItem opens a sequence-item frame at the given column.
func (w *lineWalker) pushItem(indent int) {
	index := 0
	if n := len(w.stack); n > 0 {
		index = w.stack[n-1].items
		w.stack[n-1].items++
	} else {
		index = w.rootItems
		w.rootItems++
	}
	w.stack = append(w.stack, pathFrame{indent: indent, index: index})
}

// keyColumn returns the column of the key on a line: for "- key: v" item
// lines that is two columns right of the dash.
func keyColumn(li lineInfo) int {
	if li.key != "" {
		return li.keyCol
	}
	return li.indent
}

// parentIsItem reports whether the line just consumed is nested directly in
// a sequence item: a key of a mapping inside an item, or an item of a
// sequence inside an item. Such lines are positioned by the enclosing
// item's dash.
func (w *lineWalker) parentIsItem(li lineInfo) bool {
	parent := len(w.stack) - li.frames - 1
	return li.frames > 0 && parent >= 0 && w.stack[parent].key == ""
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
	if w.openQuote != 0 {
		// Continuation of a multi-line quoted scalar.
		if closesQuote(trimmed, w.openQuote) {
			w.openQuote = 0
		}
		return lineInfo{indent: indent, skip: true}
	}
	if w.flowDepth > 0 {
		// Continuation of a multi-line flow collection.
		if w.flowQuote != 0 || !strings.HasPrefix(trimmed, "#") {
			s := trimmed
			if w.flowQuote == 0 {
				s = stripInlineComment(s)
			}
			var d int
			d, w.flowQuote, w.flowToken = flowBracketBalanceState(s, w.flowQuote, w.flowToken)
			w.flowDepth += d
			if w.flowDepth < 0 {
				w.flowDepth = 0
			}
		}
		return lineInfo{indent: indent, skip: true}
	}

	if trimmed == "" || strings.HasPrefix(trimmed, "#") || trimmed == "---" || trimmed == "..." ||
		strings.HasPrefix(trimmed, "--- ") || strings.HasPrefix(trimmed, "%") ||
		// The ": value" line of an explicit "? key" entry is not interpreted.
		trimmed == "?" || trimmed == ":" || strings.HasPrefix(trimmed, ": ") {
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
		first := true
		for {
			// Each "- " opens a sequence item; "- - x" nests one sequence in
			// another, the inner dash sitting two columns to the right.
			if first {
				info.path = w.pathOf(w.stack)
			}
			w.pushItem(indent)
			info.frames++
			if first {
				info.idxPath = w.idxPathOf(w.stack)
				first = false
			}
			if rest == "-" {
				return info
			}
			// The rest of the line sits right of the dash and any extra
			// spaces after it.
			afterDash := rest[2:]
			rest = strings.TrimSpace(afterDash)
			indent += 2 + (len(afterDash) - len(strings.TrimLeft(afterDash, " ")))
			if rest == "" || strings.HasPrefix(rest, "#") {
				return info
			}
			if rest != "-" && !strings.HasPrefix(rest, "- ") {
				break
			}
		}
	}

	// Leading anchors and tags ("&a |", "!!str x") do not affect structure.
	for rest != "" && (rest[0] == '&' || rest[0] == '!') {
		sp := strings.IndexAny(rest, " \t")
		if sp < 0 {
			break
		}
		rest = strings.TrimLeft(rest[sp:], " \t")
	}

	key, value, ok := splitKeyValue(rest)
	if !ok && strings.HasPrefix(rest, "? ") {
		// Explicit key: the whole rest of the line is the key.
		key, value, ok = strings.TrimSpace(rest[2:]), "", true
		if k, _, quoted := splitKeyValue(key + ": "); quoted {
			key = k
		}
	}
	if !ok {
		// A block scalar without a key: the value of a list item ("- |")
		// or a scalar document root.
		if isBlockScalarIndicator(rest) {
			w.blockScalarAt = info.indent
		}
		if v := stripInlineComment(rest); v != "" && (v[0] == '[' || v[0] == '{') {
			w.openFlow(v)
		} else if v != "" && (v[0] == '"' || v[0] == '\'') && !closesQuote(v[1:], v[0]) {
			w.openQuote = v[0]
		}
		return info
	}

	for len(w.stack) > 0 && w.stack[len(w.stack)-1].indent >= indent {
		w.stack = w.stack[:len(w.stack)-1]
	}
	w.stack = append(w.stack, pathFrame{indent: indent, key: key})
	info.frames++
	info.key = key
	info.keyCol = indent
	info.value = value
	info.keyPath = w.pathOf(w.stack)
	info.idxPath = w.idxPathOf(w.stack)
	if !info.isItem {
		info.path = info.keyPath
	}

	if isBlockScalarIndicator(value) {
		w.blockScalarAt = indent
	}
	if v := stripInlineComment(value); v != "" && (v[0] == '[' || v[0] == '{') {
		w.openFlow(v)
	} else if v != "" && (v[0] == '"' || v[0] == '\'') && !closesQuote(v[1:], v[0]) {
		w.openQuote = v[0]
	}
	return info
}

// closesQuote reports whether s contains the closing quote of a string
// opened with q (single quotes are escaped by doubling, double quotes with
// a backslash).
func closesQuote(s string, q byte) bool {
	for i := 0; i < len(s); i++ {
		if s[i] != q {
			if q == '"' && s[i] == '\\' {
				i++
			}
			continue
		}
		if q == '\'' && i+1 < len(s) && s[i+1] == '\'' {
			i++
			continue
		}
		return true
	}
	return false
}

// openFlow records a flow collection value that may continue on the next
// lines.
func (w *lineWalker) openFlow(v string) {
	d, quote, token := flowBracketBalanceState(v, 0, true)
	if d > 0 {
		w.flowDepth = d
		w.flowQuote = quote
		w.flowToken = token
	}
}

// flowBracketBalance returns the number of flow brackets opened minus the
// number closed in s, ignoring brackets inside quotes.
func flowBracketBalance(s string) int {
	d, _, _ := flowBracketBalanceState(s, 0, true)
	return d
}

// flowBracketBalanceState is flowBracketBalance with the quote and
// token-start state carried across lines (quoted and plain scalars may span
// several lines).
func flowBracketBalanceState(s string, quote byte, tokenStart bool) (int, byte, bool) {
	depth := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case quote != 0:
			if c == quote {
				if quote == '\'' && i+1 < len(s) && s[i+1] == '\'' {
					i++
					continue
				}
				if quote == '"' && i > 0 && s[i-1] == '\\' {
					continue
				}
				quote = 0
			}
			continue
		case c == '"' || c == '\'':
			if tokenStart {
				quote = c
				continue
			}
		case c == '[' || c == '{':
			depth++
		case c == ']' || c == '}':
			depth--
		}
		tokenStart = startsToken(s, i, tokenStart)
	}
	return depth, quote, tokenStart
}

// startsToken updates the "next character starts a scalar" state after
// consuming s[i]: true after an opening bracket, a comma, an indicator
// (": " or "- ") or at the beginning; spaces keep the state.
func startsToken(s string, i int, current bool) bool {
	next := byte(' ')
	if i+1 < len(s) {
		next = s[i+1]
	}
	switch s[i] {
	case '[', '{', ',':
		return true
	case ':', '-':
		if next == ' ' || next == '\t' {
			return true
		}
		return false
	case ' ', '\t':
		return current
	}
	return false
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
		key := s[1:end]
		if q == '\'' {
			key = strings.ReplaceAll(key, "''", "'")
		} else if unquoted, err := strconv.Unquote(`"` + key + `"`); err == nil {
			key = unquoted
		}
		return key, strings.TrimSpace(s[end+2:]), true
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

// inlineCommentIndex returns the index of the "#" that starts an inline
// comment in line, or -1. A "#" inside quotes or not preceded by whitespace
// is not a comment.
func inlineCommentIndex(line string) int {
	var quote byte
	tokenStart := true // a quote only opens a string at the start of a scalar
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case quote != 0:
			if c == quote {
				if quote == '\'' && i+1 < len(line) && line[i+1] == '\'' {
					i++
					continue
				}
				if quote == '"' && i > 0 && line[i-1] == '\\' {
					continue
				}
				quote = 0
			}
			continue
		case c == '"' || c == '\'':
			if tokenStart {
				quote = c
				continue
			}
		case c == '#':
			if i == 0 || line[i-1] == ' ' || line[i-1] == '\t' {
				return i
			}
		}
		tokenStart = startsToken(line, i, tokenStart)
	}
	return -1
}
