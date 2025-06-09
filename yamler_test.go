package yamler

import (
	"os"
	"strings"
	"testing"
)

func TestYAMLPreservation(t *testing.T) {
	// Test YAML content with comments and formatting
	yamlContent := `# Database configuration
db:
  host: localhost  # Default host
  port: 5432      # Default PostgreSQL port
  
  # User credentials
  credentials:
    username: admin
    password: secret  # Change in production!

# Application settings
app:
  name: MyApp
  # Environment specific settings
  env: development`

	// Write test file
	tmpFile := "test.yaml"
	if err := os.WriteFile(tmpFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Load the file
	doc, err := LoadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load YAML file: %v", err)
	}

	// Test getting values
	tests := []struct {
		path     string
		expected interface{}
	}{
		{"db.host", "localhost"},
		{"db.port", int64(5432)},
		{"db.credentials.username", "admin"},
		{"app.name", "MyApp"},
		{"app.env", "development"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			value, err := doc.Get(tt.path)
			if err != nil {
				t.Errorf("Failed to get %s: %v", tt.path, err)
				return
			}
			if value != tt.expected {
				t.Errorf("Got %v (%T), want %v (%T)", value, value, tt.expected, tt.expected)
			}
		})
	}

	// Test setting values
	if err := doc.Set("db.host", "newhost"); err != nil {
		t.Errorf("Failed to set db.host: %v", err)
	}

	if err := doc.Set("app.env", "production"); err != nil {
		t.Errorf("Failed to set app.env: %v", err)
	}

	// Save and reload to verify preservation
	if err := doc.Save(tmpFile); err != nil {
		t.Fatalf("Failed to save YAML file: %v", err)
	}

	// Reload and verify comments are preserved
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	// Check if comments are preserved
	contentStr := string(content)
	expectedComments := []string{
		"# Database configuration",
		"# Default PostgreSQL port",
		"# User credentials",
		"# Change in production!",
		"# Application settings",
		"# Environment specific settings",
	}

	for _, comment := range expectedComments {
		if !strings.Contains(contentStr, comment) {
			t.Errorf("Comment not preserved: %s", comment)
		}
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		content string
		path    string
		wantErr bool
	}{
		{
			name:    "invalid yaml",
			content: "key: [invalid",
			path:    "key",
			wantErr: true,
		},
		{
			name:    "non-existent path",
			content: "key: value",
			path:    "nonexistent",
			wantErr: true,
		},
		{
			name:    "invalid path",
			content: "key: value",
			path:    "key.nested",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.content)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("Load() error = %v", err)
				}
				return
			}

			_, err = doc.Get(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewKeyCreation(t *testing.T) {
	doc, err := Load("key: value")
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Test creating new keys
	tests := []struct {
		path  string
		value interface{}
	}{
		{"new_key", "new_value"},
		{"nested.key", "nested_value"},
		{"deeply.nested.key", "deep_value"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if err := doc.Set(tt.path, tt.value); err != nil {
				t.Errorf("Failed to set %s: %v", tt.path, err)
				return
			}

			value, err := doc.Get(tt.path)
			if err != nil {
				t.Errorf("Failed to get %s: %v", tt.path, err)
				return
			}

			if value != tt.value {
				t.Errorf("Got %v, want %v", value, tt.value)
			}
		})
	}
}

func TestDifferentTypes(t *testing.T) {
	doc, err := Load("key: value")
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		value    interface{}
		expected interface{}
	}{
		{
			name:     "string",
			path:     "string",
			value:    "value",
			expected: "value",
		},
		{
			name:     "int",
			path:     "int",
			value:    int64(42),
			expected: int64(42),
		},
		{
			name:     "float",
			path:     "float",
			value:    3.14,
			expected: 3.14,
		},
		{
			name:     "bool",
			path:     "bool",
			value:    true,
			expected: true,
		},
		{
			name:     "array",
			path:     "array",
			value:    []interface{}{1, 2, 3},
			expected: []interface{}{int64(1), int64(2), int64(3)},
		},
		{
			name:     "map",
			path:     "map",
			value:    map[string]interface{}{"key": "value"},
			expected: map[string]interface{}{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := doc.Set(tt.path, tt.value); err != nil {
				t.Errorf("Failed to set %s: %v", tt.path, err)
				return
			}

			value, err := doc.Get(tt.path)
			if err != nil {
				t.Errorf("Failed to get %s: %v", tt.path, err)
				return
			}

			if !deepEqual(value, tt.expected) {
				t.Errorf("Got %v, want %v", value, tt.expected)
			}
		})
	}
}

func TestComplexStructures(t *testing.T) {
	doc, err := Load("key: value")
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Test complex nested structures
	complexValue := map[string]interface{}{
		"string": "value",
		"number": int64(42),
		"float":  3.14,
		"bool":   true,
		"array": []interface{}{
			"item1",
			int64(2),
			true,
			map[string]interface{}{
				"nested": "value",
			},
		},
		"nested": map[string]interface{}{
			"key1": "value1",
			"key2": int64(123),
			"key3": []interface{}{
				"nested_item1",
				"nested_item2",
			},
		},
	}

	if err := doc.Set("complex", complexValue); err != nil {
		t.Fatalf("Failed to set complex value: %v", err)
	}

	// Test getting nested values
	tests := []struct {
		path     string
		expected interface{}
	}{
		{"complex.string", "value"},
		{"complex.number", int64(42)},
		{"complex.float", 3.14},
		{"complex.bool", true},
		{"complex.array[0]", "item1"},
		{"complex.array[1]", int64(2)},
		{"complex.array[2]", true},
		{"complex.array[3].nested", "value"},
		{"complex.nested.key1", "value1"},
		{"complex.nested.key2", int64(123)},
		{"complex.nested.key3[0]", "nested_item1"},
		{"complex.nested.key3[1]", "nested_item2"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			value, err := doc.Get(tt.path)
			if err != nil {
				t.Errorf("Failed to get %s: %v", tt.path, err)
				return
			}

			if !deepEqual(value, tt.expected) {
				t.Errorf("Got %v, want %v", value, tt.expected)
			}
		})
	}
}

func TestEmptyDocument(t *testing.T) {
	// Test empty document
	doc, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load empty document: %v", err)
	}

	// Test setting values in empty document
	if err := doc.Set("key", "value"); err != nil {
		t.Errorf("Failed to set value in empty document: %v", err)
	}

	value, err := doc.Get("key")
	if err != nil {
		t.Errorf("Failed to get value from empty document: %v", err)
	}
	if value != "value" {
		t.Errorf("Got %v, want %v", value, "value")
	}
}

func TestSpecialCharacters(t *testing.T) {
	doc, err := Load("key: value")
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Test special characters in keys and values
	tests := []struct {
		path     string
		value    interface{}
		expected interface{}
	}{
		{
			path:     "special.key",
			value:    "value with spaces",
			expected: "value with spaces",
		},
		{
			path:     "special.key.with.dots",
			value:    "value with dots",
			expected: "value with dots",
		},
		{
			path:     "special/key/with/slashes",
			value:    "value with slashes",
			expected: "value with slashes",
		},
		{
			path:     "special:key:with:colons",
			value:    "value with colons",
			expected: "value with colons",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if err := doc.Set(tt.path, tt.value); err != nil {
				t.Errorf("Failed to set %s: %v", tt.path, err)
				return
			}

			value, err := doc.Get(tt.path)
			if err != nil {
				t.Errorf("Failed to get %s: %v", tt.path, err)
				return
			}

			if value != tt.expected {
				t.Errorf("Got %v, want %v", value, tt.expected)
			}
		})
	}
}

func TestByteOperations(t *testing.T) {
	// Test byte operations
	content := []byte("key: value\narray:\n  - item1\n  - item2\n")

	doc, err := LoadBytes(content)
	if err != nil {
		t.Fatalf("Failed to load bytes: %v", err)
	}

	// Test getting values
	value, err := doc.GetString("key")
	if err != nil {
		t.Errorf("Failed to get string: %v", err)
	}
	if value != "value" {
		t.Errorf("Got %v, want %v", value, "value")
	}

	array, err := doc.GetStringSlice("array")
	if err != nil {
		t.Errorf("Failed to get string slice: %v", err)
	}
	if len(array) != 2 || array[0] != "item1" || array[1] != "item2" {
		t.Errorf("Got %v, want [item1 item2]", array)
	}

	// Test saving to bytes
	bytes, err := doc.ToBytes()
	if err != nil {
		t.Errorf("Failed to convert to bytes: %v", err)
	}

	// Reload and verify
	doc2, err := LoadBytes(bytes)
	if err != nil {
		t.Errorf("Failed to reload bytes: %v", err)
	}

	value, err = doc2.GetString("key")
	if err != nil {
		t.Errorf("Failed to get string from reloaded doc: %v", err)
	}
	if value != "value" {
		t.Errorf("Got %v, want %v", value, "value")
	}
}

func TestByteSliceModification(t *testing.T) {
	// Test byte slice modification
	content := []byte("key: value\narray:\n  - item1\n  - item2\n")

	doc, err := LoadBytes(content)
	if err != nil {
		t.Fatalf("Failed to load bytes: %v", err)
	}

	// Modify the document
	if err := doc.Set("key", "new_value"); err != nil {
		t.Errorf("Failed to set value: %v", err)
	}

	if err := doc.AppendToArray("array", "item3"); err != nil {
		t.Errorf("Failed to append to array: %v", err)
	}

	// Save to bytes
	bytes, err := doc.ToBytes()
	if err != nil {
		t.Errorf("Failed to convert to bytes: %v", err)
	}

	// Reload and verify
	doc2, err := LoadBytes(bytes)
	if err != nil {
		t.Errorf("Failed to reload bytes: %v", err)
	}

	value, err := doc2.GetString("key")
	if err != nil {
		t.Errorf("Failed to get string from reloaded doc: %v", err)
	}
	if value != "new_value" {
		t.Errorf("Got %v, want %v", value, "new_value")
	}

	array, err := doc2.GetStringSlice("array")
	if err != nil {
		t.Errorf("Failed to get string slice from reloaded doc: %v", err)
	}
	if len(array) != 3 || array[0] != "item1" || array[1] != "item2" || array[2] != "item3" {
		t.Errorf("Got %v, want [item1 item2 item3]", array)
	}
}
