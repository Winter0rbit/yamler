package yamler

import (
	"testing"
)

// TestFormattingPreservation tests that the original formatting style is preserved
func TestFormattingPreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name: "preserve_flow_style_array",
			input: `config:
  items: [1, 2, 3]
  name: test`,
			key:      "config.name",
			newValue: "updated",
			expectedOutput: `config:
  items: [1, 2, 3]
  name: updated
`,
		},
		{
			name: "preserve_block_style_array",
			input: `config:
  items:
    - 1
    - 2
    - 3
  name: test`,
			key:      "config.name",
			newValue: "updated",
			expectedOutput: `config:
  items:
    - 1
    - 2
    - 3
  name: updated
`,
		},
		{
			name: "preserve_mixed_styles",
			input: `config:
  flow_array: [a, b, c]
  block_array:
    - x
    - y
    - z
  name: test`,
			key:      "config.name",
			newValue: "updated",
			expectedOutput: `config:
  flow_array: [a, b, c]
  block_array:
    - x
    - y
    - z
  name: updated
`,
		},
		{
			name: "preserve_nested_flow_arrays",
			input: `data:
  matrix: [[1, 2], [3, 4]]
  simple: test`,
			key:      "data.simple",
			newValue: "updated",
			expectedOutput: `data:
  matrix: [[1, 2], [3, 4]]
  simple: updated
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = doc.Set(tt.key, tt.newValue)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Formatting not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestArrayOperationPreservesStyle tests that array operations preserve the original array style
func TestArrayOperationPreservesStyle(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		operation      func(*Document) error
		expectedOutput string
	}{
		{
			name: "append_to_flow_array_preserves_flow",
			input: `items: [1, 2, 3]
name: test`,
			operation: func(d *Document) error {
				return d.AppendToArray("items", 4)
			},
			expectedOutput: `items: [1, 2, 3, 4]
name: test
`,
		},
		{
			name: "append_to_block_array_preserves_block",
			input: `items:
  - 1
  - 2
  - 3
name: test`,
			operation: func(d *Document) error {
				return d.AppendToArray("items", 4)
			},
			expectedOutput: `items:
  - 1
  - 2
  - 3
  - 4
name: test
`,
		},
		{
			name: "update_flow_array_element",
			input: `items: [1, 2, 3]
name: test`,
			operation: func(d *Document) error {
				return d.UpdateArrayElement("items", 1, 99)
			},
			expectedOutput: `items: [1, 99, 3]
name: test
`,
		},
		{
			name: "update_block_array_element",
			input: `items:
  - 1
  - 2
  - 3
name: test`,
			operation: func(d *Document) error {
				return d.UpdateArrayElement("items", 1, 99)
			},
			expectedOutput: `items:
  - 1
  - 99
  - 3
name: test
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = tt.operation(doc)
			if err != nil {
				t.Fatalf("Operation error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Array style not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestCommentPreservation tests that comments are preserved during operations
func TestCommentPreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		operation      func(*Document) error
		expectedOutput string
	}{
		{
			name: "preserve_comments_on_set",
			input: `# Main config
config:
  # Database settings
  db:
    host: localhost # Default host
    port: 5432
  # Application settings  
  app:
    name: myapp # Application name
    debug: false`,
			operation: func(d *Document) error {
				return d.Set("config.db.port", 3306)
			},
			expectedOutput: `# Main config
config:
  # Database settings
  db:
    host: localhost # Default host
    port: 3306
  # Application settings  
  app:
    name: myapp # Application name
    debug: false
`,
		},
		{
			name: "preserve_comments_on_array_operations",
			input: `# Array config
items: # My items
  - item1 # First item
  - item2 # Second item
  - item3 # Third item`,
			operation: func(d *Document) error {
				return d.AppendToArray("items", "item4")
			},
			expectedOutput: `# Array config
items: # My items
  - item1 # First item
  - item2 # Second item
  - item3 # Third item
  - item4
`,
		},
		{
			name: "preserve_comments_on_flow_array_operations",
			input: `# Flow array config
items: [item1, item2, item3] # My flow items
other: value`,
			operation: func(d *Document) error {
				return d.AppendToArray("items", "item4")
			},
			expectedOutput: `# Flow array config
items: [item1, item2, item3, item4] # My flow items
other: value
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = tt.operation(doc)
			if err != nil {
				t.Fatalf("Operation error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Comments not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestKeyOrderPreservation tests that key order is preserved during operations
func TestKeyOrderPreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		operation      func(*Document) error
		expectedOutput string
	}{
		{
			name: "preserve_key_order_on_set",
			input: `first: 1
second: 2
third: 3
fourth: 4`,
			operation: func(d *Document) error {
				return d.Set("second", "updated")
			},
			expectedOutput: `first: 1
second: updated
third: 3
fourth: 4
`,
		},
		{
			name: "preserve_nested_key_order",
			input: `config:
  first: 1
  second: 2
  third: 3
other: value`,
			operation: func(d *Document) error {
				return d.Set("config.second", "updated")
			},
			expectedOutput: `config:
  first: 1
  second: updated
  third: 3
other: value
`,
		},
		{
			name: "add_new_key_at_end",
			input: `first: 1
second: 2
third: 3`,
			operation: func(d *Document) error {
				return d.Set("fourth", 4)
			},
			expectedOutput: `first: 1
second: 2
third: 3
fourth: 4
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = tt.operation(doc)
			if err != nil {
				t.Fatalf("Operation error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Key order not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestTrailingNewlinesPreservation tests that trailing newlines are preserved
func TestTrailingNewlinesPreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name: "preserve_single_trailing_newline",
			input: `config:
  name: test
  value: 123
`,
			key:      "config.name",
			newValue: "updated",
			expectedOutput: `config:
  name: updated
  value: 123
`,
		},
		{
			name: "preserve_multiple_trailing_newlines",
			input: `config:
  name: test
  value: 123


`,
			key:      "config.name",
			newValue: "updated",
			expectedOutput: `config:
  name: updated
  value: 123


`,
		},
		{
			name: "preserve_trailing_newlines_with_comments",
			input: `# Configuration
config:
  name: test
  value: 123

# End of file

`,
			key:      "config.name",
			newValue: "updated",
			expectedOutput: `# Configuration
config:
  name: updated
  value: 123

# End of file

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			err = doc.Set(tt.key, tt.newValue)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Trailing newlines not preserved.\nGot:\n%q\nWant:\n%q", result, tt.expectedOutput)
			}
		})
	}
}
