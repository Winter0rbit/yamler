// Package yamler edits YAML documents while preserving their original
// formatting: comments, indentation, flow/block styles, blank lines and
// document markers.
//
// A Document keeps the parsed yaml.v3 node tree next to the raw source text.
// Every mutation re-encodes the tree and then runs the encoder output through
// a chain of text post-processors (see formatting_apply.go) that restore the
// original look of the file using the FormattingInfo captured at load time.
package yamler

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Pool for reusing byte buffers to reduce allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

// Cache for parsed paths to avoid repeated string splitting
// Bounded to prevent unbounded growth with dynamic paths.
var pathCache = struct {
	sync.RWMutex
	values map[string][]string
}{
	values: make(map[string][]string),
}

const maxPathCacheEntries = 10000

// parsePath splits a path and caches the result
func parsePath(path string) []string {
	pathCache.RLock()
	if cached, ok := pathCache.values[path]; ok {
		pathCache.RUnlock()
		return cached
	}
	pathCache.RUnlock()

	parts := strings.Split(path, ".")

	pathCache.Lock()
	if len(pathCache.values) >= maxPathCacheEntries {
		pathCache.values = make(map[string][]string)
	}
	pathCache.values[path] = parts
	pathCache.Unlock()

	return parts
}

// Document represents a YAML document with preserved formatting
type Document struct {
	root                      *yaml.Node
	raw                       string
	arrayRoot                 bool
	trailingNewlines          int
	preserveDocumentSeparator bool // Whether to preserve document separators for array root documents
	exactTrailingNewlines     bool // Whether to preserve exact trailing newline behavior (from LoadBytes)
	// Performance optimization: cache formatting info
	formattingCache *FormattingInfo
}

// mappingRoot returns the root MappingNode of the document
func (d *Document) mappingRoot() (*yaml.Node, error) {
	if d.root == nil || len(d.root.Content) == 0 {
		return nil, fmt.Errorf("empty document root")
	}
	root := d.root.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("root is not a mapping node")
	}
	return root, nil
}

// LoadFile loads a YAML file and preserves its formatting
func LoadFile(filename string) (*Document, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return LoadBytes(content)
}

// LoadBytes loads a YAML document from a byte slice and preserves its formatting
func LoadBytes(content []byte) (*Document, error) {
	doc, err := Load(string(content))
	if err != nil {
		return nil, err
	}
	// Enable exact trailing newline preservation for file-based loading
	doc.exactTrailingNewlines = true
	return doc, nil
}

// Load parses a YAML string and preserves its formatting
func Load(content string) (*Document, error) {
	if content == "" {
		// Create empty document
		return &Document{
			root: &yaml.Node{
				Kind: yaml.DocumentNode,
				Content: []*yaml.Node{
					{
						Kind: yaml.MappingNode,
						Tag:  "!!map",
					},
				},
			},
		}, nil
	}

	// Count trailing newlines
	trailingNewlines := 0
	for i := len(content) - 1; i >= 0; i-- {
		if content[i] == '\n' {
			trailingNewlines++
		} else if content[i] == '\r' {
			// Skip carriage returns
			continue
		} else {
			break
		}
	}

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(content), &node); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	doc := &Document{
		root:             &node,
		raw:              content,
		trailingNewlines: trailingNewlines,
	}

	// Initialize formatting cache if we have raw content
	if content != "" {
		doc.formattingCache = detectFormattingInfoOptimized(content)
	}

	// Detect if this is an array document root
	if doc.isArrayRoot() {
		doc.arrayRoot = true
	}

	return doc, nil
}

// Save writes the YAML document to a file while preserving formatting
func (d *Document) Save(filename string) error {
	content, err := d.ToBytes()
	if err != nil {
		return err
	}

	return os.WriteFile(filename, content, 0644)
}

// ToBytes converts the document to bytes while preserving formatting
func (d *Document) ToBytes() ([]byte, error) {
	if d.root == nil || len(d.root.Content) == 0 {
		return []byte{}, nil
	}

	// Get buffer from pool to reduce allocations
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2) // Always use 2 spaces for encoding

	// Preserve original node styles before encoding
	if d.raw != "" {
		// Use cached formatting info or detect if not cached
		var info *FormattingInfo
		if d.formattingCache != nil {
			info = d.formattingCache
		} else {
			info = detectFormattingInfoOptimized(d.raw)
			d.formattingCache = info // Cache for future use
		}

		preserveNodeStylesWithInfo(d.root, info, "")
		// Apply zero-indent arrays formatting to nodes before encoding
		applyZeroIndentToNodes(d.root, info, "")
	}

	if err := encoder.Encode(d.root); err != nil {
		return nil, err
	}
	encoder.Close()

	// Make a copy of the buffer contents
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	// If we have raw content, apply formatting preservation
	if d.raw != "" {
		// Use cached formatting info (already computed above)
		var indentInfo *FormattingInfo
		if d.formattingCache != nil {
			indentInfo = d.formattingCache
		} else {
			indentInfo = detectFormattingInfoOptimized(d.raw)
			d.formattingCache = indentInfo // Cache for future use
		}

		// Post-process to maintain original style characteristics
		result = preserveOriginalFormatting(result, d.raw, indentInfo, d.preserveDocumentSeparator)
	}

	// Remove any trailing newlines that might have been added by the encoder
	for len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	// Add the correct number of trailing newlines
	// Logic to handle different scenarios:
	// 1. If original was empty - no trailing newlines
	// 2. If original had trailing newlines - preserve exact count
	// 3. If exactTrailingNewlines is enabled - preserve exactly (for file operations)
	// 4. Otherwise - add one trailing newline (YAML standard)
	var finalTrailingNewlines int
	if d.raw == "" {
		// Empty document - no trailing newlines
		finalTrailingNewlines = 0
	} else if d.trailingNewlines > 0 {
		// Original had trailing newlines - preserve exact count
		finalTrailingNewlines = d.trailingNewlines
	} else if d.exactTrailingNewlines {
		// Exact preservation mode (typically for file operations) - preserve exactly
		finalTrailingNewlines = 0
	} else {
		// No trailing newlines in original - add one (YAML standard)
		finalTrailingNewlines = 1
	}

	if finalTrailingNewlines > 0 {
		// Pre-allocate with exact size needed
		finalResult := make([]byte, len(result)+finalTrailingNewlines)
		copy(finalResult, result)
		for i := len(result); i < len(finalResult); i++ {
			finalResult[i] = '\n'
		}
		return finalResult, nil
	} else {
		// No trailing newlines needed
		return result, nil
	}
}

// String returns the YAML document as a string, preserving formatting.
func (d *Document) String() (string, error) {
	bytes, err := d.ToBytes()
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
