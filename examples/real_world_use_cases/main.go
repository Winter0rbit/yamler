package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Winter0rbit/yamler"
)

func main() {
	fmt.Println("=== Yamler Real-World Use Cases ===\n")

	// Example 1: CI/CD Pipeline Configuration
	cicdPipelineExample()

	// Example 2: Multi-Environment Configuration Management
	multiEnvironmentExample()

	// Example 3: Infrastructure as Code
	infrastructureExample()

	// Example 4: Application Configuration Templating
	configTemplatingExample()

	// Example 5: Configuration Migration and Transformation
	configMigrationExample()
}

func cicdPipelineExample() {
	fmt.Println("1. CI/CD Pipeline Configuration")
	fmt.Println(strings.Repeat("=", 40))

	// GitHub Actions workflow
	githubWorkflow := `name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  NODE_VERSION: 16
  DOCKER_REGISTRY: ghcr.io

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: ${{ env.NODE_VERSION }}
      - run: npm ci
      - run: npm test

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Build Docker image
        run: docker build -t ${{ env.DOCKER_REGISTRY }}/myapp:${{ github.sha }} .
`

	doc, err := yamler.Load(githubWorkflow)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Original GitHub Actions workflow:")
	fmt.Println(doc.String())

	// Dynamically add deployment job for production
	deployJob := map[string]interface{}{
		"needs":       []string{"test", "build"},
		"runs-on":     "ubuntu-latest",
		"if":          "github.ref == 'refs/heads/main'",
		"environment": "production",
		"steps": []map[string]interface{}{
			{
				"uses": "actions/checkout@v3",
			},
			{
				"name": "Deploy to production",
				"run":  "kubectl apply -f k8s/",
				"env": map[string]string{
					"KUBECONFIG": "${{ secrets.KUBECONFIG }}",
				},
			},
		},
	}

	doc.Set("jobs.deploy", deployJob)

	// Add environment variables for different branches
	doc.Set("env.BUILD_ENV", "production")
	doc.AppendToArray("on.push.branches", "release/*")

	fmt.Println("\nAfter adding deployment job:")
	fmt.Println(doc.String())
	fmt.Println()
}

func multiEnvironmentExample() {
	fmt.Println("2. Multi-Environment Configuration Management")
	fmt.Println(strings.Repeat("=", 50))

	// Base application configuration
	baseConfig := `# Application Base Configuration
app:
  name: myapp
  version: 1.0.0
  
server:
  host: localhost
  port: 8080
  timeout: 30s
  
database:
  host: localhost
  port: 5432
  name: myapp_db
  pool_size: 10
  ssl: false
  
logging:
  level: info
  format: json
  
features:
  - authentication
  - api_v1
`

	baseDoc, _ := yamler.Load(baseConfig)

	// Development overrides
	fmt.Println("Development environment configuration:")
	devDoc := baseDoc // Copy for development
	devDoc.Set("server.port", 3000)
	devDoc.Set("database.host", "localhost")
	devDoc.Set("database.ssl", false)
	devDoc.Set("logging.level", "debug")
	devDoc.AppendToArray("features", "debug_toolbar")
	devDoc.AppendToArray("features", "hot_reload")

	fmt.Println(devDoc.String())

	// Production overrides
	fmt.Println("Production environment configuration:")
	prodDoc, _ := yamler.Load(baseConfig) // Fresh copy for production
	prodDoc.Set("server.host", "0.0.0.0")
	prodDoc.Set("server.port", 80)
	prodDoc.Set("database.host", "prod-db.internal")
	prodDoc.Set("database.pool_size", 50)
	prodDoc.Set("database.ssl", true)
	prodDoc.Set("logging.level", "warn")
	prodDoc.AppendToArray("features", "metrics")
	prodDoc.AppendToArray("features", "health_checks")

	fmt.Println(prodDoc.String())

	// Save environment-specific configurations
	devDoc.Save("config-dev.yaml")
	prodDoc.Save("config-prod.yaml")
	fmt.Println("Configurations saved to config-dev.yaml and config-prod.yaml")
	fmt.Println()
}

func infrastructureExample() {
	fmt.Println("3. Infrastructure as Code")
	fmt.Println(strings.Repeat("=", 35))

	// Kubernetes deployment template
	k8sDeployment := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  namespace: default
  labels:
    app: myapp
    version: v1.0.0
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
        version: v1.0.0
    spec:
      containers:
      - name: myapp
        image: myapp:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: NODE_ENV
          value: production
        resources:
          requests:
            memory: 256Mi
            cpu: 100m
          limits:
            memory: 512Mi
            cpu: 200m
`

	doc, _ := yamler.Load(k8sDeployment)

	fmt.Println("Base Kubernetes deployment:")
	replicas, _ := doc.GetInt("spec.replicas")
	image, _ := doc.GetString("spec.template.spec.containers[0].image")
	fmt.Printf("Replicas: %d\n", replicas)
	fmt.Printf("Image: %s\n", image)
	fmt.Println()

	// Scale application based on environment
	environments := map[string]struct {
		replicas int
		image    string
		memory   string
		cpu      string
	}{
		"development": {1, "myapp:dev", "128Mi", "50m"},
		"staging":     {2, "myapp:staging", "256Mi", "100m"},
		"production":  {5, "myapp:v1.0.0", "512Mi", "200m"},
	}

	for env, config := range environments {
		fmt.Printf("Configuring for %s environment:\n", env)

		doc.Set("spec.replicas", config.replicas)
		doc.Set("spec.template.spec.containers[0].image", config.image)
		doc.Set("spec.template.spec.containers[0].resources.requests.memory", config.memory)
		doc.Set("spec.template.spec.containers[0].resources.requests.cpu", config.cpu)
		doc.Set("spec.template.spec.containers[0].resources.limits.memory", config.memory)
		doc.Set("spec.template.spec.containers[0].resources.limits.cpu", config.cpu)

		// Add environment-specific labels
		doc.Set("metadata.labels.environment", env)
		doc.Set("spec.template.metadata.labels.environment", env)

		// Add environment variable
		doc.UpdateArrayElement("spec.template.spec.containers[0].env", 0, map[string]interface{}{
			"name":  "NODE_ENV",
			"value": env,
		})

		filename := fmt.Sprintf("deployment-%s.yaml", env)
		doc.Save(filename)
		fmt.Printf("Saved to %s\n", filename)
	}
	fmt.Println()
}

func configTemplatingExample() {
	fmt.Println("4. Application Configuration Templating")
	fmt.Println(strings.Repeat("=", 45))

	// Configuration template with placeholders
	configTemplate := `# Application Configuration Template
app:
  name: {{APP_NAME}}
  version: {{APP_VERSION}}
  environment: {{ENVIRONMENT}}
  
server:
  host: {{SERVER_HOST}}
  port: {{SERVER_PORT}}
  ssl: {{SSL_ENABLED}}
  
database:
  host: {{DB_HOST}}
  port: {{DB_PORT}}
  name: {{DB_NAME}}
  username: {{DB_USER}}
  password: {{DB_PASSWORD}}
  
external_services:
  redis:
    host: {{REDIS_HOST}}
    port: {{REDIS_PORT}}
  elasticsearch:
    url: {{ELASTICSEARCH_URL}}
    
monitoring:
  enabled: {{MONITORING_ENABLED}}
  endpoint: {{MONITORING_ENDPOINT}}
`

	doc, _ := yamler.Load(configTemplate)

	// Define different deployment scenarios
	scenarios := map[string]map[string]interface{}{
		"local_development": {
			"APP_NAME":            "myapp-dev",
			"APP_VERSION":         "dev",
			"ENVIRONMENT":         "development",
			"SERVER_HOST":         "localhost",
			"SERVER_PORT":         3000,
			"SSL_ENABLED":         false,
			"DB_HOST":             "localhost",
			"DB_PORT":             5432,
			"DB_NAME":             "myapp_dev",
			"DB_USER":             "dev_user",
			"DB_PASSWORD":         "dev_password",
			"REDIS_HOST":          "localhost",
			"REDIS_PORT":          6379,
			"ELASTICSEARCH_URL":   "http://localhost:9200",
			"MONITORING_ENABLED":  false,
			"MONITORING_ENDPOINT": "",
		},
		"production": {
			"APP_NAME":            "myapp",
			"APP_VERSION":         "v1.2.3",
			"ENVIRONMENT":         "production",
			"SERVER_HOST":         "0.0.0.0",
			"SERVER_PORT":         80,
			"SSL_ENABLED":         true,
			"DB_HOST":             "prod-db.internal",
			"DB_PORT":             5432,
			"DB_NAME":             "myapp_prod",
			"DB_USER":             "prod_user",
			"DB_PASSWORD":         "{{SECRET_DB_PASSWORD}}",
			"REDIS_HOST":          "redis.internal",
			"REDIS_PORT":          6379,
			"ELASTICSEARCH_URL":   "https://es.internal:9200",
			"MONITORING_ENABLED":  true,
			"MONITORING_ENDPOINT": "https://monitoring.internal/metrics",
		},
	}

	for scenario, values := range scenarios {
		fmt.Printf("Generating configuration for %s:\n", scenario)

		// Apply template values using wildcard patterns for efficiency
		for placeholder, value := range values {
			pattern := fmt.Sprintf("{{%s}}", placeholder)
			// Find all occurrences and replace
			allValues, _ := doc.GetAll("**")
			for path, currentValue := range allValues {
				if strValue, ok := currentValue.(string); ok {
					if strings.Contains(strValue, pattern) {
						newValue := strings.ReplaceAll(strValue, pattern, fmt.Sprintf("%v", value))
						doc.Set(path, newValue)
					}
				}
			}
		}

		filename := fmt.Sprintf("config-%s.yaml", scenario)
		doc.Save(filename)
		fmt.Printf("Configuration saved to %s\n", filename)

		// Reload template for next scenario
		doc, _ = yamler.Load(configTemplate)
	}
	fmt.Println()
}

func configMigrationExample() {
	fmt.Println("5. Configuration Migration and Transformation")
	fmt.Println(strings.Repeat("=", 50))

	// Old configuration format (v1)
	oldConfig := `# Legacy Configuration Format v1
application:
  app_name: myapp
  app_version: 1.0.0
  
network:
  bind_address: localhost
  bind_port: 8080
  use_ssl: false
  
data_store:
  db_host: localhost
  db_port: 5432
  db_database: myapp
  connection_pool: 10
  
log_config:
  log_level: info
  log_file: /var/log/myapp.log
`

	oldDoc, _ := yamler.Load(oldConfig)

	fmt.Println("Original configuration (v1 format):")
	fmt.Println(oldDoc.String())

	// Migrate to new format (v2)
	fmt.Println("Migrating to v2 format...")

	// Create new document with v2 structure
	newConfig := `# Modern Configuration Format v2
version: 2.0
metadata:
  name: ""
  version: ""
  
server:
  host: ""
  port: 0
  tls:
    enabled: false
    
database:
  connection:
    host: ""
    port: 0
    database: ""
  pool:
    size: 0
    
logging:
  level: ""
  output:
    file: ""
`

	newDoc, _ := yamler.Load(newConfig)

	// Migration mapping
	migrations := map[string]string{
		"application.app_name":       "metadata.name",
		"application.app_version":    "metadata.version",
		"network.bind_address":       "server.host",
		"network.bind_port":          "server.port",
		"network.use_ssl":            "server.tls.enabled",
		"data_store.db_host":         "database.connection.host",
		"data_store.db_port":         "database.connection.port",
		"data_store.db_database":     "database.connection.database",
		"data_store.connection_pool": "database.pool.size",
		"log_config.log_level":       "logging.level",
		"log_config.log_file":        "logging.output.file",
	}

	// Perform migration
	for oldPath, newPath := range migrations {
		value, err := oldDoc.Get(oldPath)
		if err == nil {
			newDoc.Set(newPath, value)
			fmt.Printf("Migrated %s -> %s: %v\n", oldPath, newPath, value)
		}
	}

	fmt.Println("\nMigrated configuration (v2 format):")
	fmt.Println(newDoc.String())

	// Save migrated configuration
	newDoc.Save("config-v2.yaml")
	fmt.Println("Migrated configuration saved to config-v2.yaml")

	// Create migration report
	fmt.Println("\nMigration Report:")
	fmt.Printf("- Configuration format upgraded from v1 to v2\n")
	fmt.Printf("- %d fields successfully migrated\n", len(migrations))
	fmt.Printf("- Structure modernized for better organization\n")
	fmt.Printf("- SSL configuration moved to TLS section\n")
	fmt.Printf("- Database configuration reorganized\n")

	// Clean up generated files
	os.Remove("config-dev.yaml")
	os.Remove("config-prod.yaml")
	os.Remove("deployment-development.yaml")
	os.Remove("deployment-staging.yaml")
	os.Remove("deployment-production.yaml")
	os.Remove("config-local_development.yaml")
	os.Remove("config-production.yaml")
	os.Remove("config-v2.yaml")

	fmt.Println("\nCleanup completed - temporary files removed")
	fmt.Println()
}
