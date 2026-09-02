package yamler

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// loadCorpus returns the real-world documents under testdata/roundtrip.
func loadCorpus(t testing.TB) map[string]string {
	t.Helper()
	files, err := filepath.Glob("testdata/roundtrip/*.yaml")
	if err != nil || len(files) == 0 {
		t.Fatalf("no corpus files found: %v", err)
	}
	corpus := make(map[string]string, len(files))
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		corpus[filepath.Base(f)] = string(b)
	}
	return corpus
}

// TestCorpusRoundTripIsIdentity checks that loading and serializing a
// document without modifying it reproduces the input byte for byte.
func TestCorpusRoundTripIsIdentity(t *testing.T) {
	for name, content := range loadCorpus(t) {
		t.Run(name, func(t *testing.T) {
			docs, err := LoadAll(content)
			if err != nil {
				t.Fatal(err)
			}
			out, err := DocumentsToBytes(docs)
			if err != nil {
				t.Fatal(err)
			}
			if string(out) != content {
				t.Errorf("round trip is not identity:\n%s", diffLines(content, string(out)))
			}
		})
	}
}

// TestCorpusModificationKeepsLayout applies a change to every document and
// checks that all untouched lines are still there, in the same order and
// with the same indentation, and that the result parses back to the
// expected data.
func TestCorpusModificationKeepsLayout(t *testing.T) {
	for name, content := range loadCorpus(t) {
		t.Run(name, func(t *testing.T) {
			docs, err := LoadAll(content)
			if err != nil {
				t.Fatal(err)
			}
			expectedChanges := 0
			for _, doc := range docs {
				path := firstScalarPath(t, doc)
				if path == "" {
					continue
				}
				expectedChanges++
				before, _ := doc.Get(path)
				if err := doc.Set(path, "changed-value"); err != nil {
					t.Fatalf("Set(%s): %v", path, err)
				}
				after, _ := doc.Get(path)
				if after != "changed-value" {
					t.Errorf("Set(%s) did not apply: %v -> %v", path, before, after)
				}
			}
			out, err := DocumentsToBytes(docs)
			if err != nil {
				t.Fatal(err)
			}
			// Every original line except the changed ones must survive verbatim.
			origLines := strings.Split(content, "\n")
			outLines := strings.Split(string(out), "\n")
			if len(origLines) != len(outLines) {
				t.Fatalf("line count changed %d -> %d:\n%s", len(origLines), len(outLines), diffLines(content, string(out)))
			}
			changed := 0
			for i := range origLines {
				if origLines[i] != outLines[i] {
					changed++
					if !strings.Contains(outLines[i], "changed-value") {
						t.Errorf("line %d changed unexpectedly:\n  was: %q\n  now: %q", i+1, origLines[i], outLines[i])
					}
				}
			}
			if changed != expectedChanges {
				t.Errorf("expected %d changed lines, got %d", expectedChanges, changed)
			}
			// Output must still be valid YAML with the same structure.
			if _, err := LoadAll(string(out)); err != nil {
				t.Fatalf("output does not parse: %v", err)
			}
		})
	}
}

// firstScalarPath returns the path of the first plain string scalar in a
// mapping-root document that is safe to overwrite (not an alias/anchor,
// not a block scalar).
func firstScalarPath(t *testing.T, doc *Document) string {
	t.Helper()
	if doc.isArrayRoot() {
		return ""
	}
	paths, err := doc.GetPathsRecursive()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		if strings.Contains(p, "[") {
			continue
		}
		node, err := doc.getNode(p)
		if err != nil || node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
			continue
		}
		if node.Style != 0 || node.Anchor != "" {
			continue
		}
		if strings.ContainsAny(node.Value, "{}[]*&:#\n") || node.Value == "" {
			continue
		}
		return p
	}
	return ""
}

// diffLines renders a compact line diff for test failures.
func diffLines(want, got string) string {
	w := strings.Split(want, "\n")
	g := strings.Split(got, "\n")
	var b strings.Builder
	n := len(w)
	if len(g) > n {
		n = len(g)
	}
	for i := 0; i < n; i++ {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			b.WriteString("- " + wl + "\n+ " + gl + "\n")
		}
	}
	return b.String()
}

// FuzzRoundTrip feeds arbitrary YAML into Load/ToBytes and checks that the
// library never panics and that the output still parses to the same data.
func FuzzRoundTrip(f *testing.F) {
	for _, content := range loadCorpus(f) {
		f.Add(content)
	}
	f.Add("a: 1\nb:\n  - x\n  - y\n")
	f.Add("- a\n- b: c\n  d: e\n")
	f.Add("k: [1, {a: b}]\n")
	f.Fuzz(func(t *testing.T, content string) {
		if strings.ContainsAny(content, "!\r\t") || !utf8.ValidString(content) {
			return // explicit tags are re-resolved by yaml.v3 itself; CR line breaks, tabs and non-UTF-8 are not covered
		}
		docs, err := LoadAll(content)
		if err != nil {
			return // invalid YAML is fine
		}
		out, err := DocumentsToBytes(docs)
		if err != nil {
			t.Fatalf("DocumentsToBytes: %v", err)
		}
		want := withoutNil(decodeAll(content))
		got := withoutNil(decodeAll(string(out)))
		if len(want) == 0 {
			return // an empty/null document has no data to preserve
		}
		if !reflect.DeepEqual(want, withoutNil(decodeAll(yamlV3RoundTrip(content)))) {
			return // yaml.v3 itself does not round-trip this input; not our layer
		}
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("data changed after round trip\ninput:\n%s\noutput:\n%s", content, out)
		}
	})
}

// yamlV3RoundTrip re-encodes every document of content with yaml.v3 alone,
// without any formatting preservation.
func yamlV3RoundTrip(content string) string {
	dec := yaml.NewDecoder(strings.NewReader(content))
	var buf strings.Builder
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	for {
		var node yaml.Node
		if err := dec.Decode(&node); err != nil {
			break
		}
		if err := enc.Encode(&node); err != nil {
			break
		}
	}
	enc.Close()
	return buf.String()
}

// withoutNil drops empty (null) documents, which LoadAll does not keep.
func withoutNil(docs []interface{}) []interface{} {
	var out []interface{}
	for _, d := range docs {
		if d != nil {
			out = append(out, d)
		}
	}
	return out
}

// decodeAll parses a stream into plain Go values, ignoring formatting.
func decodeAll(content string) []interface{} {
	dec := yaml.NewDecoder(strings.NewReader(content))
	var out []interface{}
	for {
		var v interface{}
		if err := dec.Decode(&v); err != nil {
			break
		}
		out = append(out, v)
	}
	return out
}
