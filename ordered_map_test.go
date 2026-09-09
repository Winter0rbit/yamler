package yamler

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestOrderedMapKeepsKeyOrder(t *testing.T) {
	doc := mustLoad(t, "app:\n  name: x\n")
	err := doc.Set("database", OrderedMap{
		Keys: []string{"host", "port", "name"},
		Values: map[string]interface{}{
			"name": "app",
			"host": "localhost",
			"port": int64(5432),
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, _ := doc.String()
	want := "app:\n  name: x\ndatabase:\n  host: localhost\n  port: 5432\n  name: app\n"
	if out != want {
		t.Errorf("unexpected output:\n--- got ---\n%s--- want ---\n%s", out, want)
	}

	// A plain map is written with sorted keys, an OrderedMap is not.
	plain := mustLoad(t, "a: 1\n")
	_ = plain.Set("m", map[string]interface{}{"z": 1, "a": 2})
	if s, _ := plain.String(); !strings.Contains(s, "  a: 2\n  z: 1\n") {
		t.Errorf("plain map should be sorted:\n%s", s)
	}
}

func TestOrderedMapNested(t *testing.T) {
	doc := mustLoad(t, "root: 1\n")
	err := doc.Set("servers", []interface{}{
		OrderedMap{
			Keys: []string{"name", "config"},
			Values: map[string]interface{}{
				"name": "web",
				"config": OrderedMap{
					Keys:   []string{"port", "tls"},
					Values: map[string]interface{}{"port": int64(443), "tls": true},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out, _ := doc.String()
	want := "root: 1\nservers:\n  - name: web\n    config:\n      port: 443\n      tls: true\n"
	if out != want {
		t.Errorf("unexpected output:\n--- got ---\n%s--- want ---\n%s", out, want)
	}
}

func TestOrderedMapMarshalYAML(t *testing.T) {
	om := OrderedMap{
		Keys:   []string{"b", "a"},
		Values: map[string]interface{}{"a": int64(1), "b": int64(2)},
	}
	out, err := yaml.Marshal(om)
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != "b: 2\na: 1\n" {
		t.Errorf("yaml.Marshal(OrderedMap) = %q", out)
	}
}

func TestOrderedMapMissingValue(t *testing.T) {
	doc := mustLoad(t, "a: 1\n")
	if err := doc.Set("m", OrderedMap{Keys: []string{"x", "y"}, Values: map[string]interface{}{"x": "set"}}); err != nil {
		t.Fatal(err)
	}
	out, _ := doc.String()
	if out != "a: 1\nm:\n  x: set\n  y:\n" {
		t.Errorf("got:\n%s", out)
	}
}
