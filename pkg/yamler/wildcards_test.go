package yamler

import (
	"sort"
	"testing"
)

func TestDocument_GetAll(t *testing.T) {
	yamlContent := `
app:
  name: myapp
  version: 1.0
  settings:
    debug: true
    timeout: 30
    database:
      host: localhost
      port: 5432
config:
  development:
    debug: true
    name: dev-app
  production:
    debug: false
    name: prod-app
servers:
  - name: server1
    host: host1
  - name: server2
    host: host2
`

	doc, err := Load(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load document: %v", err)
	}

	tests := []struct {
		name     string
		pattern  string
		expected map[string]interface{}
	}{
		{
			name:    "single wildcard",
			pattern: "app.*",
			expected: map[string]interface{}{
				"app.name":    "myapp",
				"app.version": float64(1.0),
				"app.settings": map[string]interface{}{
					"debug":   true,
					"timeout": int64(30),
					"database": map[string]interface{}{
						"host": "localhost",
						"port": int64(5432),
					},
				},
			},
		},
		{
			name:    "recursive wildcard",
			pattern: "**.debug",
			expected: map[string]interface{}{
				"app.settings.debug":       true,
				"config.development.debug": true,
				"config.production.debug":  false,
			},
		},
		{
			name:    "nested single wildcard",
			pattern: "config.*.name",
			expected: map[string]interface{}{
				"config.development.name": "dev-app",
				"config.production.name":  "prod-app",
			},
		},
		// {
		// 	name:    "array wildcard",
		// 	pattern: "servers[*]",
		// 	expected: map[string]interface{}{
		// 		"servers[0]": map[string]interface{}{
		// 			"name": "server1",
		// 			"host": "host1",
		// 		},
		// 		"servers[1]": map[string]interface{}{
		// 			"name": "server2",
		// 			"host": "host2",
		// 		},
		// 	},
		// },
		{
			name:    "exact match",
			pattern: "app.name",
			expected: map[string]interface{}{
				"app.name": "myapp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := doc.GetAll(tt.pattern)
			if err != nil {
				t.Fatalf("GetAll() error = %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Errorf("GetAll() returned %d results, expected %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				if !exists {
					t.Errorf("Expected key %s not found in result", key)
					continue
				}

				if !deepEqual(actualValue, expectedValue) {
					t.Errorf("GetAll() for key %s = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestDocument_SetAll(t *testing.T) {
	yamlContent := `
config:
  development:
    debug: true
    timeout: 30
  production:
    debug: false
    timeout: 60
`

	tests := []struct {
		name     string
		pattern  string
		value    interface{}
		expected string
	}{
		{
			name:    "set all debug values",
			pattern: "config.*.debug",
			value:   "enabled",
			expected: `config:
  development:
    debug: enabled
    timeout: 30
  production:
    debug: enabled
    timeout: 60
`,
		},
		{
			name:    "set all timeout values",
			pattern: "config.*.timeout",
			value:   int64(45),
			expected: `config:
  development:
    debug: true
    timeout: 45
  production:
    debug: false
    timeout: 45
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(yamlContent)
			if err != nil {
				t.Fatalf("Failed to load document: %v", err)
			}

			err = doc.SetAll(tt.pattern, tt.value)
			if err != nil {
				t.Fatalf("SetAll() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("Failed to convert to string: %v", err)
			}

			if result != tt.expected {
				t.Errorf("SetAll() result mismatch\nGot:\n%s\nWant:\n%s", result, tt.expected)
			}
		})
	}
}

func TestDocument_GetKeys(t *testing.T) {
	yamlContent := `
app:
  name: myapp
  version: 1.0
  settings:
    debug: true
config:
  development:
    debug: true
  production:
    debug: false
`

	doc, err := Load(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load document: %v", err)
	}

	tests := []struct {
		name     string
		pattern  string
		expected []string
	}{
		{
			name:    "get all debug keys",
			pattern: "**.debug",
			expected: []string{
				"app.settings.debug",
				"config.development.debug",
				"config.production.debug",
			},
		},
		{
			name:    "get app keys",
			pattern: "app.*",
			expected: []string{
				"app.name",
				"app.settings",
				"app.version",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := doc.GetKeys(tt.pattern)
			if err != nil {
				t.Fatalf("GetKeys() error = %v", err)
			}

			sort.Strings(result)
			sort.Strings(tt.expected)

			if len(result) != len(tt.expected) {
				t.Errorf("GetKeys() returned %d keys, expected %d", len(result), len(tt.expected))
			}

			for i, key := range result {
				if i >= len(tt.expected) || key != tt.expected[i] {
					t.Errorf("GetKeys() key %d = %s, want %s", i, key, tt.expected[i])
				}
			}
		})
	}
}

func TestDocument_GetPathsRecursive(t *testing.T) {
	yamlContent := `
app:
  name: myapp
  settings:
    debug: true
config:
  timeout: 30
`

	doc, err := Load(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load document: %v", err)
	}

	paths, err := doc.GetPathsRecursive()
	if err != nil {
		t.Fatalf("GetPathsRecursive() error = %v", err)
	}

	expected := []string{
		"app",
		"app.name",
		"app.settings",
		"app.settings.debug",
		"config",
		"config.timeout",
	}

	if len(paths) != len(expected) {
		t.Errorf("GetPathsRecursive() returned %d paths, expected %d", len(paths), len(expected))
	}

	for i, path := range paths {
		if i >= len(expected) || path != expected[i] {
			t.Errorf("GetPathsRecursive() path %d = %s, want %s", i, path, expected[i])
		}
	}
}

func TestWildcardPatternMatching(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"app.name", "app.name", true},
		{"app.name", "app.*", true},
		{"app.settings.debug", "app.*", false},
		{"app.settings.debug", "app.**", true},
		{"app.settings.debug", "**.debug", true},
		{"config.development.debug", "**.debug", true},
		{"servers[0].name", "servers.*", false}, // array index doesn't match *
		{"servers[0].name", "servers.*.name", false},
		{"app.name", "config.*", false},
		{"very.deep.nested.value", "**.value", true},
		{"very.deep.nested.value", "very.**.value", true},
		{"very.deep.nested.value", "very.deep.*", false},
		{"very.deep.nested.value", "very.deep.**", true},
	}

	for _, tt := range tests {
		t.Run(tt.path+"_matches_"+tt.pattern, func(t *testing.T) {
			got := pathMatches(tt.path, tt.pattern)
			if got != tt.want {
				t.Errorf("pathMatches(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestFilterByPattern(t *testing.T) {
	data := map[string]interface{}{
		"app.name":                 "myapp",
		"app.version":              "1.0",
		"app.settings.debug":       true,
		"config.development.debug": true,
		"config.production.debug":  false,
		"servers[0].name":          "server1",
	}

	tests := []struct {
		pattern  string
		expected map[string]interface{}
	}{
		{
			pattern: "app.*",
			expected: map[string]interface{}{
				"app.name":    "myapp",
				"app.version": "1.0",
			},
		},
		{
			pattern: "**.debug",
			expected: map[string]interface{}{
				"app.settings.debug":       true,
				"config.development.debug": true,
				"config.production.debug":  false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			result := FilterByPattern(data, tt.pattern)

			if len(result) != len(tt.expected) {
				t.Errorf("FilterByPattern() returned %d items, expected %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				actualValue, exists := result[key]
				if !exists {
					t.Errorf("Expected key %s not found in result", key)
					continue
				}

				if !deepEqual(actualValue, expectedValue) {
					t.Errorf("FilterByPattern() for key %s = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}
