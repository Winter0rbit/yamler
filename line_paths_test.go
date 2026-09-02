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
		"spec.containers.image",
		"spec.containers.ports",
		"spec.containers.ports", // - containerPort
		"spec.containers.script",
		"", "", // block scalar content
		"spec.volumes",
		"spec.volumes", // - name
		"spec.volumes.quoted.key",
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
	if !li.isItem || li.key != "name" || li.path != "" || li.keyPath != "name" {
		t.Errorf("root item: %+v", li)
	}
}
