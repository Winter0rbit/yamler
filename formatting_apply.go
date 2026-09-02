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

	// Preserve folded scalar formatting (drops blank lines the encoder adds
	// after a block, so it must run before blank lines are restored)
	newStr = preserveFoldedScalars(newStr, original, info)

	// Apply empty line patterns (after indentation to avoid conflicts)
	newStr = applyEmptyLinePatterns(newStr, info)

	// Align inline comments
	newStr = alignInlineComments(newStr, info)

	// Restore document separators
	newStr = restoreDocumentSeparators(newStr, info, original, preserveDocumentSeparator)

	// Final cleanup: ensure empty lines are truly empty (no indentation)
	// but preserve original empty lines with indentation
	newStr = cleanupEmptyLines(newStr, original)

	return []byte(newStr)
}

// cleanupEmptyLines makes whitespace-only lines truly empty, except for
// lines that were whitespace-only in the original at the same position
// (kept verbatim) and whitespace-only lines inside block scalars, which are
// content.
func cleanupEmptyLines(content, original string) string {
	lines := strings.Split(content, "\n")
	originalLines := strings.Split(original, "\n")
	w := newLineWalker()
	structural := false // whether the last non-blank line was a key or item
	for i, line := range lines {
		li := w.next(line)
		if !isBlankLine(line) {
			structural = !li.skip && (li.key != "" || li.isItem)
			continue
		}
		if li.skip && w.blockScalarAt >= 0 {
			continue // block scalar content: keep what the encoder produced
		}
		if structural && i < len(originalLines) && isBlankLine(originalLines[i]) && originalLines[i] != "" && !strings.ContainsAny(originalLines[i], "\t\r") {
			lines[i] = originalLines[i]
		} else {
			lines[i] = ""
		}
	}
	return strings.Join(lines, "\n")
}

// isBlankLine reports whether a line contains only YAML whitespace.
func isBlankLine(line string) bool {
	return strings.Trim(line, " \t\r") == ""
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

// convertToCustomIndentation re-indents the encoder output (which always
// uses two spaces) to the document's indentation size. Structure is taken
// from the line walker so that mapping keys nested in list items keep
// their "dash plus two" offset and block-scalar contents move with their
// key.
func convertToCustomIndentation(content string, targetIndentSize int) string {
	if targetIndentSize == 2 {
		return content
	}

	lines := strings.Split(content, "\n")
	w := newLineWalker()
	type level struct {
		frame    pathFrame
		old, new int
	}
	var levels []level // one entry per walker frame
	lastDelta := 0
	blockOwner := -1 // index in lines of the key owning an open block scalar
	blockShift := 0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		li := w.next(line)
		if li.skip {
			if blockOwner >= 0 && w.blockScalarAt >= 0 {
				lines[i] = strings.Repeat(" ", li.indent+blockShift) + strings.TrimLeft(line, " ")
			} else if strings.HasPrefix(trimmed, "#") {
				lines[i] = strings.Repeat(" ", max(li.indent+lastDelta, 0)) + trimmed
			}
			continue
		}
		blockOwner = -1
		// Keep levels in sync with the walker's stack: drop frames the
		// walker popped, then add the frames it pushed for this line.
		for k := range levels {
			if k >= len(w.stack) || levels[k].frame.indent != w.stack[k].indent || levels[k].frame.key != w.stack[k].key || levels[k].frame.index != w.stack[k].index {
				levels = levels[:k]
				break
			}
		}
		// Frames pushed for this line: the item frame (if any) and the key frame (if any).
		pushed := len(w.stack) - len(levels)
		for k := len(w.stack) - pushed; k < len(w.stack); k++ {
			fr := w.stack[k]
			var parent *level
			if k > 0 {
				parent = &levels[k-1]
			}
			newIndent := 0
			switch {
			case parent == nil:
				newIndent = fr.indent
			case fr.key == "":
				// A list item under a key: encoder puts it at key+2.
				newIndent = parent.new + (fr.indent - parent.old)
				if w.stack[k-1].key != "" {
					newIndent = parent.new + targetIndentSize*(fr.indent-parent.old)/2
				}
			case w.stack[k-1].key == "":
				// Key inside a list item keeps its dash offset.
				newIndent = parent.new + (fr.indent - parent.old)
			default:
				newIndent = parent.new + targetIndentSize*(fr.indent-parent.old)/2
			}
			levels = append(levels, level{frame: fr, old: fr.indent, new: newIndent})
		}
		// The line's own indentation is that of its outermost pushed frame.
		if pushed > 0 {
			own := levels[len(levels)-pushed]
			lastDelta = own.new - own.old
			lines[i] = strings.Repeat(" ", own.new) + strings.TrimLeft(line, " ")
			if w.blockScalarAt >= 0 {
				blockOwner = i
				keyNew := levels[len(levels)-1].new
				blockShift = keyNew + targetIndentSize - (levels[len(levels)-1].old + 2)
				if strings.ContainsAny(stripInlineComment(li.value), "123456789") {
					// Explicit indentation indicator: content keeps its
					// offset from the key.
					blockShift = keyNew - levels[len(levels)-1].old
				}
			}
		}
	}

	return strings.Join(lines, "\n")
}

// applyEmptyLinePatterns re-inserts the blank lines that preceded keys,
// items and comment lines in the original document.
func applyEmptyLinePatterns(content string, info *FormattingInfo) string {
	if len(info.EmptyLines) == 0 {
		return content
	}
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines)+len(info.EmptyLines))
	w := newLineWalker()
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			result = append(result, line)
			continue
		}
		li := w.next(line)
		var key string
		switch {
		case li.skip && strings.HasPrefix(trimmed, "#"):
			key = trimmed
		case !li.skip && (li.key != "" || li.isItem):
			key = li.idxPath
		}
		if n := info.EmptyLines[key]; n > 0 && i > 0 && strings.TrimSpace(lines[i-1]) != "" {
			for j := 0; j < n; j++ {
				result = append(result, "")
			}
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// applyExactIndentations moves every key and sequence item that existed in
// the original document back to its original column. Lines that were not
// in the original (added by modifications), block-scalar contents and
// comments move together with the structure they belong to.
//
// Columns are computed in a first pass over the unmodified text, so the
// walker always sees a consistent structure, and written in a second pass.
func applyExactIndentations(content string, info *FormattingInfo) string {
	if len(info.KeyIndents) == 0 && len(info.CommentIndents) == 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	target := make([]int, len(lines))
	w := newLineWalker()
	// Per stack frame: the column the frame ended up at and its offset from
	// the recorded column (or the parent's offset when nothing was recorded).
	type frameInfo struct {
		col      int // column the frame ended up at
		delta    int // col minus the recorded column (for children with records)
		shift    int // col minus the column in the encoder output (for children without)
		childCol int // column of the first child key or item placed under it, -1 if none
	}
	var frames []frameInfo
	lastDelta := 0

	for i, line := range lines {
		blockBefore := w.blockScalarAt
		li := w.next(line)
		target[i] = li.indent
		if li.skip {
			trimmed := strings.TrimSpace(line)
			switch {
			case trimmed == "":
			case strings.HasPrefix(trimmed, "#") && w.blockScalarAt < 0:
				if want, ok := info.CommentIndents[trimmed]; ok && want >= 0 && (blockBefore < 0 || want <= blockBefore) {
					// A comment right after a block scalar must not be
					// indented into it.
					target[i] = want
				} else {
					target[i] = max(li.indent+lastDelta, 0)
				}
			default:
				target[i] = max(li.indent+lastDelta, 0)
			}
			continue
		}
		first := len(w.stack) - li.frames
		if first < 0 {
			first = 0
		}
		if len(frames) > first {
			frames = frames[:first]
		}
		parentDelta, parentShift, parentCol := 0, 0, -1
		if first > 0 && first-1 < len(frames) {
			parentDelta = frames[first-1].delta
			parentShift = frames[first-1].shift
			parentCol = frames[first-1].col
		}

		var recorded int
		var ok bool
		switch {
		case li.isItem:
			recorded, ok = info.KeyIndents[li.path+"[]"]
		case li.key != "":
			recorded, ok = info.KeyIndents[li.keyPath]
		}

		want := li.indent + parentShift
		if ok && !w.parentIsItem(li) {
			want = recorded + parentDelta
		}
		// Siblings share a column: follow the first child placed under the
		// same parent.
		if first > 0 && first-1 < len(frames) && frames[first-1].childCol >= 0 && (li.key != "" || li.isItem) {
			want = frames[first-1].childCol
		}
		// Never leave the parent: a key sits right of its parent key, an
		// item at or right of it.
		if parentCol >= 0 && (li.key != "" || li.isItem) {
			minCol := parentCol + 1
			if li.isItem {
				minCol = parentCol
			}
			if want < minCol {
				want = minCol
			}
		}
		if want < 0 {
			want = 0
		}
		target[i] = want
		lastDelta = want - li.indent
		if first > 0 && first-1 < len(frames) && frames[first-1].childCol < 0 && (li.key != "" || li.isItem) {
			frames[first-1].childCol = want
		}

		delta := parentDelta
		if ok {
			delta = want - recorded
		}
		shift := want - li.indent
		for len(frames) < len(w.stack) {
			frames = append(frames, frameInfo{col: want, delta: delta, shift: shift, childCol: -1})
		}
		if li.isItem && li.key != "" {
			// "- key:" lines: the inline key has its own column and record.
			keyCol := want + (li.keyCol - li.indent)
			keyDelta := delta
			if rec, ok := info.KeyIndents[li.keyPath]; ok {
				keyDelta = keyCol - rec
			}
			frames[len(frames)-1] = frameInfo{col: keyCol, delta: keyDelta, shift: shift, childCol: -1}
			// The inline key is the first child of the item frame.
			frames[len(frames)-2].childCol = keyCol
		}
	}

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if cur := getLineIndentation(line); cur != target[i] {
			lines[i] = strings.Repeat(" ", target[i]) + strings.TrimLeft(line, " ")
		}
	}
	return strings.Join(lines, "\n")
}

// shiftBlock changes the indentation of lines[start] and of every following
// line that is nested under it (indented deeper than blockIndent) by delta
// columns. Blank lines are left untouched.
func shiftBlock(lines []string, start, blockIndent, delta int) {
	startTrimmed := strings.TrimSpace(lines[start])
	startIsKey := !(startTrimmed == "-" || strings.HasPrefix(startTrimmed, "- "))
	for j := start; j < len(lines); j++ {
		trimmed := strings.TrimSpace(lines[j])
		if trimmed == "" {
			continue
		}
		indent := getLineIndentation(lines[j])
		if j > start && indent <= blockIndent {
			// Zero-indent list items sit at the column of their key but
			// still belong to its block.
			isItem := trimmed == "-" || strings.HasPrefix(trimmed, "- ")
			if !(startIsKey && isItem && indent == blockIndent) {
				break
			}
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
		keyCol := keyColumn(li)
		if (first != "-" && !strings.HasPrefix(first, "- ")) || itemIndent <= keyCol {
			continue
		}
		shift := itemIndent - keyCol
		// Shift the whole sequence block (items and their nested content).
		// A scratch walker tells block-scalar and flow continuation lines,
		// which may be indented less than the items, apart from the end of
		// the block.
		scratch := w.clone()
		for ; j < len(lines); j++ {
			trimmed := strings.TrimSpace(lines[j])
			if trimmed == "" {
				continue
			}
			sl := scratch.next(lines[j])
			if sl.indent < itemIndent && !(sl.skip && !strings.HasPrefix(trimmed, "#")) {
				break
			}
			if sl.indent >= shift {
				lines[j] = lines[j][shift:]
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

// lastLineIs reports whether the last non-empty line of content is exactly marker.
func lastLineIs(content, marker string) bool {
	content = strings.TrimRight(content, " \t\r\n")
	if i := strings.LastIndexByte(content, '\n'); i >= 0 {
		content = content[i+1:]
	}
	return strings.TrimSpace(content) == marker
}
