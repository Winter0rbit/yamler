# Yamler Examples

This directory contains comprehensive examples demonstrating various features and use cases of the Yamler library.

## Available Examples

### 1. Basic Usage (`basic_usage/`)
Fundamental operations with YAML documents:
- Loading and parsing YAML
- Reading values with type-safe getters
- Setting and updating values
- Working with arrays and nested structures
- Basic document manipulation

**Run:** `cd basic_usage && go run main.go`

### 2. Comment Alignment (`comment_alignment/`)
Demonstrates the flexible comment alignment features:
- Relative alignment (preserves original spacing)
- Absolute alignment (aligns to specific column)
- Disabled comments (removes inline comments)
- Dynamic mode switching

**Run:** `cd comment_alignment && go run main.go`

### 3. Docker Compose (`docker_compose/`)
Real-world example working with Docker Compose files:
- Loading and modifying Docker Compose configurations
- Adding new services and dependencies
- Working with complex nested structures
- Environment-specific configurations

**Run:** `cd docker_compose && go run main.go`

### 4. Kubernetes (`kubernetes/`)
Kubernetes manifest manipulation:
- Deployment scaling and configuration
- Resource limits and requests
- Environment variables and ConfigMaps
- Volume mounts and metadata

**Run:** `cd kubernetes && go run main.go`

### 5. Ansible (`ansible/`)
Ansible playbook management (array-root documents):
- Working with Ansible playbook structure
- Adding tasks and handlers
- Managing variables and configurations
- Array-root document support

**Run:** `cd ansible && go run main.go`

### 6. Wildcard Patterns (`wildcard_patterns/`)
Advanced pattern matching and bulk operations:
- Finding values with wildcard patterns (`*`, `**`)
- Bulk updates across multiple paths
- Pattern-based configuration management
- Complex multi-environment configurations

**Run:** `cd wildcard_patterns && go run main.go`

### 7. File Operations (`file_operations/`)
File system operations and configuration management:
- Creating and saving new YAML files
- Loading and modifying existing files
- Working with multiple configuration files
- Merging configurations
- Temporary file handling

**Run:** `cd file_operations && go run main.go`

## Key Features Demonstrated

### Format Preservation
All examples maintain original YAML formatting including:
- Indentation styles (2, 4, 6, 8 spaces)
- Comment alignment and positioning
- Flow vs block styles
- Key ordering
- Empty lines and spacing

### Performance Optimizations
Examples showcase Yamler's performance features:
- Formatting information caching
- Efficient path parsing
- Memory-optimized operations
- Bulk operations for large documents

### Real-World Use Cases
Examples cover common DevOps scenarios:
- Configuration management
- Multi-environment deployments
- Infrastructure as Code
- CI/CD pipeline configurations
- Microservices configuration

## Running Examples

Each example is self-contained and can be run independently:

```bash
# Run a specific example
cd examples/basic_usage
go run main.go

# Or run all examples
for dir in examples/*/; do
    if [ -f "$dir/main.go" ]; then
        echo "Running example: $(basename "$dir")"
        (cd "$dir" && go run main.go)
        echo "---"
    fi
done
```

## Example Output

Each example produces detailed output showing:
- Original YAML content
- Step-by-step operations
- Intermediate results
- Final transformed YAML
- Performance metrics (where applicable)

## Integration with Your Project

These examples can serve as templates for integrating Yamler into your own projects:

1. **Copy relevant examples** to your project
2. **Modify the YAML structures** to match your configuration format
3. **Adapt the operations** to your specific use cases
4. **Add error handling** appropriate for your application

## Dependencies

All examples use only the core Yamler library:

```go
import "github.com/Winter0rbit/yamler"
```

No additional dependencies are required.

## Contributing

When adding new examples:

1. Create a new directory under `examples/`
2. Include a `main.go` file with comprehensive comments
3. Use realistic YAML structures relevant to the use case
4. Demonstrate both basic and advanced features
5. Include error handling
6. Update this README with the new example

## Support

For questions about these examples or Yamler usage:
- Check the main documentation in the repository root
- Review the test files for additional usage patterns
- Open an issue for specific questions or bug reports 