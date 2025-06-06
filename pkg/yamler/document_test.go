package yamler

import (
	"os"
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
			name: "complex yaml",
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
