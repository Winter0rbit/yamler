package main

import (
	"fmt"
	"log"

	"github.com/Winter0rbit/yamler"
)

func main() {
	// Original YAML with comments
	input := `database:
  host: localhost # Primary database host
  port: 5432      # Standard PostgreSQL port
  timeout: 30     # Connection timeout in seconds
app:
  name: myapp # Application name
  debug: false    # Debug mode flag
  version: 1.0    # Current version`

	fmt.Println("=== Original YAML ===")
	fmt.Println(input)

	// Load document
	doc, err := yamler.Load(input)
	if err != nil {
		log.Fatal(err)
	}

	// Example 1: Relative alignment (default)
	fmt.Println("\n=== 1. Relative alignment (preserves original spacing) ===")
	doc.EnableRelativeCommentAlignment()
	doc.Set("app.debug", true) // Change to demonstrate
	result, _ := doc.String()
	fmt.Println(result)

	// Example 2: Absolute alignment to column 25
	fmt.Println("=== 2. Absolute alignment to column 25 ===")
	doc.SetAbsoluteCommentAlignment(25)
	doc.Set("app.version", "2.0") // Another change
	result, _ = doc.String()
	fmt.Println(result)

	// Example 3: Absolute alignment to column 30
	fmt.Println("=== 3. Absolute alignment to column 30 ===")
	doc.SetAbsoluteCommentAlignment(30)
	result, _ = doc.String()
	fmt.Println(result)

	// Example 4: Disable comments
	fmt.Println("=== 4. Disable comments ===")
	doc.DisableCommentAlignment()
	result, _ = doc.String()
	fmt.Println(result)

	// Example 5: Return to relative alignment
	fmt.Println("=== 5. Return to relative alignment ===")
	doc.EnableRelativeCommentAlignment()
	result, _ = doc.String()
	fmt.Println(result)
}
