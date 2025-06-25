# Yamler Examples

This directory contains comprehensive examples demonstrating all features of the Yamler library.

## 📂 Available Examples

### 🚀 [Basic Usage](basic_usage/)
**Fundamental operations and type-safe getters**
- Format preservation magic
- Type-safe operations (string, int, float, bool, arrays, maps)
- Comment alignment features (relative, absolute, disabled)
- Performance features demo
- Complex flow object handling

```bash
cd basic_usage && go run main.go
```

### 💬 [Comment Alignment](comment_alignment/)
**Flexible comment formatting control**
- Relative alignment (preserves original spacing)
- Absolute alignment (align to specific column)
- Comment removal
- Real-world configuration examples

```bash
cd comment_alignment && go run main.go
```

### 🐳 [Docker Compose](docker_compose/)
**Real-world container orchestration**
- Multi-service Docker Compose manipulation
- Service scaling and configuration updates
- Environment variable management
- Volume and network configuration

```bash
cd docker_compose && go run main.go
```

### ☸️ [Kubernetes](kubernetes/)
**Manifest manipulation and scaling**
- Deployment scaling and updates
- ConfigMap and Secret management
- Multi-environment configuration
- Resource limit adjustments

```bash
cd kubernetes && go run main.go
```

### 📋 [Ansible](ansible/)
**Playbook management (array-root documents)**
- Ansible playbook manipulation
- Task management and updates
- Variable and template handling
- Role-based configuration

```bash
cd ansible && go run main.go
```

### 🎯 [Wildcard Patterns](wildcard_patterns/)
**Bulk operations and pattern matching**
- Single-level wildcards (`*.field`)
- Recursive wildcards (`**.field`)
- Bulk value updates
- Pattern-based queries

```bash
cd wildcard_patterns && go run main.go
```

### 📁 [File Operations](file_operations/)
**File system integration and merging**
- File loading and saving
- Document merging with format preservation
- Batch file processing
- Configuration templating

```bash
cd file_operations && go run main.go
```

### ⚡ [Advanced Performance](advanced_performance/)
**Performance optimization features**
- Caching performance demonstration
- Memory efficiency analysis
- Bulk operations vs individual operations
- Large document handling
- Repeated path operations optimization

```bash
cd advanced_performance && go run main.go
```

### 🌍 [Real-World Use Cases](real_world_use_cases/)
**Production-ready scenarios**
- CI/CD pipeline configuration (GitHub Actions)
- Multi-environment configuration management
- Infrastructure as Code (Kubernetes deployments)
- Configuration templating and substitution
- Configuration migration and transformation

```bash
cd real_world_use_cases && go run main.go
```

## 🏃‍♂️ Running All Examples

Run all examples at once:

```bash
./run_all.sh
```

Or run individual examples:

```bash
cd <example_directory>
go run main.go
```

## 📊 Example Categories

### 🎯 **Beginner Examples**
- **Basic Usage** - Start here for fundamental concepts
- **Comment Alignment** - Learn comment formatting control
- **File Operations** - Basic file manipulation

### 🚀 **Intermediate Examples**  
- **Docker Compose** - Real container orchestration
- **Kubernetes** - Kubernetes manifest handling
- **Wildcard Patterns** - Advanced pattern matching

### 🔥 **Advanced Examples**
- **Ansible** - Complex array-root document handling
- **Advanced Performance** - Performance optimization
- **Real-World Use Cases** - Production scenarios

## 🎨 Key Features Demonstrated

### ✨ **Format Preservation**
All examples demonstrate Yamler's core strength - perfect preservation of:
- Original indentation and spacing
- Comments and their positioning
- Array styles (flow vs block)
- Complex nested structures

### 🔧 **Type-Safe Operations**
Examples show comprehensive type-safe operations:
- `GetString()`, `GetInt()`, `GetFloat()`, `GetBool()`
- `GetStringSlice()`, `GetIntSlice()`, `GetMap()`
- `SetString()`, `SetInt()`, `SetBool()`, etc.
- Array operations with type safety

### 🎯 **Advanced Features**
- **Wildcard Patterns**: `*.field` and `**.recursive` matching
- **Comment Alignment**: Flexible positioning control
- **Document Merging**: Structure-preserving merges
- **Performance Optimization**: Caching and bulk operations
- **Array Operations**: CRUD with style preservation

### 🌊 **Complex Structures**
Perfect handling of:
- Multiline flow objects `{key: value, nested: {data: here}}`
- Flow arrays `[1, 2, 3]` and `[ 1 , 2 , 3 ]`
- Mixed flow/block styles in same document
- Custom indentation (2, 4, 6, 8 spaces)

## 📈 **Performance Highlights**

Examples demonstrate real performance improvements:
- **21% faster** ToBytes operations
- **48% faster** formatting detection  
- **79% faster** path parsing with caching
- **14-25% improvement** in real-world scenarios

## 🎯 **Real-World Compatibility: 100%**

All examples work with production configurations:
- ✅ **Docker Compose** files
- ✅ **Kubernetes** manifests  
- ✅ **Ansible** playbooks
- ✅ **GitHub Actions** workflows
- ✅ **Application** configurations

## 🛠️ **Development Setup**

Each example is self-contained with its own `go.mod`:

```bash
# Run any example
cd <example_name>
go mod tidy  # Download dependencies
go run main.go
```

## 📚 **Learning Path**

**Recommended order for learning:**

1. **Basic Usage** - Core concepts and operations
2. **Comment Alignment** - Formatting control
3. **File Operations** - File handling basics
4. **Docker Compose** - Real-world container config
5. **Wildcard Patterns** - Advanced pattern matching
6. **Kubernetes** - Complex manifest handling
7. **Advanced Performance** - Optimization techniques
8. **Real-World Use Cases** - Production scenarios
9. **Ansible** - Array-root document mastery

## 🤝 **Contributing Examples**

Want to add more examples? Great! Please:

1. Create a new directory under `examples/`
2. Add a complete, runnable `main.go`
3. Include a `go.mod` with proper module setup
4. Add comprehensive comments explaining the concepts
5. Update this README with your example
6. Test thoroughly and ensure it demonstrates unique features

## 📄 **Example Template**

```go
package main

import (
    "fmt"
    "log"
    "github.com/Winter0rbit/yamler"
)

func main() {
    fmt.Println("=== Your Example Title ===\n")
    
    // Your example code here
    yamlContent := `your: yaml`
    
    doc, err := yamler.Load(yamlContent)
    if err != nil {
        log.Fatal(err)
    }
    
    // Demonstrate specific features
    fmt.Println("Original:")
    fmt.Println(doc.String())
    
    // Make modifications
    doc.Set("your.field", "new_value")
    
    fmt.Println("Modified:")
    fmt.Println(doc.String())
}
```

---

**Happy coding with Yamler! 🎉** 