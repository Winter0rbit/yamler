package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Winter0rbit/yamler"
)

func main() {
	fmt.Println("=== Yamler Basic Usage Examples ===\n")

	// Example 1: Basic Operations with Format Preservation
	basicOperations()

	// Example 2: Type-Safe Operations
	typeSafeOperations()

	// Example 3: Comment Alignment Features
	commentAlignmentDemo()

	// Example 4: Performance Features Demo
	performanceDemo()

	// Example 5: Complex Flow Objects
	complexFlowDemo()
}

func basicOperations() {
	fmt.Println("1. Basic Operations with Format Preservation")
	fmt.Println(strings.Repeat("=", 50))

	yamlContent := `# Application Configuration
app:
  name: myapp         # Application name
  version: "1.0"      # Current version
  debug: true         # Debug mode
  
  # Server settings
  server:
    host: localhost   # Server host
    port: 8080       # Server port
    
  # Feature flags
  features:
    - authentication
    - logging
    - metrics        # Performance tracking
`

	doc, err := yamler.Load(yamlContent)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original YAML:")
	fmt.Println(doc.String())

	// Modify values while preserving formatting
	doc.Set("app.version", "2.0")
	doc.SetBool("app.debug", false)
	doc.SetInt("app.server.port", 9090)
	doc.AppendToArray("app.features", "monitoring")

	fmt.Println("\nAfter modifications (formatting preserved!):")
	fmt.Println(doc.String())
	fmt.Println()
}

func typeSafeOperations() {
	fmt.Println("2. Type-Safe Operations")
	fmt.Println(strings.Repeat("=", 30))

	yamlContent := `
database:
  host: localhost
  port: 5432
  timeout: 30.5
  ssl: yes
  pools: [primary, replica, analytics]
  config:
    max_connections: 100
    retry_attempts: 3
`

	doc, err := yamler.Load(yamlContent)
	if err != nil {
		log.Fatal(err)
	}

	// Type-safe getters
	host, _ := doc.GetString("database.host")
	port, _ := doc.GetInt("database.port")
	timeout, _ := doc.GetFloat("database.timeout")
	ssl, _ := doc.GetBool("database.ssl")
	pools, _ := doc.GetStringSlice("database.pools")
	config, _ := doc.GetMap("database.config")

	fmt.Printf("Host: %s (type: %T)\n", host, host)
	fmt.Printf("Port: %d (type: %T)\n", port, port)
	fmt.Printf("Timeout: %.1f (type: %T)\n", timeout, timeout)
	fmt.Printf("SSL: %v (type: %T)\n", ssl, ssl)
	fmt.Printf("Pools: %v (type: %T)\n", pools, pools)
	fmt.Printf("Config: %v (type: %T)\n", config, config)

	// Type-safe setters
	doc.SetString("database.host", "prod-db.example.com")
	doc.SetInt("database.port", 3306)
	doc.SetFloat("database.timeout", 60.0)
	doc.SetBool("database.ssl", true)
	doc.SetStringSlice("database.pools", []string{"primary", "replica"})

	fmt.Println("\nAfter type-safe modifications:")
	fmt.Println(doc.String())
	fmt.Println()
}

func commentAlignmentDemo() {
	fmt.Println("3. Comment Alignment Features")
	fmt.Println(strings.Repeat("=", 35))

	yamlContent := `
app:
  name: myapp    # App name
  port: 8080        # Port number
  debug: true # Debug flag
  timeout: 30     # Timeout in seconds
`

	doc, err := yamler.Load(yamlContent)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original (relative alignment):")
	fmt.Println(doc.String())

	// Absolute alignment at column 25
	doc.SetAbsoluteCommentAlignment(25)
	fmt.Println("Absolute alignment at column 25:")
	fmt.Println(doc.String())

	// Disable comments
	doc.DisableCommentAlignment()
	fmt.Println("Comments disabled:")
	fmt.Println(doc.String())

	// Re-enable relative alignment
	doc.EnableRelativeCommentAlignment()
	fmt.Println("Relative alignment restored:")
	fmt.Println(doc.String())
	fmt.Println()
}

func performanceDemo() {
	fmt.Println("4. Performance Features Demo")
	fmt.Println(strings.Repeat("=", 35))

	// Create a larger YAML for performance testing
	yamlContent := `
services:
`
	for i := 0; i < 100; i++ {
		yamlContent += fmt.Sprintf(`  service%d:
    name: app%d
    port: %d
    debug: true
    replicas: 3
`, i, i, 8000+i)
	}

	doc, err := yamler.Load(yamlContent)
	if err != nil {
		log.Fatal(err)
	}

	// Measure performance of repeated operations
	start := time.Now()
	for i := 0; i < 50; i++ {
		// These operations benefit from caching
		doc.GetString(fmt.Sprintf("services.service%d.name", i))
		doc.GetInt(fmt.Sprintf("services.service%d.port", i))
		doc.GetBool(fmt.Sprintf("services.service%d.debug", i))
	}
	duration := time.Since(start)

	fmt.Printf("150 path operations completed in: %v\n", duration)
	fmt.Printf("Average per operation: %v\n", duration/150)

	// Bulk operations with wildcards (very efficient)
	start = time.Now()
	doc.SetAll("services.*.debug", false)
	doc.SetAll("services.*.replicas", 5)
	bulkDuration := time.Since(start)

	fmt.Printf("200 bulk operations completed in: %v\n", bulkDuration)
	fmt.Printf("Bulk operations are ~%.1fx faster than individual operations\n",
		float64(duration)/float64(bulkDuration))
	fmt.Println()
}

func complexFlowDemo() {
	fmt.Println("5. Complex Flow Objects")
	fmt.Println(strings.Repeat("=", 30))

	yamlContent := `
# Complex nested flow structures
matrix: [
  [1, 2, 3],
  [4, 5, 6],
  [7, 8, 9]
]

metadata: {
  created: 2023-01-01,
  author: developer,
  tags: [yaml, config, test],
  nested: {
    level1: {
      level2: [a, b, c]
    }
  }
}

# Mixed styles
config:
  inline_array: [1, 2, 3]
  block_array:
    - item1
    - item2
  inline_object: {key: value, number: 42}
  block_object:
    key1: value1
    key2: value2
`

	doc, err := yamler.Load(yamlContent)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original complex structure:")
	fmt.Println(doc.String())

	// Modify complex structures
	doc.UpdateArrayElement("matrix", 0, []int{10, 20, 30})
	doc.Set("metadata.version", "2.0")
	doc.AppendToArray("metadata.tags", "production")
	doc.Set("metadata.nested.level1.level3", "new_level")

	fmt.Println("After complex modifications (structure preserved!):")
	fmt.Println(doc.String())
	fmt.Println()
}
