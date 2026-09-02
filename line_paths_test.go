package yamler

import (
	"strings"
	"testing"
)

func TestLineWalkerPaths(t *testing.T) {
	input := `apiVersion: v1
spec:
  containers:
  - name: nginx
    image: nginx
    ports:
    - containerPort: 80
    script: |
      - not: an item
      key: value
  volumes:
    - name: data
      "quoted.key": 1
  # comment
  other: [1, 2]
top: x
`
	want := []string{
		"apiVersion",
		"spec",
		"spec.containers",
		"spec.containers", // - name (item of containers, key name)
		"spec.containers[].image",
		"spec.containers[].ports",
		"spec.containers[].ports", // - containerPort
		"spec.containers[].script",
		"", "", // block scalar content
		"spec.volumes",
		"spec.volumes", // - name
		"spec.volumes[].quoted.key",
		"", // comment
		"spec.other",
		"top",
	}
	w := newLineWalker()
	lines := strings.Split(strings.TrimSuffix(input, "\n"), "\n")
	for i, line := range lines {
		li := w.next(line)
		got := li.path
		if li.skip {
			got = ""
		}
		if got != want[i] {
			t.Errorf("line %d %q: path %q, want %q", i, line, got, want[i])
		}
	}
	// Item lines carry the key found after the dash.
	w = newLineWalker()
	li := w.next("- name: x")
	if !li.isItem || li.key != "name" || li.path != "" || li.keyPath != "[].name" {
		t.Errorf("root item: %+v", li)
	}
}

func TestLineWalkerIdxPaths(t *testing.T) {
	input := `a:
  - x
  - y: 1
    z: 2
b:
- p: 1
`
	want := []string{"a", "a[0]", "a[1].y", "a[1].z", "b", "b[0].p"}
	w := newLineWalker()
	for i, line := range strings.Split(strings.TrimSuffix(input, "\n"), "\n") {
		if li := w.next(line); li.idxPath != want[i] {
			t.Errorf("line %q: idxPath %q, want %q", line, li.idxPath, want[i])
		}
	}
	w = newLineWalker()
	for i, line := range []string{"- a", "- b: 1", "  c: 2", "- d"} {
		li := w.next(line)
		want := []string{"[0]", "[1].b", "[1].c", "[3]"}[i]
		if i == 3 {
			want = "[2]"
		}
		if li.idxPath != want {
			t.Errorf("root item %q: idxPath %q, want %q", line, li.idxPath, want)
		}
	}
}

func TestInlineCommentIndex(t *testing.T) {
	cases := map[string]int{
		`a: 1 # c`:         5,
		`a: "#fff" # c`:    10,
		`a: 'it''s' # c`:   11,
		`a: x#y`:           -1,
		`# only`:           0,
		`url: http://x/#f`: -1,
		`- item   # c`:     9,
	}
	for line, want := range cases {
		if got := inlineCommentIndex(line); got != want {
			t.Errorf("%q: got %d want %d", line, got, want)
		}
	}
}
