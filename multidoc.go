// Multi-document YAML streams ("---"-separated documents).

package yamler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// countDocuments returns the number of non-empty documents in a YAML stream.
// A trailing empty document (e.g. a file ending with "---") is not counted.
func countDocuments(content string) (int, error) {
	dec := yaml.NewDecoder(strings.NewReader(content))
	count := 0
	for {
		var node yaml.Node
		err := dec.Decode(&node)
		if errors.Is(err, io.EOF) {
			return count, nil
		}
		if err != nil {
			return count, err
		}
		if len(node.Content) == 0 {
			continue
		}
		if n := node.Content[0]; n.Kind == yaml.ScalarNode && n.Tag == "!!null" {
			continue
		}
		count++
	}
}

// isDocumentStartLine reports whether a line is a "---" document marker,
// optionally followed by a comment or a tag.
func isDocumentStartLine(line string) bool {
	line = strings.TrimRight(line, "\r\n")
	if !strings.HasPrefix(line, "---") {
		return false
	}
	return len(line) == 3 || line[3] == ' ' || line[3] == '\t'
}

// splitDocuments splits a YAML stream into the raw text of its documents.
// Each document after the first keeps its leading "---" line so that the
// separator is preserved when the documents are serialized again.
func splitDocuments(content string) []string {
	lines := strings.SplitAfter(content, "\n")
	var docs []string
	var cur strings.Builder
	inBlockScalar := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if inBlockScalar {
			// A block scalar ends at the first non-empty line that is not
			// indented; only such lines can be document markers.
			if trimmed != "" && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				inBlockScalar = false
			}
		}
		if !inBlockScalar && isDocumentStartLine(line) && cur.Len() > 0 {
			docs = append(docs, cur.String())
			cur.Reset()
		}
		if !inBlockScalar && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			v := stripInlineComment(strings.TrimPrefix(trimmed, "- "))
			if i := strings.LastIndex(v, ":"); i >= 0 {
				v = strings.TrimSpace(v[i+1:])
			}
			if v == "|" || v == ">" || strings.HasPrefix(v, "|") || strings.HasPrefix(v, ">") {
				if strings.Trim(v, "|>+-0123456789") == "" {
					inBlockScalar = true
				}
			}
		}
		cur.WriteString(line)
	}
	if cur.Len() > 0 || len(docs) == 0 {
		docs = append(docs, cur.String())
	}
	return docs
}

// LoadAll parses a YAML stream that may contain several documents separated
// by "---" and returns one Document per non-empty document. Each Document
// preserves the formatting of its own part of the stream; use
// DocumentsToBytes or SaveAll to serialize them back together.
func LoadAll(content string) ([]*Document, error) {
	if strings.Trim(content, " \t\r\n") == "" {
		return nil, nil
	}
	var docs []*Document
	for _, raw := range splitDocuments(content) {
		n, err := countDocuments(raw)
		if err != nil {
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}
		if n == 0 {
			// Attach an empty trailing document ("---\n") to the previous one
			// so the separator survives a round trip.
			if len(docs) > 0 {
				prev := docs[len(docs)-1]
				prev.trailingRaw += raw
			}
			continue
		}
		doc, err := Load(raw)
		if err != nil {
			return nil, err
		}
		doc.exactTrailingNewlines = true
		docs = append(docs, doc)
	}
	return docs, nil
}

// LoadAllBytes is LoadAll for a byte slice.
func LoadAllBytes(content []byte) ([]*Document, error) {
	return LoadAll(string(content))
}

// LoadAllFile loads every document of a multi-document YAML file.
func LoadAllFile(filename string) ([]*Document, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return LoadAllBytes(content)
}

// DocumentsToBytes serializes documents loaded with LoadAll back into one
// stream. Documents after the first are separated by "---" unless they
// already carry their own separator.
func DocumentsToBytes(docs []*Document) ([]byte, error) {
	var buf bytes.Buffer
	for i, doc := range docs {
		content, err := doc.ToBytes()
		if err != nil {
			return nil, fmt.Errorf("document %d: %w", i, err)
		}
		if i > 0 {
			if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != '\n' {
				buf.WriteByte('\n')
			}
			if !isDocumentStartLine(firstLine(content)) {
				buf.WriteString("---\n")
			}
		}
		buf.Write(content)
		buf.WriteString(doc.trailingRaw)
	}
	return buf.Bytes(), nil
}

// SaveAll writes a multi-document stream to a file.
func SaveAll(filename string, docs []*Document) error {
	content, err := DocumentsToBytes(docs)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, content, 0o644)
}

func firstLine(b []byte) string {
	if i := bytes.IndexByte(b, '\n'); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}
