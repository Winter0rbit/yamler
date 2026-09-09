// Types describing the formatting captured from the original YAML source.

package yamler

import (
	"gopkg.in/yaml.v3"
)

// CommentAlignmentMode defines how inline comments should be aligned
type CommentAlignmentMode int

const (
	// CommentAlignmentRelative preserves original spacing between value and comment
	CommentAlignmentRelative CommentAlignmentMode = iota
	// CommentAlignmentAbsolute aligns all comments to the same column
	CommentAlignmentAbsolute
	// CommentAlignmentDisabled disables comment alignment processing
	CommentAlignmentDisabled
)

// ArrayStyle represents different array formatting styles
type ArrayStyle struct {
	IsFlow      bool // true for [1,2,3], false for block style
	IsMultiline bool // true for multiline flow arrays
	HasSpaces   bool // true for [ 1 , 2 , 3 ] (spaces around elements)
	IsCompact   bool // true for [1,2,3] (no spaces)
	Indentation int  // custom indentation level
}

// FormattingInfo holds information about the original YAML formatting
type FormattingInfo struct {
	IndentSize       int
	UseTabs          bool
	EmptyLines       map[string]int         // Number of empty lines before each key
	FlowStyles       map[string]bool        // Nodes that should remain in flow style
	ScalarStyles     map[string]yaml.Style  // Preserve literal/folded scalars
	MultilineFlow    map[string]bool        // Multiline flow objects
	ZeroIndentArrays map[string]bool        // Arrays that start without additional indentation
	HasDocumentStart bool                   // Whether the original had "---"
	HasDocumentEnd   bool                   // Whether the original had "..."
	CommentAlignment map[string]int         // Spacing or column position for inline comments
	CommentSpacing   int                    // Common spacing for comment alignment
	AlignmentMode    CommentAlignmentMode   // How to align comments
	ArrayStyles      map[string]*ArrayStyle // Array formatting styles
	KeyIndents       map[string]int         // Exact indentation for each key
	FlowObjectStyles map[string]string      // Original flow object strings to preserve exact formatting
	CommentIndents   map[string]int         // Indentation of standalone comment lines by text (-1 if ambiguous)
	Structural       bool                   // Whether the document root is a mapping or a sequence
}

// clone returns a deep copy of the formatting snapshot.
func (info *FormattingInfo) clone() *FormattingInfo {
	c := *info
	c.EmptyLines = copyMap(info.EmptyLines)
	c.FlowStyles = copyMap(info.FlowStyles)
	c.ScalarStyles = copyMap(info.ScalarStyles)
	c.MultilineFlow = copyMap(info.MultilineFlow)
	c.ZeroIndentArrays = copyMap(info.ZeroIndentArrays)
	c.CommentAlignment = copyMap(info.CommentAlignment)
	c.KeyIndents = copyMap(info.KeyIndents)
	c.FlowObjectStyles = copyMap(info.FlowObjectStyles)
	c.CommentIndents = copyMap(info.CommentIndents)
	c.ArrayStyles = make(map[string]*ArrayStyle, len(info.ArrayStyles))
	for k, v := range info.ArrayStyles {
		s := *v
		c.ArrayStyles[k] = &s
	}
	return &c
}

func copyMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return nil
	}
	c := make(map[K]V, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}
