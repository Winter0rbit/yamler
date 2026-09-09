package yamler

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

const apiDoc = `# Service config
app:
  name: myapp       # name
  debug: true
  ports: [80, 443]

  servers:
    - web1   # primary
    - web2
    - web3

database:
  host: localhost
  port: 5432
  pools:
    - name: primary
      size: 10
    - name: replica
      size: 5

features: {auth: true, metrics: false}
`

func mustLoad(t *testing.T, s string) *Document {
	t.Helper()
	d, err := Load(s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestHas(t *testing.T) {
	d := mustLoad(t, apiDoc)
	for _, p := range []string{"app", "app.name", "app.ports[1]", "database.pools[1].size", "features.auth", ""} {
		if !d.Has(p) {
			t.Errorf("Has(%q) = false", p)
		}
	}
	for _, p := range []string{"nope", "app.nope", "app.ports[2]", "app.name.x", "database.pools[5].size"} {
		if d.Has(p) {
			t.Errorf("Has(%q) = true", p)
		}
	}
	arr := mustLoad(t, "- name: a\n- name: b\n")
	if !arr.Has("[1].name") || arr.Has("[2]") || arr.Has("name") {
		t.Error("Has on array-root document")
	}
}

func TestKeys(t *testing.T) {
	d := mustLoad(t, apiDoc)
	got, err := d.Keys("")
	if err != nil || !reflect.DeepEqual(got, []string{"app", "database", "features"}) {
		t.Errorf("Keys(\"\") = %v, %v", got, err)
	}
	got, err = d.Keys("app")
	if err != nil || !reflect.DeepEqual(got, []string{"name", "debug", "ports", "servers"}) {
		t.Errorf("Keys(app) = %v, %v", got, err)
	}
	got, err = d.Keys("database.pools[0]")
	if err != nil || !reflect.DeepEqual(got, []string{"name", "size"}) {
		t.Errorf("Keys(database.pools[0]) = %v, %v", got, err)
	}
	if _, err := d.Keys("app.ports"); !errors.Is(err, ErrType) {
		t.Errorf("Keys on array: %v", err)
	}
	if _, err := d.Keys("nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Keys on missing: %v", err)
	}
}

func TestCopy(t *testing.T) {
	d := mustLoad(t, apiDoc)
	c := d.Copy()
	if err := c.Set("app.name", "copy"); err != nil {
		t.Fatal(err)
	}
	if err := c.AppendToArray("app.ports", 8080); err != nil {
		t.Fatal(err)
	}
	if v, _ := d.GetString("app.name"); v != "myapp" {
		t.Errorf("original changed: %q", v)
	}
	if n, _ := d.GetArrayLength("app.ports"); n != 2 {
		t.Errorf("original array changed: %d", n)
	}
	out, _ := d.String()
	if out != apiDoc {
		t.Errorf("original serialization changed:\n%s", out)
	}
	cout, _ := c.String()
	if !strings.Contains(cout, "name: copy       # name") || !strings.Contains(cout, "ports: [80, 443, 8080]") {
		t.Errorf("copy lost formatting:\n%s", cout)
	}

	// Anchors in the copy point at the copy's own nodes.
	a := mustLoad(t, "base: &b\n  x: 1\nchild:\n  <<: *b\n")
	ac := a.Copy()
	_ = ac.Set("base.x", 2)
	if v, _ := a.GetInt("base.x"); v != 1 {
		t.Error("anchor shared between original and copy")
	}
	if s, _ := ac.String(); !strings.Contains(s, "<<: *b") {
		t.Errorf("copy lost alias:\n%s", s)
	}
}

func TestDelete(t *testing.T) {
	cases := []struct {
		name string
		path string
		want string
	}{
		{"scalar key", "app.debug", strings.Replace(apiDoc, "  debug: true\n", "", 1)},
		{"nested map", "database", `# Service config
app:
  name: myapp       # name
  debug: true
  ports: [80, 443]

  servers:
    - web1   # primary
    - web2
    - web3

features: {auth: true, metrics: false}
`},
		{"block array element", "app.servers[1]", strings.Replace(apiDoc, "    - web2\n", "", 1)},
		{"first array element with comment", "app.servers[0]", strings.Replace(apiDoc, "    - web1   # primary\n", "", 1)},
		{"flow array element", "app.ports[0]", strings.Replace(apiDoc, "[80, 443]", "[443]", 1)},
		{"map in array", "database.pools[0]", strings.Replace(apiDoc, "    - name: primary\n      size: 10\n", "", 1)},
		{"key in flow map", "features.metrics", strings.Replace(apiDoc, "{auth: true, metrics: false}", "{auth: true}", 1)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := mustLoad(t, apiDoc)
			if err := d.Delete(c.path); err != nil {
				t.Fatal(err)
			}
			if d.Has(c.path) && !strings.HasSuffix(c.path, "]") {
				t.Errorf("path still present after Delete")
			}
			out, _ := d.String()
			if out != c.want {
				t.Errorf("unexpected output:\n--- got ---\n%s--- want ---\n%s", out, c.want)
			}
		})
	}
}

func TestDeleteErrors(t *testing.T) {
	d := mustLoad(t, apiDoc)
	if err := d.Delete("app.nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing key: %v", err)
	}
	if err := d.Delete("nope.key"); !errors.Is(err, ErrNotFound) {
		t.Errorf("missing parent: %v", err)
	}
	if err := d.Delete("app.ports[9]"); !errors.Is(err, ErrIndex) {
		t.Errorf("index: %v", err)
	}
	if err := d.Delete("app.name[0]"); !errors.Is(err, ErrType) {
		t.Errorf("index into scalar: %v", err)
	}
	if err := d.Delete(""); !errors.Is(err, ErrPath) {
		t.Errorf("empty path: %v", err)
	}
	out, _ := d.String()
	if out != apiDoc {
		t.Errorf("failed deletes changed the document:\n%s", out)
	}
}

func TestDeleteArrayRoot(t *testing.T) {
	d := mustLoad(t, "---\n- name: a\n  x: 1\n- name: b\n- name: c\n")
	if err := d.Delete("[1]"); err != nil {
		t.Fatal(err)
	}
	if err := d.Delete("[0].x"); err != nil {
		t.Fatal(err)
	}
	out, _ := d.String()
	if out != "---\n- name: a\n- name: c\n" {
		t.Errorf("got:\n%s", out)
	}
}

func TestDeleteAll(t *testing.T) {
	d := mustLoad(t, `services:
  web:
    image: nginx
    debug: true
    env:
      - A=1
      - DEBUG=1
      - B=2
  db:
    image: postgres
    debug: false
debug: true
`)
	n, err := d.DeleteAll("**.debug")
	if err != nil || n != 3 {
		t.Fatalf("DeleteAll(**.debug) = %d, %v", n, err)
	}
	n, err = d.DeleteAll("services.web.env[*]")
	if err != nil || n != 3 {
		t.Fatalf("DeleteAll(env[*]) = %d, %v", n, err)
	}
	n, err = d.DeleteAll("nothing.*")
	if err != nil || n != 0 {
		t.Fatalf("DeleteAll(no match) = %d, %v", n, err)
	}
	out, _ := d.String()
	want := `services:
  web:
    image: nginx
    env: []
  db:
    image: postgres
`
	if out != want {
		t.Errorf("unexpected output:\n--- got ---\n%s--- want ---\n%s", out, want)
	}
}

func TestDeleteAllArrayIndicesInOrder(t *testing.T) {
	// Removing several elements of one array must not shift indices mid-way.
	d := mustLoad(t, "items:\n  - a\n  - b\n  - c\n  - d\n")
	n, err := d.DeleteAll("items[*]")
	if err != nil || n != 4 {
		t.Fatalf("got %d, %v", n, err)
	}
	if l, _ := d.GetArrayLength("items"); l != 0 {
		t.Errorf("array not emptied: %d", l)
	}
}

func TestSentinelErrors(t *testing.T) {
	d := mustLoad(t, "a:\n  b: text\n  list: [1, 2]\n")
	checks := []struct {
		name string
		err  error
		want error
	}{
		{"missing key", second(d.Get("a.nope")), ErrNotFound},
		{"missing nested", second(d.GetString("x.y.z")), ErrNotFound},
		{"type", second(d.GetInt("a.b")), ErrType},
		{"index", second(d.Get("a.list[5]")), ErrIndex},
		{"bad index", second(d.Get("a.list[x]")), ErrPath},
		{"not array", second(d.GetArrayLength("a.b")), ErrType},
		{"array root op", d.AddArrayElement(1), ErrRoot},
	}
	for _, c := range checks {
		if !errors.Is(c.err, c.want) {
			t.Errorf("%s: %v does not wrap %v", c.name, c.err, c.want)
		}
	}
	_, err := Load("a: [unclosed\n")
	if !errors.Is(err, ErrParse) {
		t.Errorf("parse: %v", err)
	}
	_, err = Load("a: 1\n---\nb: 2\n")
	if !errors.Is(err, ErrMultiDocument) {
		t.Errorf("multi-doc: %v", err)
	}
	_, err = LoadFile("/nonexistent/file.yaml")
	if !errors.Is(err, ErrIO) {
		t.Errorf("io: %v", err)
	}
	rule, _ := LoadSchemaFromString(`{"type":"object","required":["z"]}`)
	if err := d.Validate(rule); !errors.Is(err, ErrValidation) {
		t.Errorf("validation: %v", err)
	}
}

func second[T any](_ T, err error) error { return err }
