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
		expected string
	}{
		{"db.host", "localhost"},
		{"db.port", "5432"},
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
				t.Errorf("Got %v, want %v", value, tt.expected)
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

func TestArrayOperations(t *testing.T) {
	yamlContent := `# Array test
items:
  # List of items
  - name: item1  # First item
    value: 100
  - name: item2  # Second item
    value: 200
  # More items can be added here
tags:
  - tag1  # Primary tag
  - tag2  # Secondary tag`

	doc, err := Load(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Test getting array values
	tags, err := doc.Get("tags")
	if err != nil {
		t.Fatalf("Failed to get tags: %v", err)
	}

	tagsArr, ok := tags.([]interface{})
	if !ok {
		t.Fatal("tags is not an array")
	}

	if len(tagsArr) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tagsArr))
	}

	if tagsArr[0] != "tag1" {
		t.Errorf("Expected tag1, got %v", tagsArr[0])
	}
}

func TestErrorHandling(t *testing.T) {
	yamlContent := `key1: value1
key2: value2`

	doc, err := Load(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Test getting non-existent key
	_, err = doc.Get("nonexistent.key")
	if err == nil {
		t.Error("Expected error for non-existent key, got nil")
	}

	// Test invalid path
	_, err = doc.Get("key1.nonexistent")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}

	// Test setting value on non-existent path (should succeed as we create intermediate nodes)
	err = doc.Set("new.nested.key", "value")
	if err != nil {
		t.Error("Expected success for setting new nested path")
	}

	// Verify the new path was created
	value, err := doc.Get("new.nested.key")
	if err != nil {
		t.Errorf("Failed to get new nested key: %v", err)
	}
	if value != "value" {
		t.Errorf("Got %v, want 'value'", value)
	}
}

func TestNewKeyCreation(t *testing.T) {
	yamlContent := `existing:
  key: value`

	doc, err := Load(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Test creating new key at root level
	err = doc.Set("newkey", "newvalue")
	if err != nil {
		t.Fatalf("Failed to set new key: %v", err)
	}

	value, err := doc.Get("newkey")
	if err != nil {
		t.Fatalf("Failed to get new key: %v", err)
	}

	if value != "newvalue" {
		t.Errorf("Expected newvalue, got %v", value)
	}
}

func TestDifferentTypes(t *testing.T) {
	yamlContent := `# Different types test
string: hello
integer: 42
float: 3.14
boolean: true
null_value: null`

	doc, err := Load(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Test getting values
	t.Run("string", func(t *testing.T) {
		value, err := doc.Get("string")
		if err != nil {
			t.Errorf("Failed to get string: %v", err)
			return
		}
		if value != "hello" {
			t.Errorf("Got %v, want hello", value)
		}
	})

	t.Run("integer", func(t *testing.T) {
		value, err := doc.Get("integer")
		if err != nil {
			t.Errorf("Failed to get integer: %v", err)
			return
		}
		if value != "42" {
			t.Errorf("Got %v, want 42", value)
		}
	})

	t.Run("float", func(t *testing.T) {
		value, err := doc.Get("float")
		if err != nil {
			t.Errorf("Failed to get float: %v", err)
			return
		}
		if value != "3.14" {
			t.Errorf("Got %v, want 3.14", value)
		}
	})

	t.Run("boolean", func(t *testing.T) {
		value, err := doc.Get("boolean")
		if err != nil {
			t.Errorf("Failed to get boolean: %v", err)
			return
		}
		if value != true {
			t.Errorf("Got %v, want true", value)
		}
	})

	t.Run("null_value", func(t *testing.T) {
		value, err := doc.Get("null_value")
		if err != nil {
			t.Errorf("Failed to get null: %v", err)
			return
		}
		if value != nil {
			t.Errorf("Got %v, want nil", value)
		}
	})

	// Test setting different types
	testValues := map[string]interface{}{
		"new_string": "test",
		"new_int":    42,
		"new_float":  3.14,
		"new_bool":   true,
		"new_array":  []interface{}{"a", "b", "c"},
		"new_map":    map[string]interface{}{"key": "value"},
	}

	for key, value := range testValues {
		if err := doc.Set(key, value); err != nil {
			t.Errorf("Failed to set %s: %v", key, err)
		}
	}
}

func TestComplexStructures(t *testing.T) {
	yamlContent := `# Complex configuration
service:
  # Database settings
  database:
    primary:
      host: localhost
      port: 5432
      credentials:
        user: admin
        pass: secret
    replicas:
      - host: replica1
        port: 5432
      - host: replica2
        port: 5432

  # Cache settings
  cache:
    redis:
      - host: cache1
        port: 6379
      - host: cache2
        port: 6379

  # Application settings
  app:
    name: MyApp
    environment: |
      This is a multiline
      environment description
      with preserved formatting
    features:
      - name: feature1
        enabled: true
        config:
          timeout: 30
      - name: feature2
        enabled: false
        config:
          timeout: 60`

	doc, err := Load(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Test nested structure access
	tests := []struct {
		path     string
		expected interface{}
	}{
		{"service.database.primary.host", "localhost"},
		{"service.database.primary.port", "5432"},
		{"service.database.primary.credentials.user", "admin"},
		{"service.app.name", "MyApp"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
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

	// Test multiline string preservation
	envValue, err := doc.Get("service.app.environment")
	if err != nil {
		t.Fatalf("Failed to get environment: %v", err)
	}
	expectedEnv := "This is a multiline\nenvironment description\nwith preserved formatting\n"
	if envValue != expectedEnv {
		t.Errorf("Multiline string not preserved correctly.\nGot:\n%s\nWant:\n%s", envValue, expectedEnv)
	}

	// Test nested array access
	features, err := doc.Get("service.app.features")
	if err != nil {
		t.Fatalf("Failed to get features: %v", err)
	}

	featuresArr, ok := features.([]interface{})
	if !ok {
		t.Fatal("features is not an array")
	}

	if len(featuresArr) != 2 {
		t.Errorf("Expected 2 features, got %d", len(featuresArr))
	}

	// Test setting nested values
	err = doc.Set("service.database.primary.host", "newhost")
	if err != nil {
		t.Errorf("Failed to set database host: %v", err)
	}

	// Test creating new nested structure
	err = doc.Set("service.newservice.key.subkey", "value")
	if err != nil {
		t.Errorf("Failed to create new nested structure: %v", err)
	}

	// Verify the new structure
	value, err := doc.Get("service.newservice.key.subkey")
	if err != nil {
		t.Errorf("Failed to get new nested value: %v", err)
	}
	if value != "value" {
		t.Errorf("Got %v, want 'value'", value)
	}

	// Save and verify formatting
	tmpFile := "test_complex.yaml"
	if err := doc.Save(tmpFile); err != nil {
		t.Fatalf("Failed to save YAML: %v", err)
	}
	defer os.Remove(tmpFile)

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	// Verify comments and formatting are preserved
	contentStr := string(content)
	expectedParts := []string{
		"# Complex configuration",
		"# Database settings",
		"# Cache settings",
		"# Application settings",
		"environment: |",
		"  This is a multiline",
		"  environment description",
		"  with preserved formatting",
	}

	for _, part := range expectedParts {
		if !strings.Contains(contentStr, part) {
			t.Errorf("Expected content not found: %s", part)
		}
	}
}

func TestEmptyDocument(t *testing.T) {
	// Test creating empty document
	doc, err := Load("")
	if err != nil {
		t.Fatalf("Failed to create empty document: %v", err)
	}

	// Test setting root value
	err = doc.Set("", map[string]interface{}{
		"key": "value",
	})
	if err != nil {
		t.Fatalf("Failed to set root value: %v", err)
	}

	// Verify the value
	value, err := doc.Get("key")
	if err != nil {
		t.Fatalf("Failed to get value: %v", err)
	}
	if value != "value" {
		t.Errorf("Got %v, want 'value'", value)
	}
}

func TestSpecialCharacters(t *testing.T) {
	yamlContent := `special:
  key-with-dash: value1
  "key:with:colon": value2
  'key with spaces': value3
  binary: !!binary YWJjZGVm
  reference: &ref value4
  alias: *ref`

	doc, err := Load(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	tests := []struct {
		path     string
		expected string
	}{
		{"special.key-with-dash", "value1"},
		{"special.key:with:colon", "value2"},
		{"special.key with spaces", "value3"},
		{"special.reference", "value4"},
		{"special.alias", "value4"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
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
	// Test data
	yamlContent := []byte(`# Test configuration
app:
  name: TestApp
  version: 1.0.0
  # Database settings
  database:
    host: localhost
    port: 5432
    # Credentials
    credentials:
      username: admin
      password: secret`)

	// Test LoadBytes
	doc, err := LoadBytes(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load YAML from bytes: %v", err)
	}

	// Test getting values
	tests := []struct {
		path     string
		expected string
	}{
		{"app.name", "TestApp"},
		{"app.version", "1.0.0"},
		{"app.database.host", "localhost"},
		{"app.database.credentials.username", "admin"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
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

	// Test modifying values
	err = doc.Set("app.name", "UpdatedApp")
	if err != nil {
		t.Fatalf("Failed to set app.name: %v", err)
	}

	// Test ToBytes
	output, err := doc.ToBytes()
	if err != nil {
		t.Fatalf("Failed to convert to bytes: %v", err)
	}

	// Load the output back and verify
	newDoc, err := LoadBytes(output)
	if err != nil {
		t.Fatalf("Failed to load modified YAML: %v", err)
	}

	// Verify the modified value
	value, err := newDoc.Get("app.name")
	if err != nil {
		t.Fatalf("Failed to get app.name: %v", err)
	}
	if value != "UpdatedApp" {
		t.Errorf("Got %v, want UpdatedApp", value)
	}

	// Test String method
	str, err := doc.String()
	if err != nil {
		t.Fatalf("Failed to convert to string: %v", err)
	}

	// Verify that comments are preserved
	expectedComments := []string{
		"# Test configuration",
		"# Database settings",
		"# Credentials",
	}

	for _, comment := range expectedComments {
		if !strings.Contains(str, comment) {
			t.Errorf("Comment not preserved in string output: %s", comment)
		}
	}

	// Test empty document
	emptyDoc, err := LoadBytes([]byte{})
	if err != nil {
		t.Fatalf("Failed to load empty bytes: %v", err)
	}

	// Set and get values in empty document
	err = emptyDoc.Set("key", "value")
	if err != nil {
		t.Fatalf("Failed to set value in empty document: %v", err)
	}

	value, err = emptyDoc.Get("key")
	if err != nil {
		t.Fatalf("Failed to get value from empty document: %v", err)
	}
	if value != "value" {
		t.Errorf("Got %v, want value", value)
	}

	// Test invalid YAML bytes
	_, err = LoadBytes([]byte(`invalid: yaml: content`))
	if err == nil {
		t.Error("Expected error for invalid YAML content, got nil")
	}
}

func TestByteSliceModification(t *testing.T) {
	// Initial YAML
	yamlContent := []byte(`data:
  list:
    - item1
    - item2
  nested:
    key: value`)

	// Load and modify
	doc, err := LoadBytes(yamlContent)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Add new items at different levels
	modifications := []struct {
		path  string
		value interface{}
	}{
		{"data.list[2]", "item3"},
		{"data.nested.newkey", "newvalue"},
		{"data.nested.deep.deeper", map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}},
	}

	for _, mod := range modifications {
		err := doc.Set(mod.path, mod.value)
		if err != nil {
			t.Errorf("Failed to set %s: %v", mod.path, err)
		}
	}

	// Convert back to bytes
	output, err := doc.ToBytes()
	if err != nil {
		t.Fatalf("Failed to convert to bytes: %v", err)
	}

	// Load modified content and verify
	modifiedDoc, err := LoadBytes(output)
	if err != nil {
		t.Fatalf("Failed to load modified content: %v", err)
	}

	// Verify modifications
	value, err := modifiedDoc.Get("data.nested.newkey")
	if err != nil {
		t.Fatalf("Failed to get new key: %v", err)
	}
	if value != "newvalue" {
		t.Errorf("Got %v, want newvalue", value)
	}

	value, err = modifiedDoc.Get("data.nested.deep.deeper.key1")
	if err != nil {
		t.Fatalf("Failed to get nested key: %v", err)
	}
	if value != "value1" {
		t.Errorf("Got %v, want value1", value)
	}
}
