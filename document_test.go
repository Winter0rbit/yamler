package yamler

import (
	"os"
	"strings"
	"testing"
)

func TestLoadFile(t *testing.T) {
	// Create a temporary file
	content := []byte("key: value\narray:\n  - item1\n  - item2\n")
	tmpfile, err := os.CreateTemp("", "test*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading the file
	doc, err := LoadFile(tmpfile.Name())
	if err != nil {
		t.Errorf("LoadFile() error = %v", err)
		return
	}

	// Verify the content
	value, err := doc.GetString("key")
	if err != nil {
		t.Errorf("GetString() error = %v", err)
		return
	}
	if value != "value" {
		t.Errorf("GetString() = %v, want %v", value, "value")
	}

	array, err := doc.GetStringSlice("array")
	if err != nil {
		t.Errorf("GetStringSlice() error = %v", err)
		return
	}
	if len(array) != 2 || array[0] != "item1" || array[1] != "item2" {
		t.Errorf("GetStringSlice() = %v, want [item1 item2]", array)
	}
}

func TestLoadBytes(t *testing.T) {
	content := []byte("key: value\narray:\n  - item1\n  - item2\n")

	doc, err := LoadBytes(content)
	if err != nil {
		t.Errorf("LoadBytes() error = %v", err)
		return
	}

	// Verify the content
	value, err := doc.GetString("key")
	if err != nil {
		t.Errorf("GetString() error = %v", err)
		return
	}
	if value != "value" {
		t.Errorf("GetString() = %v, want %v", value, "value")
	}

	array, err := doc.GetStringSlice("array")
	if err != nil {
		t.Errorf("GetStringSlice() error = %v", err)
		return
	}
	if len(array) != 2 || array[0] != "item1" || array[1] != "item2" {
		t.Errorf("GetStringSlice() = %v, want [item1 item2]", array)
	}
}

func TestLoad(t *testing.T) {
	content := "key: value\narray:\n  - item1\n  - item2\n"

	doc, err := Load(content)
	if err != nil {
		t.Errorf("Load() error = %v", err)
		return
	}

	// Verify the content
	value, err := doc.GetString("key")
	if err != nil {
		t.Errorf("GetString() error = %v", err)
		return
	}
	if value != "value" {
		t.Errorf("GetString() = %v, want %v", value, "value")
	}

	array, err := doc.GetStringSlice("array")
	if err != nil {
		t.Errorf("GetStringSlice() error = %v", err)
		return
	}
	if len(array) != 2 || array[0] != "item1" || array[1] != "item2" {
		t.Errorf("GetStringSlice() = %v, want [item1 item2]", array)
	}
}

func TestLoadError(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "invalid yaml",
			content: "key: [invalid",
			wantErr: true,
		},
		{
			name:    "empty content",
			content: "",
			wantErr: false,
		},
		{
			name:    "valid yaml",
			content: "key: value",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDocument_String(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "simple yaml",
			content: "key: value",
			want:    "key: value\n",
		},
		{
			name: "complex yaml with 2-space indentation",
			content: `key: value
array:
  - item1
  - item2
nested:
  key: value`,
			want: `key: value
array:
  - item1
  - item2
nested:
  key: value
`,
		},
		{
			name: "complex yaml with 4-space indentation",
			content: `key: value
array:
    - item1
    - item2
nested:
    key: value`,
			want: `key: value
array:
    - item1
    - item2
nested:
    key: value
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.content)
			if err != nil {
				t.Errorf("Load() error = %v", err)
				return
			}

			got, err := doc.String()
			if err != nil {
				t.Errorf("Document.String() error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("Document.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectIndentation(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected int
	}{
		{
			name: "2-space indentation",
			yaml: `name: test
description: A test
options:
  test:
    folder_id: abc123
    class: small`,
			expected: 2,
		},
		{
			name: "4-space indentation",
			yaml: `name: test
description: A test
options:
    test:
        folder_id: abc123
        class: small`,
			expected: 4,
		},
		{
			name: "no indentation",
			yaml: `name: test
description: A test`,
			expected: 2, // default
		},
		{
			name: "mixed indentation - first wins",
			yaml: `name: test
options:
  first: value
    second: value`,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectIndentation(tt.yaml)
			if result != tt.expected {
				t.Errorf("detectIndentation() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

func TestPreserveFormatting(t *testing.T) {
	originalYAML := `import: ../../infra/generic-platform/infra/common.group.yml

type: database

name: test-service
description: Database cluster for service test-service-*

options:
  test:
    folder_id: folder-abc123def456
    class: small.standard
    disk_size: 16
  prod:
    folder_id: folder-abc123def456
    class: micro.standard
    disk_size: 24

provides:
  - name: api
    protocol: tcp
    description: Database API
    port: 6379`

	doc, err := Load(originalYAML)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Make some changes
	err = doc.Set("options.test.class", "medium.standard")
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	err = doc.Set("options.test.disk_size", 32)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	err = doc.Set("options.prod.class", "large.standard")
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	err = doc.Set("options.prod.disk_size", 48)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Convert back to string
	result, err := doc.String()
	if err != nil {
		t.Fatalf("Failed to convert to string: %v", err)
	}

	t.Logf("Original:\n%s", originalYAML)
	t.Logf("Result:\n%s", result)

	// Check that indentation is preserved (2 spaces)
	lines := strings.Split(result, "\n")

	// Find a line that should have 2-space indentation
	found2Space := false
	found4Space := false

	for _, line := range lines {
		if strings.HasPrefix(line, "  test:") || strings.HasPrefix(line, "  prod:") {
			found2Space = true
		}
		if strings.HasPrefix(line, "    folder_id:") || strings.HasPrefix(line, "    class:") {
			found4Space = true
		}
	}

	if !found2Space {
		t.Error("Expected to find 2-space indentation for 'test:' or 'prod:'")
	}
	if !found4Space {
		t.Error("Expected to find 4-space indentation for nested properties")
	}

	// Verify changes were applied
	testClass, err := doc.GetString("options.test.class")
	if err != nil {
		t.Fatalf("Failed to get test class: %v", err)
	}
	if testClass != "medium.standard" {
		t.Errorf("Expected test class to be 'medium.standard', got '%s'", testClass)
	}
}
