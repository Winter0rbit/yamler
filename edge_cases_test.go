package yamler

import (
	"strings"
	"testing"
)

func TestMergeKeyIsNotTagged(t *testing.T) {
	input := `x-common: &common
  restart: always
  logging:
    driver: json-file

services:
  web:
    <<: *common
    image: nginx
  db:
    <<: *common
    image: postgres
`
	doc, err := Load(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Set("services.web.image", "nginx:1.25"); err != nil {
		t.Fatal(err)
	}
	out, err := doc.String()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "!!merge") {
		t.Fatalf("merge key must be emitted without an explicit tag, got:\n%s", out)
	}
	expected := strings.Replace(input, "image: nginx\n", "image: nginx:1.25\n", 1)
	if out != expected {
		t.Fatalf("unexpected output:\n--- got ---\n%s--- want ---\n%s", out, expected)
	}
}

func TestLoadRejectsMultiDocumentStream(t *testing.T) {
	_, err := Load("a: 1\n---\nb: 2\n")
	if err == nil {
		t.Fatal("expected an error for a multi-document stream")
	}
	if !strings.Contains(err.Error(), "LoadAll") {
		t.Fatalf("error should point to LoadAll, got: %v", err)
	}

	// A single document with a leading separator and a trailing empty
	// document is still a single document.
	for _, in := range []string{"---\na: 1\n", "a: 1\n---\n", "---\na: 1\n...\n"} {
		if _, err := Load(in); err != nil {
			t.Fatalf("Load(%q) unexpectedly failed: %v", in, err)
		}
	}
}

func TestLoadAllRoundTrip(t *testing.T) {
	input := `# Service
apiVersion: v1
kind: Service
metadata:
  name: web   # service name
---
# Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 3
`
	docs, err := LoadAll(input)
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(docs))
	}

	// Untouched stream round-trips byte for byte.
	out, err := DocumentsToBytes(docs)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != input {
		t.Fatalf("round trip changed the stream:\n--- got ---\n%s--- want ---\n%s", out, input)
	}

	// Modifications stay local to their document.
	if err := docs[1].Set("spec.replicas", 5); err != nil {
		t.Fatal(err)
	}
	kind, _ := docs[0].GetString("kind")
	if kind != "Service" {
		t.Fatalf("first document changed: kind=%q", kind)
	}
	out, err = DocumentsToBytes(docs)
	if err != nil {
		t.Fatal(err)
	}
	expected := strings.Replace(input, "replicas: 3", "replicas: 5", 1)
	if string(out) != expected {
		t.Fatalf("unexpected output:\n--- got ---\n%s--- want ---\n%s", out, expected)
	}
}

func TestLoadAllSingleDocument(t *testing.T) {
	docs, err := LoadAll("a: 1\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
	out, _ := DocumentsToBytes(docs)
	if string(out) != "a: 1\n" {
		t.Fatalf("got %q", out)
	}
}

func TestZeroIndentArraysPreserved(t *testing.T) {
	input := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: nginx
        image: nginx:1.20
        ports:
        - containerPort: 80
        env:
        - name: A
          value: "1"
      volumes:
      - name: data
        emptyDir: {}
`
	doc, err := Load(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Set("spec.replicas", 5); err != nil {
		t.Fatal(err)
	}
	out, _ := doc.String()
	expected := strings.Replace(input, "replicas: 3", "replicas: 5", 1)
	if out != expected {
		t.Fatalf("zero-indent arrays not preserved:\n--- got ---\n%s--- want ---\n%s", out, expected)
	}

	// Appending keeps the style too.
	if err := doc.AppendToArray("spec.template.spec.containers[0].ports", map[string]interface{}{"containerPort": 443}); err != nil {
		t.Fatal(err)
	}
	out, _ = doc.String()
	if !strings.Contains(out, "        - containerPort: 80\n        - containerPort: 443\n") {
		t.Fatalf("appended element should use zero-indent style:\n%s", out)
	}
}

func TestZeroIndentGitHubActions(t *testing.T) {
	input := `name: CI
on:
  push:
    branches: [main]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    - name: Test
      run: go test ./...
`
	doc, err := Load(input)
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Set("jobs.test.runs-on", "ubuntu-22.04"); err != nil {
		t.Fatal(err)
	}
	out, _ := doc.String()
	expected := strings.Replace(input, "ubuntu-latest", "ubuntu-22.04", 1)
	if out != expected {
		t.Fatalf("unexpected output:\n--- got ---\n%s--- want ---\n%s", out, expected)
	}
}

func TestMixedIndentArraysNotForcedToZero(t *testing.T) {
	// Normal indented arrays must stay indented when the document does not
	// use zero-indent style.
	input := `a:
  - 1
  - 2
b:
  c:
    - x
`
	doc, _ := Load(input)
	_ = doc.Set("b.d", 1)
	out, _ := doc.String()
	if !strings.HasPrefix(out, "a:\n  - 1\n  - 2\nb:\n  c:\n    - x\n") {
		t.Fatalf("indented arrays changed:\n%s", out)
	}
}

func TestSplitDocumentsEdgeCases(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"a: 1\n", 1},
		{"---\na: 1\n---\nb: 2\n", 2},
		{"a: |\n  ---\n  not a marker\n---\nb: 2\n", 2},
		{"steps:\n- run: |\n    ---\n    text\n---\nb: 2\n", 2},
		{"a: 1\n---\n", 1},
		{"a: 1\n---\n---\nb: 2\n", 2},
	}
	for _, c := range cases {
		docs, err := LoadAll(c.in)
		if err != nil {
			t.Errorf("LoadAll(%q): %v", c.in, err)
			continue
		}
		if len(docs) != c.want {
			t.Errorf("LoadAll(%q): got %d documents, want %d", c.in, len(docs), c.want)
			continue
		}
		out, err := DocumentsToBytes(docs)
		if err != nil {
			t.Errorf("DocumentsToBytes(%q): %v", c.in, err)
			continue
		}
		if string(out) != c.in {
			t.Errorf("round trip of %q produced %q", c.in, out)
		}
	}
}
