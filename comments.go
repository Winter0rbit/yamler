// Inline comment alignment: detection results are applied in
// alignInlineComments, and the Document methods configure the mode.

package yamler

import (
	"strings"
)

// alignInlineComments aligns inline comments according to the specified mode
func alignInlineComments(content string, info *FormattingInfo) string {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments-only lines
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Look for lines with inline comments
		if commentPos := strings.Index(line, "#"); commentPos >= 0 {
			// Extract the part before the comment
			beforeComment := line[:commentPos]
			comment := line[commentPos:]

			// Check if this line has a key that should be aligned
			if colonPos := strings.Index(beforeComment, ":"); colonPos >= 0 {
				key := strings.TrimSpace(beforeComment[:colonPos])

				switch info.AlignmentMode {
				case CommentAlignmentDisabled:
					// Remove comments entirely
					lines[i] = strings.TrimRight(beforeComment, " ")

				case CommentAlignmentRelative:
					// Relative alignment: preserve original spacing
					if alignmentValue, exists := info.CommentAlignment[key]; exists && alignmentValue > 0 {
						// Extract the value part after colon
						valueStart := colonPos + 1
						if valueStart < len(beforeComment) {
							valuePart := beforeComment[valueStart:]
							trimmedValue := strings.TrimSpace(valuePart)

							// Reconstruct with exact spacing
							keyPart := beforeComment[:colonPos] // Just up to colon
							if len(trimmedValue) > 0 {
								padding := strings.Repeat(" ", alignmentValue)
								lines[i] = keyPart + ": " + trimmedValue + padding + comment
							}
						}
					}

				case CommentAlignmentAbsolute:
					// Absolute alignment: align to specific column
					targetColumn := info.CommentSpacing
					if targetColumn > 0 {
						// Remove trailing spaces from before comment
						beforeComment = strings.TrimRight(beforeComment, " ")

						// Add spaces to reach target column
						spacesNeeded := targetColumn - len(beforeComment)
						if spacesNeeded > 0 {
							padding := strings.Repeat(" ", spacesNeeded)
							lines[i] = beforeComment + padding + comment
						} else {
							// If we can't fit, use at least one space
							lines[i] = beforeComment + " " + comment
						}
					}
				}
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
