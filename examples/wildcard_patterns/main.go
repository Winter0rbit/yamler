package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Winter0rbit/yamler"
)

func main() {
	// Complex configuration with multiple environments
	config := `environments:
  development:
    database:
      host: dev-db.local
      port: 5432
      debug: true
      timeout: 30
    api:
      host: dev-api.local
      port: 3000
      debug: true
      timeout: 10
    cache:
      host: dev-cache.local
      port: 6379
      debug: true
      timeout: 5
  
  staging:
    database:
      host: staging-db.example.com
      port: 5432
      debug: false
      timeout: 60
    api:
      host: staging-api.example.com
      port: 3000
      debug: false
      timeout: 20
    cache:
      host: staging-cache.example.com
      port: 6379
      debug: false
      timeout: 10
  
  production:
    database:
      host: prod-db.example.com
      port: 5432
      debug: false
      timeout: 120
    api:
      host: prod-api.example.com
      port: 3000
      debug: false
      timeout: 30
    cache:
      host: prod-cache.example.com
      port: 6379
      debug: false
      timeout: 15

global:
  monitoring:
    enabled: true
    port: 9090
  logging:
    level: info
    format: json
  security:
    ssl_enabled: true
    cors_enabled: true`

	fmt.Println("=== Original Configuration ===")
	fmt.Println(config)

	// Load the configuration
	doc, err := yamler.Load(config)
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Working with wildcard patterns
	fmt.Println("\n=== Working with Wildcard Patterns ===")

	// Get all debug settings across all environments
	fmt.Println("\n1. Find all debug settings:")
	debugPaths, err := doc.GetKeys("**.debug")
	if err != nil {
		log.Printf("Error getting debug paths: %v", err)
	} else {
		for _, path := range debugPaths {
			value, _ := doc.Get(path)
			fmt.Printf("  %s: %v\n", path, value)
		}
	}

	// Get all database hosts
	fmt.Println("\n2. Find all database hosts:")
	dbHosts, err := doc.GetAll("**.database.host")
	if err != nil {
		log.Printf("Error getting database hosts: %v", err)
	} else {
		for path, value := range dbHosts {
			fmt.Printf("  %s: %v\n", path, value)
		}
	}

	// Get all timeout values
	fmt.Println("\n3. Find all timeout values:")
	timeouts, err := doc.GetAll("**.timeout")
	if err != nil {
		log.Printf("Error getting timeouts: %v", err)
	} else {
		for path, value := range timeouts {
			fmt.Printf("  %s: %v\n", path, value)
		}
	}

	// Bulk operations using wildcards
	fmt.Println("\n=== Bulk Operations ===")

	// Enable debug mode for all development services
	fmt.Println("\n1. Enabling debug for all development services:")
	doc.SetAll("environments.development.*.debug", true)

	// Update all production timeouts
	fmt.Println("2. Updating production timeouts:")
	doc.SetAll("environments.production.*.timeout", 180)

	// Disable debug for all staging and production
	fmt.Println("3. Disabling debug for staging and production:")
	doc.SetAll("environments.staging.*.debug", false)
	doc.SetAll("environments.production.*.debug", false)

	// Update all API ports to 8080
	fmt.Println("4. Updating all API ports:")
	doc.SetAll("**.api.port", 8080)

	// Add monitoring to all services
	fmt.Println("\n=== Adding Monitoring to All Services ===")

	// Get all service paths and extract unique service types
	allPaths, err := doc.GetPathsRecursive()
	if err != nil {
		log.Printf("Error getting paths: %v", err)
	} else {
		uniqueServices := make(map[string]bool)

		for _, path := range allPaths {
			// Parse paths like "environments.development.database.host"
			parts := strings.Split(path, ".")
			if len(parts) >= 3 && parts[0] == "environments" {
				serviceType := parts[2] // database, api, cache
				if serviceType == "database" || serviceType == "api" || serviceType == "cache" {
					uniqueServices[serviceType] = true
				}
			}
		}

		// Add monitoring configuration to each service type in each environment
		for serviceType := range uniqueServices {
			pattern := fmt.Sprintf("environments.*.%s.monitoring", serviceType)
			monitoringConfig := map[string]interface{}{
				"enabled": true,
				"port":    9090,
				"path":    "/metrics",
			}
			doc.SetAll(pattern, monitoringConfig)
		}
	}

	// Working with specific patterns
	fmt.Println("\n=== Pattern Matching Examples ===")

	// Find all ports
	fmt.Println("\n1. All ports:")
	ports, err := doc.GetAll("**.port")
	if err != nil {
		log.Printf("Error getting ports: %v", err)
	} else {
		for path, value := range ports {
			fmt.Printf("  %s: %v\n", path, value)
		}
	}

	// Find all hosts in production
	fmt.Println("\n2. All production hosts:")
	prodHosts, err := doc.GetAll("environments.production.*.host")
	if err != nil {
		log.Printf("Error getting production hosts: %v", err)
	} else {
		for path, value := range prodHosts {
			fmt.Printf("  %s: %v\n", path, value)
		}
	}

	// Find all cache configurations
	fmt.Println("\n3. All cache configurations:")
	cacheConfigs, err := doc.GetAll("**.cache.*")
	if err != nil {
		log.Printf("Error getting cache configs: %v", err)
	} else {
		for path, value := range cacheConfigs {
			fmt.Printf("  %s: %v\n", path, value)
		}
	}

	// Advanced pattern operations
	fmt.Println("\n=== Advanced Pattern Operations ===")

	// Count services per environment
	environments := []string{"development", "staging", "production"}
	for _, env := range environments {
		pattern := fmt.Sprintf("environments.%s.*", env)
		services, err := doc.GetKeys(pattern)
		if err != nil {
			log.Printf("Error getting services for %s: %v", env, err)
		} else {
			fmt.Printf("%s environment has %d services\n", env, len(services))
		}
	}

	// Find all enabled debug modes
	debugEnabled, err := doc.GetAll("**.debug")
	if err != nil {
		log.Printf("Error getting debug settings: %v", err)
	} else {
		enabledCount := 0
		for _, value := range debugEnabled {
			if debug, ok := value.(bool); ok && debug {
				enabledCount++
			}
		}
		fmt.Printf("Debug is enabled in %d services\n", enabledCount)
	}

	// Display final configuration
	fmt.Println("\n=== Final Configuration ===")
	result, err := doc.String()
	if err != nil {
		log.Fatal("Failed to convert to string:", err)
	}
	fmt.Println(result)
}
