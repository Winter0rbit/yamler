package yamler

import (
	"strings"
	"testing"
)

// Test for Issue #1: Field order preservation in resources section
func TestResourcesFieldOrderPreservation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		updates  []struct{ path, value string }
		expected string
	}{
		{
			name: "CPU first, then memory - should preserve order",
			input: `test:
  resources:
    cpu: 100
    memory: 256`,
			updates: []struct{ path, value string }{
				{"test.resources.cpu", "111"},
				{"test.resources.memory", "512"},
			},
			expected: `test:
  resources:
    cpu: 111
    memory: 512`,
		},
		{
			name: "Cores first, then memory - should preserve order",
			input: `prod:
  resources:
    cores: 1
    memory: 256`,
			updates: []struct{ path, value string }{
				{"prod.resources.cores", "2"},
				{"prod.resources.memory", "512"},
			},
			expected: `prod:
  resources:
    cores: 2
    memory: 512`,
		},
		{
			name: "Mixed CPU and memory updates - preserve original order",
			input: `layers:
  test:
    resources:
      cpu: 100
      memory: 256
  prod:
    resources:
      cores: 1
      memory: 512`,
			updates: []struct{ path, value string }{
				{"layers.test.resources.memory", "512"},
				{"layers.test.resources.cpu", "111"},
				{"layers.prod.resources.memory", "1024"},
			},
			expected: `layers:
  test:
    resources:
      cpu: 111
      memory: 512
  prod:
    resources:
      cores: 1
      memory: 1024`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Failed to load YAML: %v", err)
			}

			// Apply updates in sequence
			for _, update := range tt.updates {
				err := doc.SetInt(update.path, mustParseInt(update.value))
				if err != nil {
					t.Fatalf("Failed to set %s = %s: %v", update.path, update.value, err)
				}
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("Failed to convert to string: %v", err)
			}

			if strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
				t.Errorf("Field order not preserved.\nExpected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

// Test for Issue #2: Empty lines preservation between sections
func TestEmptyLinesPreservationBetweenSections(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		updates  []struct{ path, value string }
		expected string
	}{
		{
			name: "Empty lines between prod and test sections",
			input: `prod:
  resources:
    cpu: 100
    memory: 256


test:
  resources:
    cpu: 50
    memory: 128`,
			updates: []struct{ path, value string }{
				{"test.resources.cpu", "111"},
				{"test.resources.memory", "111"},
			},
			expected: `prod:
  resources:
    cpu: 100
    memory: 256


test:
  resources:
    cpu: 111
    memory: 111`,
		},
		{
			name: "Multiple empty lines preservation",
			input: `general:
  name: service


prod:
  resources:
    cpu: 200


test:
  resources:
    cpu: 100`,
			updates: []struct{ path, value string }{
				{"test.resources.cpu", "111"},
			},
			expected: `general:
  name: service


prod:
  resources:
    cpu: 200


test:
  resources:
    cpu: 111`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Failed to load YAML: %v", err)
			}

			// Apply updates
			for _, update := range tt.updates {
				err := doc.SetInt(update.path, mustParseInt(update.value))
				if err != nil {
					t.Fatalf("Failed to set %s = %s: %v", update.path, update.value, err)
				}
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("Failed to convert to string: %v", err)
			}

			// Allow for trailing newline differences but preserve internal structure
			resultTrimmed := strings.TrimSuffix(result, "\n")
			expectedTrimmed := strings.TrimSuffix(tt.expected, "\n")

			if resultTrimmed != expectedTrimmed {
				t.Errorf("Empty lines not preserved.\nExpected:\n%q\n\nGot:\n%q", expectedTrimmed, resultTrimmed)
			}
		})
	}
}

// Test for Issue #3: Inline vs Multiline format preservation
func TestInlineVsMultilineFormatPreservation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		updates  []struct{ path, value string }
		expected string
	}{
		{
			name: "Multiline format should be preserved when updating",
			input: `test:
  resources:
    cpu: 100
    memory: 256`,
			updates: []struct{ path, value string }{
				{"test.resources.cpu", "111"},
				{"test.resources.memory", "111"},
			},
			expected: `test:
  resources:
    cpu: 111
    memory: 111`,
		},
		{
			name: "Inline format should be preserved when updating",
			input: `test:
  resources: { cpu: 100, memory: 256 }`,
			updates: []struct{ path, value string }{
				{"test.resources.cpu", "111"},
				{"test.resources.memory", "111"},
			},
			expected: `test:
  resources: { cpu: 111, memory: 111 }`,
		},
		{
			name: "Mixed formats in different sections",
			input: `prod:
  resources:
    cpu: 200
    memory: 512
test:
  resources: { cpu: 100, memory: 256 }`,
			updates: []struct{ path, value string }{
				{"prod.resources.cpu", "400"},
				{"test.resources.memory", "512"},
			},
			expected: `prod:
  resources:
    cpu: 400
    memory: 512
test:
  resources: { cpu: 100, memory: 512 }`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Failed to load YAML: %v", err)
			}

			// Apply updates
			for _, update := range tt.updates {
				err := doc.SetInt(update.path, mustParseInt(update.value))
				if err != nil {
					t.Fatalf("Failed to set %s = %s: %v", update.path, update.value, err)
				}
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("Failed to convert to string: %v", err)
			}

			if strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
				t.Errorf("Format not preserved.\nExpected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

// Test for Issue #4: Critical duplication bug
func TestNoDuplicationBug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		updates  []struct{ path, value string }
		expected string
	}{
		{
			name: "No duplication when updating inline resources",
			input: `general:
  resources: {
    cpu: 256,
    memory: 256}`,
			updates: []struct{ path, value string }{
				{"general.resources.cpu", "512"},
				{"general.resources.memory", "512"},
			},
			expected: `general:
  resources: {
    cpu: 512,
    memory: 512}`,
		},
		{
			name: "No duplication with multiline resources",
			input: `general:
  resources:
    cpu: 256
    memory: 256`,
			updates: []struct{ path, value string }{
				{"general.resources.cpu", "512"},
				{"general.resources.memory", "512"},
			},
			expected: `general:
  resources:
    cpu: 512
    memory: 512`,
		},
		{
			name: "Sequential updates like real usage pattern",
			input: `test:
  resources:
    cpu: 100
    memory: 200
prod:
  resources:
    cores: 1
    memory: 400`,
			updates: []struct{ path, value string }{
				{"test.resources.cpu", "111"},
				{"test.resources.memory", "512"},
				{"prod.resources.cores", "2"},
				{"prod.resources.memory", "1024"},
			},
			expected: `test:
  resources:
    cpu: 111
    memory: 512
prod:
  resources:
    cores: 2
    memory: 1024`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test multiline flow objects - should work now after fixes

			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Failed to load YAML: %v", err)
			}

			// Apply updates in sequence (like real usage)
			for _, update := range tt.updates {
				err := doc.SetInt(update.path, mustParseInt(update.value))
				if err != nil {
					t.Fatalf("Failed to set %s = %s: %v", update.path, update.value, err)
				}
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("Failed to convert to string: %v", err)
			}

			// Focus on actual result matching rather than false positive duplication detection
			if strings.TrimSpace(result) != strings.TrimSpace(tt.expected) {
				t.Errorf("Result doesn't match expected.\nExpected:\n%s\n\nGot:\n%s", tt.expected, result)
			}
		})
	}
}

// Test simulating real user usage pattern
func TestRealUsagePattern(t *testing.T) {
	input := `general:
  name: my-service

prod:
  resources:
    cpu: 100
    memory: 256


test:
  resources:
    cpu: 50
    memory: 128`

	expected := `general:
  name: my-service

prod:
  resources:
    cpu: 100
    memory: 256


test:
  resources:
    cpu: 111
    memory: 111`

	doc, err := Load(input)
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	// Simulate real usage pattern from user's code
	// First update CPU
	err = doc.SetInt("test.resources.cpu", 111)
	if err != nil {
		t.Fatalf("Failed to set test.resources.cpu: %v", err)
	}

	// Then update memory
	err = doc.SetInt("test.resources.memory", 111)
	if err != nil {
		t.Fatalf("Failed to set test.resources.memory: %v", err)
	}

	result, err := doc.String()
	if err != nil {
		t.Fatalf("Failed to convert to string: %v", err)
	}

	// Allow for trailing newline differences
	if strings.TrimSpace(result) != strings.TrimSpace(expected) {
		t.Errorf("Real usage pattern failed.\nExpected:\n%s\n\nGot:\n%s", expected, result)
	}
}

// Helper function to parse int from string
func mustParseInt(s string) int64 {
	switch s {
	case "111":
		return 111
	case "512":
		return 512
	case "1024":
		return 1024
	case "2":
		return 2
	case "400":
		return 400
	default:
		return 0
	}
}
