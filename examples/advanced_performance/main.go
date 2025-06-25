package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/Winter0rbit/yamler"
)

func main() {
	fmt.Println("=== Yamler Advanced Performance Examples ===\n")

	// Example 1: Caching Performance
	cachingPerformanceDemo()

	// Example 2: Memory Efficiency
	memoryEfficiencyDemo()

	// Example 3: Bulk Operations Performance
	bulkOperationsDemo()

	// Example 4: Large Document Handling
	largeDocumentDemo()

	// Example 5: Repeated Path Operations
	repeatedPathDemo()
}

func cachingPerformanceDemo() {
	fmt.Println("1. Caching Performance Demo")
	fmt.Println(strings.Repeat("=", 40))

	// Create a complex configuration
	yamlContent := generateComplexConfig(50)

	doc, err := yamler.Load(yamlContent)
	if err != nil {
		log.Fatal(err)
	}

	// First run - cache building
	fmt.Println("First run (building caches):")
	start := time.Now()
	for i := 0; i < 50; i++ {
		doc.GetString(fmt.Sprintf("services.service%d.name", i))
		doc.GetInt(fmt.Sprintf("services.service%d.port", i))
		doc.GetBool(fmt.Sprintf("services.service%d.debug", i))
	}
	firstRun := time.Since(start)
	fmt.Printf("Time: %v\n", firstRun)

	// Second run - using cached data
	fmt.Println("\nSecond run (using caches):")
	start = time.Now()
	for i := 0; i < 50; i++ {
		doc.GetString(fmt.Sprintf("services.service%d.name", i))
		doc.GetInt(fmt.Sprintf("services.service%d.port", i))
		doc.GetBool(fmt.Sprintf("services.service%d.debug", i))
	}
	secondRun := time.Since(start)
	fmt.Printf("Time: %v\n", secondRun)

	speedup := float64(firstRun) / float64(secondRun)
	fmt.Printf("Cache speedup: %.1fx faster\n\n", speedup)
}

func memoryEfficiencyDemo() {
	fmt.Println("2. Memory Efficiency Demo")
	fmt.Println(strings.Repeat("=", 35))

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Create and process multiple documents
	docs := make([]*yamler.Document, 10)
	for i := 0; i < 10; i++ {
		yamlContent := generateComplexConfig(20)
		doc, err := yamler.Load(yamlContent)
		if err != nil {
			log.Fatal(err)
		}

		// Perform operations to trigger caching
		doc.SetAll("services.*.debug", false)
		doc.GetAll("services.*.port")

		docs[i] = doc
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	memUsed := m2.Alloc - m1.Alloc
	fmt.Printf("Memory used for 10 documents: %d KB\n", memUsed/1024)
	fmt.Printf("Average per document: %d KB\n", memUsed/(1024*10))
	fmt.Printf("Total allocations: %d\n", m2.TotalAlloc-m1.TotalAlloc)
	fmt.Println()
}

func bulkOperationsDemo() {
	fmt.Println("3. Bulk Operations Performance")
	fmt.Println(strings.Repeat("=", 40))

	yamlContent := generateComplexConfig(100)
	doc, err := yamler.Load(yamlContent)
	if err != nil {
		log.Fatal(err)
	}

	// Individual operations
	fmt.Println("Individual operations:")
	start := time.Now()
	for i := 0; i < 100; i++ {
		doc.Set(fmt.Sprintf("services.service%d.debug", i), false)
		doc.Set(fmt.Sprintf("services.service%d.replicas", i), 5)
	}
	individualTime := time.Since(start)
	fmt.Printf("Time: %v\n", individualTime)

	// Reset values
	doc, _ = yamler.Load(yamlContent)

	// Bulk operations
	fmt.Println("\nBulk operations:")
	start = time.Now()
	doc.SetAll("services.*.debug", false)
	doc.SetAll("services.*.replicas", 5)
	bulkTime := time.Since(start)
	fmt.Printf("Time: %v\n", bulkTime)

	speedup := float64(individualTime) / float64(bulkTime)
	fmt.Printf("Bulk operations speedup: %.1fx faster\n\n", speedup)
}

func largeDocumentDemo() {
	fmt.Println("4. Large Document Handling")
	fmt.Println(strings.Repeat("=", 35))

	// Generate a large YAML document
	yamlContent := generateLargeConfig(500)
	fmt.Printf("Document size: %d KB\n", len(yamlContent)/1024)

	// Load performance
	start := time.Now()
	doc, err := yamler.Load(yamlContent)
	if err != nil {
		log.Fatal(err)
	}
	loadTime := time.Since(start)
	fmt.Printf("Load time: %v\n", loadTime)

	// Query performance
	start = time.Now()
	results, _ := doc.GetAll("**.port")
	queryTime := time.Since(start)
	fmt.Printf("Query all ports (%d results): %v\n", len(results), queryTime)

	// Modification performance
	start = time.Now()
	doc.SetAll("**.debug", false)
	modifyTime := time.Since(start)
	fmt.Printf("Bulk modification time: %v\n", modifyTime)

	// Serialization performance
	start = time.Now()
	result, _ := doc.String()
	serializeTime := time.Since(start)
	fmt.Printf("Serialization time (%d KB): %v\n", len(result)/1024, serializeTime)
	fmt.Println()
}

func repeatedPathDemo() {
	fmt.Println("5. Repeated Path Operations")
	fmt.Println(strings.Repeat("=", 35))

	yamlContent := generateComplexConfig(50)
	doc, err := yamler.Load(yamlContent)
	if err != nil {
		log.Fatal(err)
	}

	// Test path parsing cache
	paths := []string{
		"services.service0.name",
		"services.service0.port",
		"services.service0.debug",
		"services.service1.name",
		"services.service1.port",
		"services.service1.debug",
	}

	// First run - builds path cache
	start := time.Now()
	for i := 0; i < 1000; i++ {
		for _, path := range paths {
			doc.Get(path)
		}
	}
	firstRun := time.Since(start)
	fmt.Printf("First run (6000 operations): %v\n", firstRun)

	// Second run - uses path cache
	start = time.Now()
	for i := 0; i < 1000; i++ {
		for _, path := range paths {
			doc.Get(path)
		}
	}
	secondRun := time.Since(start)
	fmt.Printf("Second run (6000 operations): %v\n", secondRun)

	speedup := float64(firstRun) / float64(secondRun)
	fmt.Printf("Path cache speedup: %.1fx faster\n", speedup)
	fmt.Printf("Average per operation: %v\n", secondRun/6000)
	fmt.Println()
}

func generateComplexConfig(serviceCount int) string {
	var builder strings.Builder

	builder.WriteString("# Complex microservices configuration\n")
	builder.WriteString("version: \"3.8\"\n")
	builder.WriteString("services:\n")

	for i := 0; i < serviceCount; i++ {
		builder.WriteString(fmt.Sprintf(`  service%d:
    name: app%d
    image: myapp:v1.0
    port: %d
    debug: true
    replicas: 3
    resources:
      memory: 512Mi
      cpu: 200m
    environment:
      - NODE_ENV=production
      - PORT=%d
      - DEBUG=true
    health_check:
      endpoint: /health
      interval: 30s
      timeout: 10s
    dependencies:
      - database
      - redis
    tags: [microservice, api, production]
`, i, i, 8000+i, 8000+i))
	}

	builder.WriteString(`
database:
  host: postgres.internal
  port: 5432
  name: myapp_db
  pool_size: 20
  timeout: 30s
  ssl: true

redis:
  host: redis.internal  
  port: 6379
  db: 0
  timeout: 5s
  pool_size: 10

monitoring:
  enabled: true
  prometheus:
    port: 9090
    path: /metrics
  jaeger:
    endpoint: http://jaeger:14268/api/traces
  logs:
    level: info
    format: json
`)

	return builder.String()
}

func generateLargeConfig(serviceCount int) string {
	var builder strings.Builder

	builder.WriteString("# Large configuration with many services\n")
	builder.WriteString("apiVersion: v1\n")
	builder.WriteString("kind: ConfigMap\n")
	builder.WriteString("metadata:\n")
	builder.WriteString("  name: large-config\n")
	builder.WriteString("  namespace: default\n")
	builder.WriteString("data:\n")

	for i := 0; i < serviceCount; i++ {
		builder.WriteString(fmt.Sprintf(`  service%d.yaml: |
    name: service%d
    port: %d
    debug: true
    replicas: 5
    image: myapp:latest
    resources:
      requests:
        memory: 256Mi
        cpu: 100m
      limits:
        memory: 512Mi
        cpu: 200m
    env:
      NODE_ENV: production
      LOG_LEVEL: info
      METRICS_PORT: %d
    volumes:
      - name: data
        mountPath: /data
      - name: logs
        mountPath: /var/log
    networks:
      - frontend
      - backend
    labels:
      app: service%d
      version: v1.0
      tier: backend
`, i, i, 3000+i, 9000+i, i))
	}

	return builder.String()
}
