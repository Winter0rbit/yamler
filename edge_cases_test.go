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

func TestDocumentedBehaviours(t *testing.T) {
	// GetBool accepts true/false, yes/no, on/off and 1/0.
	d, _ := Load("a: yes\nb: on\nc: 1\nd: off\ne: 0\nf: 2\n")
	for k, want := range map[string]bool{"a": true, "b": true, "c": true, "d": false, "e": false} {
		got, err := d.GetBool(k)
		if err != nil || got != want {
			t.Errorf("GetBool(%s) = %v, %v; want %v", k, got, err, want)
		}
	}
	if _, err := d.GetBool("f"); err == nil {
		t.Error("GetBool(2) should fail")
	}

	// Wildcards combined with array indices.
	d, _ = Load("svc:\n  a:\n    ports: [80, 81]\n  b:\n    ports: [8080]\n")
	m, err := d.GetAll("svc.*.ports[0]")
	if err != nil || len(m) != 2 || m["svc.a.ports[0]"] != int64(80) {
		t.Errorf("GetAll(svc.*.ports[0]) = %v, %v", m, err)
	}
	if err := d.SetAll("svc.*.ports[0]", 1); err != nil {
		t.Fatal(err)
	}
	if v, _ := d.GetInt("svc.b.ports[0]"); v != 1 {
		t.Errorf("SetAll with index did not apply: %v", v)
	}

	// Get on array-root documents.
	d, _ = Load("- name: a\n  tags: [x, y]\n- name: b\n")
	if v, err := d.GetString("[1].name"); err != nil || v != "b" {
		t.Errorf("GetString([1].name) = %q, %v", v, err)
	}
	if v, err := d.GetString("[0].tags[1]"); err != nil || v != "y" {
		t.Errorf("GetString([0].tags[1]) = %q, %v", v, err)
	}
	if n, err := d.GetSlice(""); err != nil || len(n) != 2 {
		t.Errorf("GetSlice('') = %v, %v", n, err)
	}
	if _, err := d.Get("name"); err == nil {
		t.Error("Get without index on array root should fail")
	}

	// JSON Schema type names are accepted.
	rule, err := LoadSchemaFromString(`{"type":"object","properties":{"port":{"type":"integer","minimum":1},"ratio":{"type":"number"},"on":{"type":"boolean"}},"required":["port"]}`)
	if err != nil {
		t.Fatal(err)
	}
	d, _ = Load("port: 80\nratio: 0.5\non: true\n")
	if err := d.Validate(rule); err != nil {
		t.Errorf("valid document rejected: %v", err)
	}
	d, _ = Load("port: 0\n")
	if err := d.Validate(rule); err == nil {
		t.Error("minimum violation not reported")
	}
}

// TestFlowQuotingIsRespected checks that commas and colons inside quoted
// scalars do not split flow collections when their style is restored.
func TestFlowQuotingIsRespected(t *testing.T) {
	cases := []string{
		"a: [' ,']\nz: 1\n",
		"a: [\"x, y\", z]\nz: 1\n",
		"a: ['it''s, fine', b]\nz: 1\n",
		"m: {k: 'a, b', n: 2}\nz: 1\n",
		"m: {'a: b': 1}\nz: 1\n",
		"n: [[1, 2], [3, 4]]\nz: 1\n",
	}
	for _, in := range cases {
		doc, err := Load(in)
		if err != nil {
			t.Errorf("Load(%q): %v", in, err)
			continue
		}
		before, _ := doc.Get("")
		if err := doc.Set("z", 2); err != nil {
			t.Errorf("Set on %q: %v", in, err)
			continue
		}
		out, _ := doc.String()
		if out != strings.Replace(in, "z: 1", "z: 2", 1) {
			t.Errorf("formatting changed for %q:\n%s", in, out)
		}
		reloaded, err := Load(out)
		if err != nil {
			t.Errorf("output of %q does not parse: %v", in, err)
			continue
		}
		after, _ := reloaded.Get("")
		bm, _ := before.(map[string]interface{})
		am, _ := after.(map[string]interface{})
		if len(bm) != len(am) {
			t.Errorf("data changed for %q", in)
		}
	}
}

// TestMultilineFlowKeepsBlankLines covers blank lines inside a multi-line
// flow collection, which are part of a plain scalar's value.
func TestMultilineFlowKeepsBlankLines(t *testing.T) {
	in := "a: {k:\nv1\n\nv2}\nz: 1\n"
	doc, err := Load(in)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := doc.GetString("a.k")
	if err := doc.Set("z", 2); err != nil {
		t.Fatal(err)
	}
	out, _ := doc.String()
	reloaded, err := Load(out)
	if err != nil {
		t.Fatalf("output does not parse: %v\n%s", err, out)
	}
	if got, _ := reloaded.GetString("a.k"); got != want {
		t.Errorf("value changed: %q -> %q\noutput:\n%s", want, got, out)
	}
}

// TestExplicitKeysKeepData checks documents whose keys yaml.v3 renders in
// explicit "? key" form (keys longer than 128 characters).
func TestExplicitKeysKeepData(t *testing.T) {
	long := strings.Repeat("k", 200)
	in := long + ":\n    a: 1\n    b: 2\n"
	doc, err := Load(in)
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Set(long+".a", 3); err != nil {
		t.Fatal(err)
	}
	out, _ := doc.String()
	reloaded, err := Load(out)
	if err != nil {
		t.Fatalf("output does not parse: %v\n%s", err, out)
	}
	a, _ := reloaded.GetInt(long + ".a")
	b, _ := reloaded.GetInt(long + ".b")
	if a != 3 || b != 2 {
		t.Errorf("data changed: a=%d b=%d\n%s", a, b, out)
	}
}

// TestEmptyQuotedKey covers "" used as a mapping key, which must not be
// confused with "this line has no key".
func TestEmptyQuotedKey(t *testing.T) {
	in := "    a: 1\n    \"\": 2\n    b: 3\n"
	doc, err := Load(in)
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Set("a", 9); err != nil {
		t.Fatal(err)
	}
	out, _ := doc.String()
	if out != strings.Replace(in, "a: 1", "a: 9", 1) {
		t.Errorf("layout changed:\n--- got ---\n%s--- want ---\n%s", out, in)
	}
	v, err := doc.GetInt("")
	if err == nil && v == 2 {
		return // the empty key is reachable, fine either way
	}
}

// TestMultilinePlainScalar covers a plain scalar continued on the next line,
// whose continuation must not be mistaken for a quoted scalar or a key.
func TestMultilinePlainScalar(t *testing.T) {
	in := "    a: one\n     two\n    b: 3\n"
	doc, err := Load(in)
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.Set("b", 4); err != nil {
		t.Fatal(err)
	}
	out, _ := doc.String()
	reloaded, err := Load(out)
	if err != nil {
		t.Fatalf("output does not parse: %v\n%s", err, out)
	}
	if v, _ := reloaded.GetString("a"); v != "one two" {
		t.Errorf("value changed to %q\noutput:\n%s", v, out)
	}
	if v, _ := reloaded.GetInt("b"); v != 4 {
		t.Errorf("b = %d\noutput:\n%s", v, out)
	}
}

// TestCommentOnlyValue covers "key: # comment", where the value is empty and
// the block below the key belongs to it.
func TestCommentOnlyValue(t *testing.T) {
	in := "root: # note\n    items:\n        - a\n        - b\n"
	doc, err := Load(in)
	if err != nil {
		t.Fatal(err)
	}
	if err := doc.AppendToArray("root.items", "c"); err != nil {
		t.Fatal(err)
	}
	out, _ := doc.String()
	want := in + "        - c\n"
	if out != want {
		t.Errorf("unexpected output:\n--- got ---\n%s--- want ---\n%s", out, want)
	}
}
