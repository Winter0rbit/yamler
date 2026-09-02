// Inline comment alignment: detection results are applied in
// alignInlineComments, and the Document methods configure the mode.

package yamler

import "strings"

// alignInlineComments applies the configured comment alignment mode to
// every line that carries an inline comment.
func alignInlineComments(content string, info *FormattingInfo) string {
	lines := strings.Split(content, "\n")
	w := newLineWalker()
	for i, line := range lines {
		li := w.next(line)
		if li.skip {
			continue
		}
		commentPos := inlineCommentIndex(line)
		if commentPos <= 0 {
			continue
		}
		before := strings.TrimRight(line[:commentPos], " ")
		comment := line[commentPos:]
		if strings.TrimSpace(before) == "" {
			continue
		}

		switch info.AlignmentMode {
		case CommentAlignmentDisabled:
			lines[i] = before
		case CommentAlignmentRelative:
			if spaces, ok := info.CommentAlignment[li.idxPath]; ok && spaces > 0 {
				lines[i] = before + strings.Repeat(" ", spaces) + comment
			}
		case CommentAlignmentAbsolute:
			if target := info.CommentSpacing; target > 0 {
				spaces := target - len(before)
				if spaces < 1 {
					spaces = 1
				}
				lines[i] = before + strings.Repeat(" ", spaces) + comment
			}
		}
	}
	return strings.Join(lines, "\n")
}

// SetCommentAlignment configures how inline comments should be aligned
func (d *Document) SetCommentAlignment(mode CommentAlignmentMode) {
	if d.formattingCache == nil {
		d.formattingCache = detectFormattingInfoOptimized(d.raw)
	}
	d.formattingCache.AlignmentMode = mode
}

// SetAbsoluteCommentAlignment aligns all comments to the specified column
func (d *Document) SetAbsoluteCommentAlignment(column int) {
	if d.formattingCache == nil {
		d.formattingCache = detectFormattingInfoOptimized(d.raw)
	}
	d.formattingCache.AlignmentMode = CommentAlignmentAbsolute
	d.formattingCache.CommentSpacing = column
}

// EnableRelativeCommentAlignment preserves original spacing between values and comments
func (d *Document) EnableRelativeCommentAlignment() {
	if d.formattingCache == nil {
		d.formattingCache = detectFormattingInfoOptimized(d.raw)
	}
	d.formattingCache.AlignmentMode = CommentAlignmentRelative
}

// DisableCommentAlignment disables all comment alignment processing
func (d *Document) DisableCommentAlignment() {
	if d.formattingCache == nil {
		d.formattingCache = detectFormattingInfoOptimized(d.raw)
	}
	d.formattingCache.AlignmentMode = CommentAlignmentDisabled
}
