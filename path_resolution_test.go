package yamler

import (
	"strings"
	"testing"
)

func TestPathResolutionIssues(t *testing.T) {
	inputYAML := `name: shiva-generator

general:
  resources:
    cpu: 512
    memory: 512


prod:
  trace_project: vertispaas
  datacenters:
    sas: { count: 2 }
    vla: { count: 2 }
  resources:
    cpu: 512
    memory: 512
  config:
    files:
      - shiva/generator/common.include.yml
      - shiva/generator/prod.include.yml

test:
  datacenters:
    sas: { count: 1 }
    vla: { count: 1 }
  resources: {
    cpu: 512,
    memory: 512}
  config:
    files:
      - shiva/generator/common.include.yml
      - shiva/generator/test.include.yml`

	t.Run("Path resolution should not modify wrong section", func(t *testing.T) {
		doc, err := LoadBytes([]byte(inputYAML))
		if err != nil {
			t.Fatalf("Failed to load YAML: %v", err)
		}

		// These updates should ONLY affect test.resources, not general.resources
		err = doc.SetInt("test.resources.cpu", 111)
		if err != nil {
			t.Fatalf("Failed to set test.resources.cpu: %v", err)
		}

		err = doc.SetInt("test.resources.memory", 111)
		if err != nil {
			t.Fatalf("Failed to set test.resources.memory: %v", err)
		}

		result, err := doc.ToBytes()
		if err != nil {
			t.Fatalf("Failed to convert to bytes: %v", err)
		}

		resultStr := string(result)

		// Check that general.resources was NOT modified
		if strings.Contains(resultStr, "general:\n  resources: {") {
			t.Errorf("ERROR: general.resources was incorrectly modified!\nResult:\n%s", resultStr)
		}

		// Check for duplication in general section
		generalSection := extractSection(resultStr, "general:")
		if strings.Count(generalSection, "cpu:") > 1 {
			t.Errorf("ERROR: Found duplicate cpu entries in general section!\nGeneral section:\n%s", generalSection)
		}

		// Check that test.resources was correctly modified
		testValue, err := doc.GetInt("test.resources.cpu")
		if err != nil {
			t.Errorf("Failed to read test.resources.cpu: %v", err)
		} else if testValue != 111 {
			t.Errorf("test.resources.cpu = %d, want 111", testValue)
		}

		// Check that general.resources was NOT modified
		generalValue, err := doc.GetInt("general.resources.cpu")
		if err != nil {
			t.Errorf("Failed to read general.resources.cpu: %v", err)
		} else if generalValue != 512 {
			t.Errorf("general.resources.cpu = %d, want 512 (should not be modified)", generalValue)
		}
	})

	t.Run("Should preserve trailing newline presence", func(t *testing.T) {
		// Test with no trailing newline
		inputNoNewline := "name: test\nvalue: 1"
		doc, err := LoadBytes([]byte(inputNoNewline))
		if err != nil {
			t.Fatalf("Failed to load YAML: %v", err)
		}

		err = doc.SetString("name", "updated")
		if err != nil {
			t.Fatalf("Failed to set name: %v", err)
		}

		result, err := doc.ToBytes()
		if err != nil {
			t.Fatalf("Failed to convert to bytes: %v", err)
		}

		resultStr := string(result)
		if strings.HasSuffix(resultStr, "\n") {
			t.Errorf("Added unwanted trailing newline. Input had no newline, output should not have one either.\nResult: %q", resultStr)
		}

		// Test with trailing newline
		inputWithNewline := "name: test\nvalue: 1\n"
		doc2, err := LoadBytes([]byte(inputWithNewline))
		if err != nil {
			t.Fatalf("Failed to load YAML: %v", err)
		}

		err = doc2.SetString("name", "updated")
		if err != nil {
			t.Fatalf("Failed to set name: %v", err)
		}

		result2, err := doc2.ToBytes()
		if err != nil {
			t.Fatalf("Failed to convert to bytes: %v", err)
		}

		resultStr2 := string(result2)
		if !strings.HasSuffix(resultStr2, "\n") {
			t.Errorf("Lost trailing newline. Input had newline, output should preserve it.\nResult: %q", resultStr2)
		}
	})
}

// Helper function to extract a section from YAML
func extractSection(yaml, sectionName string) string {
	lines := strings.Split(yaml, "\n")
	var sectionLines []string
	inSection := false
	sectionIndent := -1

	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), sectionName) {
			inSection = true
			sectionIndent = len(line) - len(strings.TrimLeft(line, " \t"))
			sectionLines = append(sectionLines, line)
			continue
		}

		if inSection {
			currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))
			// If line is less indented than section or is another top-level key, end section
			if strings.TrimSpace(line) != "" && currentIndent <= sectionIndent {
				break
			}
			sectionLines = append(sectionLines, line)
		}
	}

	return strings.Join(sectionLines, "\n")
}
