package yamler

import (
	"testing"
)

// TestInlineObjectSpacePreservation tests that spaces inside curly braces are preserved
func TestInlineObjectSpacePreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name: "preserve_spaces_in_flow_objects",
			input: `datacenters:
  sas: { count: 2 }
  vla: { count: 2 }`,
			key:      "datacenters.sas.count",
			newValue: 3,
			expectedOutput: `datacenters:
  sas: { count: 3 }
  vla: { count: 2 }
`,
		},
		{
			name: "preserve_spaces_in_nested_flow_objects",
			input: `config:
  databases:
    prod: { host: localhost, port: 5432 }
    test: { host: testhost, port: 5433 }`,
			key:      "config.databases.prod.port",
			newValue: 5434,
			expectedOutput: `config:
  databases:
    prod: { host: localhost, port: 5434 }
    test: { host: testhost, port: 5433 }
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
				t.Errorf("Spaces in flow objects not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestFieldOrderPreservation tests that field order is preserved during updates
func TestFieldOrderPreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		operations     []func(*Document) error
		expectedOutput string
	}{
		{
			name: "preserve_field_order_when_updating_memory_then_cpu",
			input: `resources:
  cpu: 111
  memory: 111`,
			operations: []func(*Document) error{
				func(d *Document) error { return d.Set("resources.memory", 222) },
				func(d *Document) error { return d.Set("resources.cpu", 333) },
			},
			expectedOutput: `resources:
  cpu: 333
  memory: 222
`,
		},
		{
			name: "preserve_field_order_when_updating_cpu_then_memory",
			input: `resources:
  memory: 111
  cpu: 111`,
			operations: []func(*Document) error{
				func(d *Document) error { return d.Set("resources.cpu", 333) },
				func(d *Document) error { return d.Set("resources.memory", 222) },
			},
			expectedOutput: `resources:
  memory: 222
  cpu: 333
`,
		},
		{
			name: "preserve_complex_field_order",
			input: `app:
  name: myapp
  version: 1.0
  debug: true
  port: 8080`,
			operations: []func(*Document) error{
				func(d *Document) error { return d.Set("app.debug", false) },
				func(d *Document) error { return d.Set("app.port", 9090) },
				func(d *Document) error { return d.Set("app.version", "2.0") },
			},
			expectedOutput: `app:
  name: myapp
  version: "2.0"
  debug: false
  port: 9090
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			for _, operation := range tt.operations {
				err = operation(doc)
				if err != nil {
					t.Fatalf("Operation error = %v", err)
				}
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Field order not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestInlineObjectFormatPreservation tests that inline objects don't get expanded to multiline
func TestInlineObjectFormatPreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name:     "preserve_inline_format_when_updating_values",
			input:    `resources: {cpu: 111, memory: 111}`,
			key:      "resources.cpu",
			newValue: 222,
			expectedOutput: `resources: {cpu: 222, memory: 111}
`,
		},
		{
			name:     "preserve_inline_format_with_spaces",
			input:    `resources: { cpu: 111, memory: 111 }`,
			key:      "resources.memory",
			newValue: 222,
			expectedOutput: `resources: { cpu: 111, memory: 222 }
`,
		},
		{
			name: "preserve_inline_format_nested",
			input: `config:
  resources: {cpu: 111, memory: 111}
  name: test`,
			key:      "config.resources.cpu",
			newValue: 222,
			expectedOutput: `config:
  resources: {cpu: 222, memory: 111}
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

			err = doc.Set(tt.key, tt.newValue)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expectedOutput {
				t.Errorf("Inline object format not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestTrailingNewlinesPreservation tests that trailing newlines are preserved
func TestTrailingNewlinesInSectionsPreservation(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		key            string
		newValue       interface{}
		expectedOutput string
	}{
		{
			name: "preserve_section_spacing",
			input: `section1:
  key1: value1


section2:
  key2: value2`,
			key:      "section1.key1",
			newValue: "updated",
			expectedOutput: `section1:
  key1: updated


section2:
  key2: value2
`,
		},
		{
			name: "preserve_multiple_newlines_between_sections",
			input: `app:
  name: test



database:
  host: localhost`,
			key:      "app.name",
			newValue: "updated",
			expectedOutput: `app:
  name: updated



database:
  host: localhost
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
				t.Errorf("Section spacing not preserved.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}
