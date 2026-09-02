# Yamler Features Documentation

**Complete guide to all Yamler features and capabilities.**

## 🎯 Core Features

### 🎨 Format Preservation

Yamler's primary strength is maintaining the original YAML formatting when a document is modified.

**What's Preserved:**
- ✅ **Original indentation** (2, 4, 6, 8 spaces, zero-indent lists)
- ✅ **Comments and positioning** (inline, block, header comments)
- ✅ **Array styles** (flow `[1,2,3]`, block, spaced `[ 1 , 2 , 3 ]`)
- ✅ **Key ordering** (maintains original sequence)
- ✅ **Empty lines and spacing** (blank line patterns)
- ✅ **String styles** (plain, quoted, single-quoted, literal, folded)
- ✅ **Complex flow objects** (multiline `{key: value, nested: {data}}`)
- ✅ **Custom indentations** (non-standard spacing patterns)

**Example:**
```yaml
# Original YAML
app:
  name: myapp         # Application name
  version: "1.0"      # Current version
  servers: [web1, web2]  # Inline array style
  
database:
  host: localhost     # Database host
  port: 5432         # Standard port
```

After modifications with Yamler:
```yaml
# Original YAML
app:
  name: myapp         # Application name  
  version: "2.0"      # Current version (UPDATED!)
  servers: [web1, web2, web3]  # Inline array style (UPDATED!)
  
database:
  host: localhost     # Database host
  port: 3306         # Standard port (UPDATED!)
```

### 🔒 Type-Safe Operations

Comprehensive type-safe API for all YAML data types with automatic conversion.

#### Basic Type Operations
```go
// String operations
name, err := doc.GetString("app.name")
doc.SetString("app.name", "newapp")

// Numeric operations
port, err := doc.GetInt("database.port")        // Returns int64
timeout, err := doc.GetFloat("app.timeout")     // Returns float64
doc.SetInt("database.port", 5432)
doc.SetFloat("app.timeout", 30.5)

// Boolean operations (flexible parsing)
debug, err := doc.GetBool("app.debug")
doc.SetBool("app.debug", true)
```

#### Array Operations
```go
// Array access and manipulation
servers, err := doc.GetStringSlice("app.servers")
ports, err := doc.GetIntSlice("app.ports")
doc.SetStringSlice("app.servers", []string{"web1", "web2"})

// Individual array elements
server, err := doc.GetArrayElement("servers", 0)
port, err := doc.GetTypedArrayElement("ports", 0, "int") // "string", "int", "float", "bool"
doc.Set("ports[0]", 8080)
```

#### Map Operations
```go
// Map access
config, err := doc.GetMap("database")
doc.Set("database", map[string]interface{}{
    "host": "localhost",
    "port": 5432,
})
```

#### Flexible Boolean Parsing
Yamler supports multiple boolean formats:
- `true/false` (standard)
- `yes/no` (YAML style)
- `1/0` (numeric)
- `on/off` (configuration style)

```go
// All of these are read as booleans by GetBool
// ssl: yes      → true
// debug: on     → true
// cache: 1      → true
// logging: off  → false
enabled, err := doc.GetBool("ssl")
```

### 🛠️ Advanced Array Operations

Full CRUD operations on arrays with perfect style preservation.

#### Array Information
```go
// Get array length
length, err := doc.GetArrayLength("servers")

// Check if array exists
exists := length > 0
```

#### Array Modifications
```go
// Append elements
doc.AppendToArray("servers", "web3")
doc.AppendToArray("ports", 9090)

// Insert at specific position
doc.InsertIntoArray("servers", 1, "web1.5")

// Update existing elements
doc.UpdateArrayElement("servers", 0, "web1-updated")

// Remove elements
doc.RemoveFromArray("servers", 2)
```

#### Array Style Preservation
Yamler maintains different array styles:

**Flow Arrays:**
```yaml
# Compact: [1,2,3]
# Spaced: [ 1 , 2 , 3 ]
# Multiline flow:
items: [
  item1,
  item2,
  item3
]
```

**Block Arrays:**
```yaml
items:
  - item1
  - item2
  - item3
```

### 🎯 Wildcard Pattern Operations

Powerful pattern matching for bulk operations across YAML documents.

#### Single-Level Wildcards (`*`)
```go
// Get all immediate children
appPorts := doc.GetAll("apps.*.port")       // All app ports
serviceNames := doc.GetAll("services.*.name") // All service names

// Bulk updates
doc.SetAll("apps.*.debug", false)           // Disable debug for all apps
doc.SetAll("services.*.replicas", 3)        // Scale all services
```

#### Recursive Wildcards (`**`)
```go
// Get all matching values anywhere in document
allDebugFlags := doc.GetAll("**.debug")     // All debug flags
allTimeouts := doc.GetAll("**.timeout")     // All timeout values

// Deep bulk updates
doc.SetAll("**.debug", false)               // Disable all debug flags
doc.SetAll("**.ssl", true)                  // Enable SSL everywhere
```

#### Pattern Examples
```go
// Complex patterns
doc.GetAll("services.*.containers[0].port") // First container port for all services
doc.GetAll("**.env[*].name")                // All environment variable names
doc.GetAll("apps.web.*.config")             // All config under apps.web

// Get matching keys (not values)
keys, _ := doc.GetKeys("apps.*")            // ["apps.web", "apps.api"]
paths, _ := doc.GetPathsRecursive()         // Every leaf path in the document

// SetAll only updates paths that already exist; it never creates keys.
```

### 🧩 Document Merging

Intelligent document merging with format and comment preservation.

#### Basic Merging
```go
// Load documents
doc1, _ := yamler.LoadFile("base.yaml")
doc2, _ := yamler.LoadFile("override.yaml")

// Merge with format preservation
err := doc1.Merge(doc2)
```

#### Targeted Merging
```go
// Merge at specific path
err := doc1.MergeAt("database", doc2)

// Merge configuration sections
err := doc1.MergeAt("app.settings", settingsDoc)
```

#### Merge Behavior
- **Values**: Override from source document
- **Arrays**: Complete replacement (not append)
- **Comments**: Target comments are kept for keys that already existed; comments of new keys come from the source
- **Formatting**: Maintained from base document
- **New keys**: Appended to the end of the mapping

### 📑 Multi-Document Streams

`Load` accepts exactly one document and returns an error for a `---`-separated stream. `LoadAll` returns one `Document` per document; each keeps its own formatting and separator:

```go
docs, _ := yamler.LoadAllFile("manifests.yaml")
docs[1].Set("spec.replicas", 5)
yamler.SaveAll("manifests.yaml", docs)
```

### ✅ Schema Validation

Validation rules are loaded from YAML or JSON with `LoadSchemaFromString` / `LoadSchemaFromFile`. JSON Schema type names (`object`, `integer`, `number`, `boolean`) are accepted alongside `map`, `int`, `float`, `bool`, `string`, `array`, `any`:

```go
rule, _ := yamler.LoadSchemaFromString(`{"type": "object", "required": ["name"], "properties": {"name": {"type": "string", "minLength": 1}}}`)
if err := doc.Validate(rule); err != nil {
    log.Println(err) // path : required field name is missing
}
```

### 💬 Comment Alignment System

Flexible comment positioning and formatting control.

#### Alignment Modes

**1. Relative Alignment (Default)**
Preserves original spacing between value and comment:
```yaml
name: myapp    # App name
port: 8080        # Port number
debug: true # Debug flag
```

**2. Absolute Alignment**
Aligns all comments to specific column:
```go
doc.SetAbsoluteCommentAlignment(25)
```
```yaml
name: myapp              # App name
port: 8080               # Port number
debug: true              # Debug flag
```

**3. Disabled Comments**
Removes all inline comments:
```go
doc.DisableCommentAlignment()
```
```yaml
name: myapp
port: 8080
debug: true
```

#### Comment Control API
```go
// Set alignment mode
doc.SetCommentAlignment(yamler.CommentAlignmentRelative)
doc.SetCommentAlignment(yamler.CommentAlignmentAbsolute)
doc.SetCommentAlignment(yamler.CommentAlignmentDisabled)

// Convenience methods
doc.EnableRelativeCommentAlignment()
doc.SetAbsoluteCommentAlignment(30)
doc.DisableCommentAlignment()
```

### 🌊 Complex Flow Object Support

Perfect handling of complex nested flow structures.

#### Multiline Flow Objects
```yaml
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
```

#### Multiline Flow Arrays
```yaml
matrix: [
  [1, 2, 3],
  [4, 5, 6],
  [7, 8, 9]
]
```

#### Mixed Styles
```yaml
config:
  inline_array: [1, 2, 3]
  block_array:
    - item1
    - item2
  inline_object: {key: value, number: 42}
  block_object:
    key1: value1
    key2: value2
```

All these structures are perfectly preserved during modifications.

## ⚡ Performance Features

### Advanced Caching System

Yamler implements multiple caching layers for optimal performance.

#### Formatting Information Cache
- **What**: The formatting snapshot (indentation, styles, comment spacing, blank lines) is detected once when a document is loaded and reused by every serialization.

#### Path Parsing Cache
- **What**: Parsed path expressions are cached in a bounded map (10 000 entries), so repeated access to the same paths does not split strings again.

#### Memory Optimization
- **Buffer Pooling**: Byte buffers used for serialization are pooled with `sync.Pool`.

### Benchmarks

`benchmark_test.go` contains benchmarks for loading, serialization, getters, setters, wildcard and array operations:

```bash
go test -bench . -benchmem
```

### Bulk Operations

Every mutation (`Set`, `AppendToArray`, ...) re-serializes the document to keep the formatting snapshot current. `SetAll` applies all matching updates and serializes once, so prefer it over a loop of `Set` calls when many paths change:
```go
// Individual operations: one serialization per call
for i := 0; i < 100; i++ {
    doc.Set(fmt.Sprintf("services.service%d.debug", i), false)
}

// Bulk operation: one serialization
doc.SetAll("services.*.debug", false)
```

## 📊 Array Document Support

Special support for Ansible-style array-root documents.

#### Array Root Documents
```yaml
# Ansible playbook (array at root)
- name: Install packages
  apt:
    name: "{{ item }}"
  loop:
    - nginx
    - postgresql
    
- name: Start services
  service:
    name: "{{ item }}"
    state: started
  loop:
    - nginx
    - postgresql
```

#### Array Root Operations
```go
// Load array-root document
doc, err := yamler.LoadFile("playbook.yaml")

// Access array elements
taskName, err := doc.GetString("[0].name")            // Task name
loop, err := doc.GetArrayDocumentElement(0, "loop")   // Same, by element index
all, err := doc.GetSlice("")                          // The whole list

// Modify array elements
doc.SetArrayElement(1, "service.state", "restarted")  // element index + path
doc.Set("apt.name", "nginx")                          // shorthand for element 0
doc.AddArrayElement(newTask)                          // append a task
```

## 🔧 Error Handling

Comprehensive error handling with detailed error messages.

### Error Types
```go
// File errors
doc, err := yamler.LoadFile("nonexistent.yaml")
if err != nil {
    // Handle file not found, permission errors, etc.
}

// Path errors
value, err := doc.Get("invalid.path[abc]")
if err != nil {
    // Handle invalid path syntax
}

// Type conversion errors
port, err := doc.GetInt("app.name") // name is string, not int
if err != nil {
    // Handle type mismatch
}

// Missing key errors
value, err := doc.Get("nonexistent.key")
if err != nil {
    // Handle missing keys
}
```

### Error Context
Errors are prefixed with the path that caused them, e.g. `path app.port: invalid integer value: ...` or `path items[abc]: invalid array index: abc`. Errors from `os` and `yaml.v3` are wrapped and can be inspected with `errors.Is` / `errors.As`.

## 🎨 Real-World Compatibility

### Supported Formats

**Tested with:**
- ✅ **Application configurations** (JSON-like, nested objects)
- ✅ **Docker Compose** files (services, networks, volumes)
- ✅ **Kubernetes** manifests (deployments, services, configmaps)
- ✅ **Ansible** playbooks (array-root, tasks, variables)
- ✅ **GitHub Actions** workflows (jobs, steps, matrix)
- ✅ **CI/CD configurations** (various pipeline formats)

- ✅ **Zero-indent arrays**: Kubernetes style `containers:\n- item`
- ✅ **Multi-document streams**: `LoadAll` / `SaveAll`
- ✅ **Anchors and merge keys**: `<<: *defaults`

### Format Support Matrix

| Feature | Support Level | Notes |
|---------|---------------|-------|
| **2-space indentation** | ✅ Perfect | Standard YAML |
| **4-space indentation** | ✅ Perfect | Common in configs |
| **6-space indentation** | ✅ Perfect | Custom spacing |
| **8-space indentation** | ✅ Perfect | Large team preference |
| **Tab indentation** | ❌ Rejected | Forbidden by the YAML specification |
| **Mixed indentation** | ⚠️ Partial | Keys and items keep their original column; new keys follow their parent |
| **Flow arrays** | ✅ Perfect | `[1, 2, 3]` style |
| **Spaced flow arrays** | ✅ Perfect | `[ 1 , 2 , 3 ]` style |
| **Block arrays** | ✅ Perfect | Multi-line arrays |
| **Multiline flow** | ✅ Perfect | Complex nested structures |
| **Comments** | ✅ Preserved | Inline spacing and standalone comment indentation kept |
| **Empty lines** | ✅ Perfect | Keyed by path |
| **Zero-indent lists** | ✅ Perfect | `containers:\n- name: web` |
| **Anchors / merge keys** | ✅ Perfect | `&a`, `*a`, `<<: *a` |
| **Multi-document streams** | ✅ Perfect | via `LoadAll` / `SaveAll` |
| **String styles** | ✅ Perfect | Plain, quoted, literal, folded |

## 🚀 Advanced Use Cases

### Configuration Management
- **Multi-environment** configurations
- **Template substitution** and processing
- **Configuration validation** and transformation
- **Secret management** integration

### DevOps Automation
- **CI/CD pipeline** generation and modification
- **Infrastructure as Code** template processing
- **Container orchestration** configuration management
- **Deployment automation** scripting

### Development Workflows
- **Configuration file** maintenance and updates
- **Build system** configuration management
- **Development environment** setup automation
- **Code generation** from YAML templates

### Enterprise Integration
- **Legacy system** configuration migration
- **Multi-team** configuration standardization
- **Compliance** and governance automation
- **Audit trail** maintenance for configuration changes

## 📈 Performance Characteristics

### Scalability
- **Small files** (< 1KB): Near-instant processing
- **Medium files** (1-100KB): Sub-second processing
- **Large files** (100KB-1MB): 1-5 second processing
- **Very large files** (> 1MB): Linear scaling with size

### Memory Usage
- **Base overhead**: ~2-3MB for library
- **Per document**: ~1-2KB + document size
- **Caching overhead**: ~10-20% of document size
- **Bulk operations**: Constant memory usage

### CPU Usage
- **Parsing**: Single-threaded, optimized
- **Modifications**: In-memory operations, very fast
- **Serialization**: Optimized with caching
- **Wildcards**: Efficient tree traversal

## 🔮 Future Roadmap

### Planned Features
- **Configuration diffing** tools
- **Template engine** integration
- **Plugin system** for custom operations

### Performance Improvements
- **Parallel processing** for large documents
- **Streaming operations** for very large files
- **Memory mapping** for read-only operations
- **Incremental parsing** for partial updates

### Format Support
- **Per-item layout records** for lists whose items are laid out differently
- **Custom comment styles** (configurable formatting)
- **YAML 1.2 features** (additional specification support)

---

**This documentation covers all current Yamler features. For the latest updates, check the main README and examples.** 