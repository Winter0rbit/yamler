package yamler

import (
	"testing"
)

func TestDocument_Merge(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		other    string
		expected string
		wantErr  bool
	}{
		{
			name: "merge simple maps",
			base: `name: test
version: 1.0`,
			other: `author: developer
license: MIT`,
			expected: `name: test
version: 1.0
author: developer
license: MIT
`,
		},
		{
			name: "merge with overlap",
			base: `name: test
version: 1.0
config:
  debug: true`,
			other: `name: new-test
config:
  timeout: 30
author: developer`,
			expected: `name: new-test
version: 1.0
config:
  debug: true
  timeout: 30
author: developer
`,
		},
		{
			name: "merge nested structures",
			base: `app:
  name: myapp
  settings:
    debug: true
    port: 8080`,
			other: `app:
  version: 2.0
  settings:
    timeout: 30
    debug: false`,
			expected: `app:
  name: myapp
  settings:
    debug: false
    port: 8080
    timeout: 30
  version: 2.0
`,
		},
		{
			name: "merge arrays (replace)",
			base: `items: [1, 2, 3]
name: test`,
			other: `items: [4, 5, 6]
new_field: value`,
			expected: `items: [4, 5, 6]
name: test
new_field: value
`,
		},
		{
			name: "preserve comments",
			base: `# Main config
name: test # App name
version: 1.0`,
			other: `name: new-test
author: developer`,
			expected: `# Main config
name: new-test # App name
version: 1.0
author: developer
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, err := Load(tt.base)
			if err != nil {
				t.Fatalf("Failed to load base document: %v", err)
			}

			other, err := Load(tt.other)
			if err != nil {
				t.Fatalf("Failed to load other document: %v", err)
			}

			err = base.Merge(other)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.Merge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				result, err := base.String()
				if err != nil {
					t.Fatalf("Failed to convert result to string: %v", err)
				}

				if result != tt.expected {
					t.Errorf("Document.Merge() result mismatch\nGot:\n%s\nWant:\n%s", result, tt.expected)
				}
			}
		})
	}
}

func TestDocument_MergeAt(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		path     string
		other    string
		expected string
		wantErr  bool
	}{
		{
			name: "merge at existing path",
			base: `config:
  app:
    name: test
    version: 1.0
  db:
    host: localhost`,
			path: "config.app",
			other: `author: developer
version: 2.0`,
			expected: `config:
  app:
    name: test
    version: 2.0
    author: developer
  db:
    host: localhost
`,
		},
		{
			name: "merge at new path",
			base: `config:
  app:
    name: test`,
			path: "config.database",
			other: `host: localhost
port: 5432`,
			expected: `config:
  app:
    name: test
  database:
    host: localhost
    port: 5432
`,
		},
		{
			name: "merge at root level",
			base: `existing: value`,
			path: "new_section",
			other: `field1: value1
field2: value2`,
			expected: `existing: value
new_section:
  field1: value1
  field2: value2
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base, err := Load(tt.base)
			if err != nil {
				t.Fatalf("Failed to load base document: %v", err)
			}

			other, err := Load(tt.other)
			if err != nil {
				t.Fatalf("Failed to load other document: %v", err)
			}

			err = base.MergeAt(tt.path, other)
			if (err != nil) != tt.wantErr {
				t.Errorf("Document.MergeAt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				result, err := base.String()
				if err != nil {
					t.Fatalf("Failed to convert result to string: %v", err)
				}

				if result != tt.expected {
					t.Errorf("Document.MergeAt() result mismatch\nGot:\n%s\nWant:\n%s", result, tt.expected)
				}
			}
		})
	}
}

func TestDocument_MergeErrors(t *testing.T) {
	base, _ := Load("name: test")

	// Test merging nil document
	err := base.Merge(nil)
	if err == nil {
		t.Error("Expected error when merging nil document")
	}

	// Test merging at path with nil document
	err = base.MergeAt("some.path", nil)
	if err == nil {
		t.Error("Expected error when merging nil document at path")
	}
}

func TestMergeCommentPreservation(t *testing.T) {
	base := `# Base config
name: test # Original name
version: 1.0
# Database section
db:
  host: localhost # Default host`

	other := `name: new-name
db:
  port: 5432`

	expected := `# Base config
name: new-name # Original name
version: 1.0
# Database section
db:
  host: localhost # Default host
  port: 5432
`

	baseDoc, err := Load(base)
	if err != nil {
		t.Fatalf("Failed to load base: %v", err)
	}

	otherDoc, err := Load(other)
	if err != nil {
		t.Fatalf("Failed to load other: %v", err)
	}

	err = baseDoc.Merge(otherDoc)
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	result, err := baseDoc.String()
	if err != nil {
		t.Fatalf("Failed to convert to string: %v", err)
	}

	if result != expected {
		t.Errorf("Comments not preserved correctly\nGot:\n%s\nWant:\n%s", result, expected)
	}
}
