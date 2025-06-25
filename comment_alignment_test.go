package yamler

import (
	"fmt"
	"testing"
)

func TestCommentAlignmentModes(t *testing.T) {
	input := `config:
  database:
    host: localhost # Primary host
    port: 5432      # Standard PostgreSQL port
    timeout: 30     # Connection timeout
  app:
    name: myapp # Application name
    debug: false    # Debug mode
    version: 1.0    # Version number`

	tests := []struct {
		name     string
		mode     CommentAlignmentMode
		column   int
		expected string
	}{
		{
			name: "relative_alignment",
			mode: CommentAlignmentRelative,
			expected: `config:
  database:
    host: localhost # Primary host
    port: 5432      # Standard PostgreSQL port
    timeout: 30     # Connection timeout
  app:
    name: myapp # Application name
    debug: true    # Debug mode
    version: 1.0    # Version number
`,
		},
		{
			name:   "absolute_alignment_column_25",
			mode:   CommentAlignmentAbsolute,
			column: 25,
			expected: `config:
  database:
    host: localhost      # Primary host
    port: 5432           # Standard PostgreSQL port
    timeout: 30          # Connection timeout
  app:
    name: myapp          # Application name
    debug: true          # Debug mode
    version: 1.0         # Version number
`,
		},
		{
			name: "disabled_alignment",
			mode: CommentAlignmentDisabled,
			expected: `config:
  database:
    host: localhost
    port: 5432
    timeout: 30
  app:
    name: myapp
    debug: true
    version: 1.0
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Load(input)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Configure alignment mode
			if tt.mode == CommentAlignmentAbsolute && tt.column > 0 {
				doc.SetAbsoluteCommentAlignment(tt.column)
			} else {
				doc.SetCommentAlignment(tt.mode)
			}

			// Make a change to trigger formatting
			err = doc.Set("config.app.debug", true)
			if err != nil {
				t.Fatalf("Set() error = %v", err)
			}

			result, err := doc.String()
			if err != nil {
				t.Fatalf("String() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("Comment alignment not as expected.\nGot:\n%s\nWant:\n%s", result, tt.expected)
			}
		})
	}
}

func ExampleDocument_SetAbsoluteCommentAlignment() {
	input := `database:
  host: localhost # Primary host
  port: 5432      # Standard PostgreSQL port
app:
  name: myapp # Application name
  debug: false    # Debug mode`

	doc, _ := Load(input)

	// Align all comments to column 20
	doc.SetAbsoluteCommentAlignment(20)

	// Make a change
	doc.Set("app.debug", true)

	result, _ := doc.String()
	fmt.Print(result)
	// Output:
	// database:
	//   host: localhost   # Primary host
	//   port: 5432        # Standard PostgreSQL port
	// app:
	//   name: myapp       # Application name
	//   debug: true       # Debug mode
}

func ExampleDocument_EnableRelativeCommentAlignment() {
	input := `database:
  host: localhost # Primary host
  port: 5432      # Standard PostgreSQL port
app:
  name: myapp # Application name
  debug: false    # Debug mode`

	doc, _ := Load(input)

	// Preserve original spacing (default behavior)
	doc.EnableRelativeCommentAlignment()

	// Make a change
	doc.Set("app.debug", true)

	result, _ := doc.String()
	fmt.Print(result)
	// Output:
	// database:
	//   host: localhost # Primary host
	//   port: 5432      # Standard PostgreSQL port
	// app:
	//   name: myapp # Application name
	//   debug: true    # Debug mode
}

func ExampleDocument_DisableCommentAlignment() {
	input := `database:
  host: localhost # Primary host
  port: 5432      # Standard PostgreSQL port
app:
  name: myapp # Application name
  debug: false    # Debug mode`

	doc, _ := Load(input)

	// Disable comment alignment processing
	doc.DisableCommentAlignment()

	// Make a change
	doc.Set("app.debug", true)

	result, _ := doc.String()
	fmt.Print(result)
	// Output:
	// database:
	//   host: localhost
	//   port: 5432
	// app:
	//   name: myapp
	//   debug: true
}
