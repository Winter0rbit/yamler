package main

import (
	"fmt"
	"log"

	"github.com/Winter0rbit/yamler"
)

func main() {
	// Simple example using the new clean import path
	yamlContent := `
name: my-app
version: 1.0.0
config:
  database:
    host: localhost
    port: 5432
  features:
    - authentication
    - logging
    - metrics
`

	// Load YAML
	doc, err := yamler.Load(yamlContent)
	if err != nil {
		log.Fatal(err)
	}

	// Read values
	name, _ := doc.GetString("name")
	fmt.Printf("App name: %s\n", name)

	host, _ := doc.GetString("config.database.host")
	fmt.Printf("Database host: %s\n", host)

	port, _ := doc.GetInt("config.database.port")
	fmt.Printf("Database port: %d\n", port)

	// Modify values
	doc.SetString("config.database.host", "production-db.example.com")
	doc.SetInt("config.database.port", 5433)
	doc.AppendToArray("config.features", "monitoring")

	// Output result
	result, _ := doc.String()
	fmt.Println("\nUpdated YAML:")
	fmt.Println(result)
}
