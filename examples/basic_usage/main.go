package main

import (
	"fmt"
	"log"

	"github.com/Winter0rbit/yamler"
)

func main() {
	// Basic YAML manipulation example
	yamlContent := `
app:
  name: myapp
  version: 1.0
  debug: false
database:
  host: localhost
  port: 5432
  credentials:
    username: admin
    password: secret123
servers:
  - name: web1
    ip: 192.168.1.10
  - name: web2
    ip: 192.168.1.11`

	fmt.Println("=== Original YAML ===")
	fmt.Println(yamlContent)

	// Load the YAML document
	doc, err := yamler.Load(yamlContent)
	if err != nil {
		log.Fatal("Failed to load YAML:", err)
	}

	// Basic get operations
	fmt.Println("\n=== Reading Values ===")

	appName, _ := doc.GetString("app.name")
	fmt.Printf("App name: %s\n", appName)

	dbPort, _ := doc.GetInt("database.port")
	fmt.Printf("Database port: %d\n", dbPort)

	isDebug, _ := doc.GetBool("app.debug")
	fmt.Printf("Debug mode: %t\n", isDebug)

	// Basic set operations
	fmt.Println("\n=== Updating Values ===")

	doc.Set("app.version", "2.0")
	doc.Set("app.debug", true)
	doc.Set("database.host", "db.example.com")
	doc.Set("database.credentials.password", "newpassword456")

	// Add new configuration
	doc.Set("app.environment", "production")
	doc.Set("monitoring.enabled", true)
	doc.Set("monitoring.port", 9090)

	// Working with arrays
	fmt.Println("\n=== Array Operations ===")

	// Get array length
	serverCount, _ := doc.GetArrayLength("servers")
	fmt.Printf("Number of servers: %d\n", serverCount)

	// Add new server
	newServer := map[string]interface{}{
		"name": "web3",
		"ip":   "192.168.1.12",
	}
	doc.AppendToArray("servers", newServer)

	// Update existing server
	doc.Set("servers[0].ip", "192.168.1.100")

	// Display final result
	fmt.Println("\n=== Final YAML ===")
	result, err := doc.String()
	if err != nil {
		log.Fatal("Failed to convert to string:", err)
	}
	fmt.Println(result)
}
