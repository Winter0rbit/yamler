package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/Winter0rbit/yamler"
)

func main() {
	// Create a temporary directory for our examples
	tempDir, err := os.MkdirTemp("", "yamler_examples")
	if err != nil {
		log.Fatal("Failed to create temp directory:", err)
	}
	defer os.RemoveAll(tempDir) // Clean up

	fmt.Printf("Working in temporary directory: %s\n", tempDir)

	// Example 1: Create and save a new YAML file
	fmt.Println("\n=== Creating New Configuration File ===")

	configPath := filepath.Join(tempDir, "app_config.yml")

	// Create new document
	doc, err := yamler.Load("")
	if err != nil {
		log.Fatal("Failed to create new document:", err)
	}

	// Build configuration
	doc.Set("app.name", "MyApplication")
	doc.Set("app.version", "1.0.0")
	doc.Set("app.environment", "development")

	doc.Set("database.host", "localhost")
	doc.Set("database.port", 5432)
	doc.Set("database.name", "myapp_dev")
	doc.Set("database.ssl", false)

	doc.Set("redis.host", "localhost")
	doc.Set("redis.port", 6379)
	doc.Set("redis.db", 0)

	// Add array of servers
	servers := []map[string]interface{}{
		{"name": "web1", "ip": "192.168.1.10", "role": "frontend"},
		{"name": "web2", "ip": "192.168.1.11", "role": "frontend"},
		{"name": "api1", "ip": "192.168.1.20", "role": "backend"},
	}
	doc.Set("servers", servers)

	// Save the file
	err = doc.Save(configPath)
	if err != nil {
		log.Fatal("Failed to save config:", err)
	}
	fmt.Printf("Created configuration file: %s\n", configPath)

	// Example 2: Load and modify existing file
	fmt.Println("\n=== Loading and Modifying Existing File ===")

	// Load the file we just created
	loadedDoc, err := yamler.LoadFile(configPath)
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Read some values
	appName, _ := loadedDoc.GetString("app.name")
	dbPort, _ := loadedDoc.GetInt("database.port")
	fmt.Printf("Loaded app: %s, DB port: %d\n", appName, dbPort)

	// Modify configuration for production
	loadedDoc.Set("app.environment", "production")
	loadedDoc.Set("database.host", "prod-db.example.com")
	loadedDoc.Set("database.ssl", true)
	loadedDoc.Set("database.name", "myapp_prod")

	// Add production-specific settings
	loadedDoc.Set("logging.level", "warn")
	loadedDoc.Set("logging.file", "/var/log/myapp.log")
	loadedDoc.Set("monitoring.enabled", true)
	loadedDoc.Set("monitoring.port", 9090)

	// Add new server
	newServer := map[string]interface{}{
		"name": "lb1",
		"ip":   "192.168.1.5",
		"role": "loadbalancer",
	}
	loadedDoc.AppendToArray("servers", newServer)

	// Save as production config
	prodConfigPath := filepath.Join(tempDir, "app_config_prod.yml")
	err = loadedDoc.Save(prodConfigPath)
	if err != nil {
		log.Fatal("Failed to save production config:", err)
	}
	fmt.Printf("Created production config: %s\n", prodConfigPath)

	// Example 3: Working with multiple configuration files
	fmt.Println("\n=== Working with Multiple Configuration Files ===")

	// Create database-specific config
	dbConfigPath := filepath.Join(tempDir, "database.yml")
	dbDoc, _ := yamler.Load("")

	dbDoc.Set("connections.primary.host", "db1.example.com")
	dbDoc.Set("connections.primary.port", 5432)
	dbDoc.Set("connections.primary.database", "myapp")
	dbDoc.Set("connections.primary.pool_size", 20)

	dbDoc.Set("connections.readonly.host", "db2.example.com")
	dbDoc.Set("connections.readonly.port", 5432)
	dbDoc.Set("connections.readonly.database", "myapp")
	dbDoc.Set("connections.readonly.pool_size", 10)

	dbDoc.Set("migrations.enabled", true)
	dbDoc.Set("migrations.path", "./migrations")

	err = dbDoc.Save(dbConfigPath)
	if err != nil {
		log.Fatal("Failed to save database config:", err)
	}

	// Create logging config
	logConfigPath := filepath.Join(tempDir, "logging.yml")
	logDoc, _ := yamler.Load("")

	logDoc.Set("level", "info")
	logDoc.Set("format", "json")
	logDoc.Set("outputs", []string{"stdout", "file"})
	logDoc.Set("file.path", "/var/log/app.log")
	logDoc.Set("file.max_size", "100MB")
	logDoc.Set("file.max_backups", 5)

	err = logDoc.Save(logConfigPath)
	if err != nil {
		log.Fatal("Failed to save logging config:", err)
	}

	// Example 4: Merging configurations
	fmt.Println("\n=== Merging Configurations ===")

	// Load main config
	mainDoc, err := yamler.LoadFile(configPath)
	if err != nil {
		log.Fatal("Failed to load main config:", err)
	}

	// Load and merge database config
	dbConfigDoc, err := yamler.LoadFile(dbConfigPath)
	if err != nil {
		log.Fatal("Failed to load database config:", err)
	}

	// Merge database config into main config
	err = mainDoc.MergeAt("database", dbConfigDoc)
	if err != nil {
		log.Fatal("Failed to merge database config:", err)
	}

	// Load and merge logging config
	logConfigDoc, err := yamler.LoadFile(logConfigPath)
	if err != nil {
		log.Fatal("Failed to load logging config:", err)
	}

	err = mainDoc.MergeAt("logging", logConfigDoc)
	if err != nil {
		log.Fatal("Failed to merge logging config:", err)
	}

	// Save merged configuration
	mergedConfigPath := filepath.Join(tempDir, "merged_config.yml")
	err = mainDoc.Save(mergedConfigPath)
	if err != nil {
		log.Fatal("Failed to save merged config:", err)
	}

	// Example 5: Display all created files
	fmt.Println("\n=== Created Files ===")

	files := []string{
		configPath,
		prodConfigPath,
		dbConfigPath,
		logConfigPath,
		mergedConfigPath,
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			log.Printf("Error getting info for %s: %v", file, err)
			continue
		}

		fmt.Printf("\nFile: %s\n", filepath.Base(file))
		fmt.Printf("Size: %d bytes\n", info.Size())

		// Read and display first few lines
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Error reading %s: %v", file, err)
			continue
		}

		lines := string(content)
		if len(lines) > 200 {
			lines = lines[:200] + "..."
		}
		fmt.Printf("Content preview:\n%s\n", lines)
	}

	fmt.Printf("\nAll files created in: %s\n", tempDir)
	fmt.Println("Note: Temporary directory will be cleaned up automatically")
}
