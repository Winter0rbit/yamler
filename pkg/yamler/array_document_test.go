package yamler

import (
	"testing"
)

// TestArrayDocumentSupport tests support for array document roots like Ansible playbooks
func TestArrayDocumentSupport(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		operation      func(*Document) error
		expectedOutput string
	}{
		{
			name: "ansible_playbook_style",
			input: `---
- name: Configure web servers
  hosts: webservers
  become: yes
  vars:
    http_port: 80
    max_clients: 200
  tasks:
    - name: Install nginx
      package:
        name: nginx
        state: present
- name: Configure database
  hosts: dbservers
  become: yes
  vars:
    db_port: 5432`,
			operation: func(d *Document) error {
				return d.SetArrayElement(0, "vars.max_clients", 500)
			},
			expectedOutput: `- name: Configure web servers
  hosts: webservers
  become: yes
  vars:
    http_port: 80
    max_clients: 500
  tasks:
    - name: Install nginx
      package:
        name: nginx
        state: present
- name: Configure database
  hosts: dbservers
  become: yes
  vars:
    db_port: 5432
`,
		},
		{
			name: "simple_array_document",
			input: `---
- item1
- item2
- item3`,
			operation: func(d *Document) error {
				return d.SetArrayElement(1, "", "updated_item2")
			},
			expectedOutput: `- item1
- updated_item2
- item3
`,
		},
		{
			name: "add_element_to_array_document",
			input: `---
- task1
- task2`,
			operation: func(d *Document) error {
				return d.AddArrayElement("task3")
			},
			expectedOutput: `- task1
- task2
- task3
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(tt.input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Check if document is correctly identified as array root
			if !doc.isArrayRoot() {
				t.Errorf("Document should be identified as array root")
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
				t.Errorf("Array document operation failed.\nGot:\n%s\nWant:\n%s", result, tt.expectedOutput)
			}
		})
	}
}

// TestArrayDocumentGetters tests getting values from array documents
func TestArrayDocumentGetters(t *testing.T) {
	input := `---
- name: task1
  action: command
  vars:
    timeout: 30
- name: task2
  action: shell
  vars:
    timeout: 60`

	doc, err := Load(input)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test getting entire element
	element, err := doc.GetArrayDocumentElement(0, "")
	if err != nil {
		t.Fatalf("GetArrayDocumentElement() error = %v", err)
	}

	elementMap, ok := element.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", element)
	}

	if elementMap["name"] != "task1" {
		t.Errorf("Expected name=task1, got %v", elementMap["name"])
	}

	// Test getting specific path
	timeout, err := doc.GetArrayDocumentElement(1, "vars.timeout")
	if err != nil {
		t.Fatalf("GetArrayDocumentElement() error = %v", err)
	}

	// Convert to int64 for comparison since YAML parses numbers as int64
	timeoutInt, ok := timeout.(int64)
	if !ok {
		t.Fatalf("Expected int64, got %T with value %v", timeout, timeout)
	}

	if timeoutInt != 60 {
		t.Errorf("Expected timeout=60, got %v", timeoutInt)
	}
}
