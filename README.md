# Yamler

**A powerful Go YAML library that preserves formatting, comments, and structure.**

## Features

Yamler is a powerful Go library for working with YAML files while **preserving original formatting, comments, and structure**. Unlike most standard YAML libraries that lose formatting during parsing, Yamler maintains every aspect of your YAML files during read and write operations.

## 🎯 Why Yamler?

**The Problem:** Standard YAML libraries destroy your carefully crafted file structure.

**Before (with standard libraries):**
```yaml
# My important config
app:
  name: myapp     # Application name
  debug: true
  items: [1, 2, 3]    # Inline array
  servers:
    - web1      # Production servers
    - web2
```

**After modification (with standard libraries):**
```yaml
app:
  debug: true
  items:
  - 1
  - 2
  - 3
  name: myapp
  servers:
  - web1
  - web2
```

**With Yamler:** Your formatting, comments, and structure remain **exactly** as you wrote them! 🎉

## ✨ Features

- 🎨 **Format Preservation** - Maintains original YAML formatting and comments
- 🔒 **Type-Safe Operations** - Strongly typed getters and setters
- 🧩 **Document Merging** - Merge YAML documents while preserving structure
- 🎯 **Wildcard Patterns** - Bulk operations with `*.field` and `**.recursive` patterns  
- 🛠️ **Array Operations** - Full CRUD operations on arrays with style preservation
- 🎭 **Flexible Boolean Parsing** - Supports `true/false`, `yes/no`, `1/0`, `on/off`
- ✅ **Schema Validation** - Built-in JSON Schema compatibility
- 🚀 **Production Ready** - Comprehensive error handling and testing

## 📦 Installation

```bash
go get github.com/Winter0rbit/yamler
```

## 🚀 Quick Start

```go
package main

import (
    "fmt"
    "log"
    "github.com/Winter0rbit/yamler"
)

func main() {
    // Load YAML file (preserves all formatting)
    doc, err := yamler.LoadFile("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Get values with type safety
    appName, _ := doc.GetString("app.name")
    debug, _ := doc.GetBool("app.debug") 
    servers, _ := doc.GetStringSlice("app.servers")
    
    fmt.Printf("App: %s, Debug: %v, Servers: %v\n", appName, debug, servers)

    // Modify values (formatting preserved!)
    doc.Set("app.version", "2.0")
    doc.AppendToArray("app.servers", "web3")

    // Save back with original formatting intact
    doc.SaveFile("config.yaml")
}
```

## 📚 Comprehensive Examples

### 1. Format Preservation in Action

**Original file:**
```yaml
# Application Configuration
app:
  name: myapp           # App identifier
  version: "1.0"
  debug: yes            # Enable debug mode
  
  # Server configuration  
  servers: [web1, web2]  # Inline style
  
  database:
    host: localhost      # DB connection
    port: 5432
    pools:               # Connection pools
      - name: primary
        size: 10
      - name: replica  
        size: 5
```

**After modifications with Yamler:**
```go
doc, _ := yamler.LoadFile("app.yaml")

// Update values
doc.Set("app.version", "2.0")
doc.Set("app.debug", false)
doc.Set("database.port", 3306)
doc.AppendToArray("app.servers", "web3")
doc.UpdateArrayElement("database.pools", 0, map[string]interface{}{
    "name": "primary",
    "size": 20,
})

doc.SaveFile("app.yaml")
```

**Result (formatting preserved!):**
```yaml
# Application Configuration
app:
  name: myapp           # App identifier
  version: "2.0"
  debug: false          # Enable debug mode
  
  # Server configuration  
  servers: [web1, web2, web3]  # Inline style preserved!
  
  database:
    host: localhost      # DB connection  
    port: 3306
    pools:               # Connection pools
      - name: primary
        size: 20
      - name: replica
        size: 5
```

### 2. Type-Safe Operations

```go
// Strong typing prevents runtime errors
name, err := doc.GetString("app.name")        // Returns string
port, err := doc.GetInt("database.port")      // Returns int64
debug, err := doc.GetBool("app.debug")        // Returns bool
tags, err := doc.GetStringSlice("app.tags")   // Returns []string
config, err := doc.GetMap("database")         // Returns map[string]interface{}

// Flexible boolean parsing
doc.Set("feature.enabled", "yes")    // Parsed as true
doc.Set("feature.ssl", "on")         // Parsed as true  
doc.Set("feature.cache", 1)          // Parsed as true
doc.Set("feature.debug", "false")    // Parsed as false

enabled, _ := doc.GetBool("feature.enabled")  // true
ssl, _ := doc.GetBool("feature.ssl")          // true
```

### 3. Array Operations with Style Preservation

```go
// Flow-style arrays stay flow-style
doc := yamler.Load("items: [1, 2, 3]")
doc.AppendToArray("items", 4)
// Result: "items: [1, 2, 3, 4]"

// Block-style arrays stay block-style  
doc = yamler.Load(`
items:
  - apple
  - banana`)
doc.AppendToArray("items", "cherry")
// Result:
// items:
//   - apple  
//   - banana
//   - cherry

// Full array CRUD
doc.RemoveFromArray("items", 1)           // Remove index 1
doc.UpdateArrayElement("items", 0, "orange") // Update index 0
doc.InsertIntoArray("items", 1, "grape")  // Insert at index 1
length, _ := doc.GetArrayLength("items")  // Get array size
```

### 4. Document Merging

```go
// Base configuration
base := yamler.Load(`
app:
  name: myapp
  version: 1.0
  settings:
    debug: true
    timeout: 30`)

// Override configuration  
override := yamler.Load(`
app:
  version: 2.0
  settings:
    ssl: true
    timeout: 60
author: developer`)

// Merge with format preservation
base.Merge(override)

// Result:
// app:
//   name: myapp      # Preserved from base
//   version: 2.0     # Updated from override  
//   settings:
//     debug: true    # Preserved from base
//     timeout: 60    # Updated from override
//     ssl: true      # Added from override
// author: developer  # Added from override

// Targeted merging
base.MergeAt("app.database", dbConfig)  // Merge only into specific path
```

### 5. Wildcard Pattern Operations

```go
config := yamler.Load(`
environments:
  development:
    debug: true
    timeout: 30
    database:
      host: dev-db
  production:  
    debug: false
    timeout: 60
    database:
      host: prod-db
  staging:
    debug: true  
    timeout: 45
    database:
      host: stage-db`)

// Get all debug settings
allDebug, _ := config.GetAll("environments.*.debug")
// Returns: {
//   "environments.development.debug": true,
//   "environments.production.debug": false, 
//   "environments.staging.debug": true
// }

// Recursive search  
allHosts, _ := config.GetAll("**.host")
// Returns: {
//   "environments.development.database.host": "dev-db",
//   "environments.production.database.host": "prod-db",
//   "environments.staging.database.host": "stage-db"  
// }

// Bulk operations
config.SetAll("environments.*.timeout", 90)  // Set all timeouts
keys, _ := config.GetKeys("environments.*")   // Get all environment names
```

### 6. Schema Validation

```go
// Define schema
schema := yamler.LoadSchema(`
type: object
properties:
  app:
    type: object
    properties:
      name:
        type: string
        minLength: 1
      version:
        type: string
        pattern: "^\\d+\\.\\d+$"
      debug:
        type: boolean
    required: [name, version]
required: [app]`)

// Validate document
doc, _ := yamler.LoadFile("config.yaml")
err := doc.Validate(schema)
if err != nil {
    fmt.Printf("Validation failed: %v\n", err)
}
```

## 🎯 Advanced Use Cases

### Configuration Management

```go
// Load environment-specific configs
baseConfig, _ := yamler.LoadFile("base.yaml")
envConfig, _ := yamler.LoadFile(fmt.Sprintf("%s.yaml", env))

// Merge environment overrides
baseConfig.Merge(envConfig)

// Apply runtime overrides
if port := os.Getenv("PORT"); port != "" {
    baseConfig.Set("server.port", port)
}

// Enable all debug flags in development
if env == "development" {
    baseConfig.SetAll("**.debug", true)
}
```

### Template Processing

```go
// Load template
template, _ := yamler.LoadFile("k8s-template.yaml") 

// Substitute values
template.Set("metadata.name", appName)
template.Set("spec.replicas", replicas)
template.SetAll("spec.template.spec.containers.*.image", imageTag)

// Generate final config
template.SaveFile(fmt.Sprintf("deploy-%s.yaml", env))
```

### Migration Scripts

```go
// Load old config format
oldConfig, _ := yamler.LoadFile("old-config.yaml")

// Extract and transform data
dbHost, _ := oldConfig.GetString("database.host")
dbPort, _ := oldConfig.GetInt("database.port")

// Create new structure while preserving comments
newConfig, _ := yamler.LoadFile("new-config-template.yaml")
newConfig.Set("database.connection.host", dbHost)
newConfig.Set("database.connection.port", dbPort)

// Migrate arrays with structure preservation
servers, _ := oldConfig.GetStringSlice("servers")
for _, server := range servers {
    newConfig.AppendToArray("infrastructure.servers", map[string]interface{}{
        "name": server,
        "role": "web",
    })
}
```

## 📊 Performance & Comparison

| Feature | Yamler | go-yaml/yaml | goccy/go-yaml |
|---------|--------|--------------|---------------|
| Format Preservation | ✅ **Yes** | ❌ | ❌ |
| Comment Preservation | ✅ **Yes** | ❌ | ❌ |
| Type-Safe API | ✅ | ❌ | ❌ |
| Array Operations | ✅ | ❌ | ❌ |
| Document Merging | ✅ | ❌ | ❌ |
| Wildcard Patterns | ✅ | ❌ | ❌ |
| Schema Validation | ✅ | ❌ | ❌ |
| Production Ready | ✅ | ✅ | ✅ |

## 🛠️ API Reference

### Document Operations
```go
// Loading
doc, err := yamler.LoadFile("config.yaml")
doc, err := yamler.LoadBytes(data)
doc, err := yamler.Load(yamlString)

// Saving  
err := doc.SaveFile("config.yaml")
bytes, err := doc.ToBytes()
str, err := doc.String()
```

### Type-Safe Getters
```go
// Scalar values
value, err := doc.Get(path)                    // interface{}
str, err := doc.GetString(path)                // string
num, err := doc.GetInt(path)                   // int64
flt, err := doc.GetFloat(path)                 // float64
flag, err := doc.GetBool(path)                 // bool

// Collections
slice, err := doc.GetSlice(path)               // []interface{}
strs, err := doc.GetStringSlice(path)          // []string
m, err := doc.GetMap(path)                     // map[string]interface{}
```

### Type-Safe Setters
```go
// Universal setter
err := doc.Set(path, value)                    // any type

// Typed setters  
err := doc.SetString(path, "value")            // string
err := doc.SetInt(path, 42)                    // int64
err := doc.SetFloat(path, 3.14)                // float64
err := doc.SetBool(path, true)                 // bool
err := doc.SetStringSlice(path, []string{...}) // []string
err := doc.SetIntSlice(path, []int64{...})     // []int64
err := doc.SetFloatSlice(path, []float64{...}) // []float64
err := doc.SetBoolSlice(path, []bool{...})     // []bool
err := doc.SetMapSlice(path, []map[string]interface{}{...}) // []map[string]interface{}
```

### Array Operations
```go
// CRUD operations
err := doc.AppendToArray(path, value)          // Add to end
err := doc.RemoveFromArray(path, index)        // Remove by index
err := doc.UpdateArrayElement(path, index, value) // Update by index
err := doc.InsertIntoArray(path, index, value) // Insert at index

// Array info
length, err := doc.GetArrayLength(path)        // Get size
element, err := doc.GetArrayElement(path, index) // Get by index
```

### Wildcard Operations
```go
// Pattern matching
results, err := doc.GetAll(pattern)            // Get all matches
keys, err := doc.GetKeys(pattern)              // Get matching keys  
err := doc.SetAll(pattern, value)              // Set all matches

// Utility
paths, err := doc.GetPathsRecursive()          // All paths in document
filtered := yamler.FilterByPattern(data, pattern) // Filter map by pattern
```

### Document Merging
```go
// Merge operations
err := doc.Merge(other)                        // Full merge
err := doc.MergeAt(path, other)                // Merge at specific path
```

### Validation
```go
// Schema validation
schema, err := yamler.LoadSchema(schemaYAML)   // Load schema
err := doc.Validate(schema)                    // Validate document
```

## 🎨 Supported Wildcard Patterns

| Pattern | Description | Example |
|---------|-------------|---------|
| `config.*` | Any key at level | `config.debug`, `config.timeout` |
| `config.**` | Any nested key (recursive) | `config.db.host`, `config.cache.redis.port` |  
| `**.debug` | Any `debug` key anywhere | `app.debug`, `services.api.debug` |
| `servers[*]` | Any array element | `servers[0]`, `servers[1]` |

## 🚨 Error Handling

Yamler provides detailed error information:

```go
doc, err := yamler.LoadFile("config.yaml")
if err != nil {
    switch {
    case errors.Is(err, yamler.ErrFileNotFound):
        // Handle missing file
    case errors.Is(err, yamler.ErrInvalidYAML):  
        // Handle syntax errors
    default:
        // Handle other errors
    }
}

// Path-specific errors
value, err := doc.GetString("nonexistent.path")
if errors.Is(err, yamler.ErrPathNotFound) {
    // Handle missing path
}
```

## 🧪 Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks  
go test -bench=. ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built on top of [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)
- Inspired by the need for configuration management that preserves human-readable formatting
- Thanks to all contributors and users of the library

---

**⭐ If Yamler helps you, please give it a star on GitHub! ⭐** 