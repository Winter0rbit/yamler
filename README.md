# Yamler

**A Go YAML library that edits documents while preserving their formatting, comments and structure.**

[![Go Reference](https://pkg.go.dev/badge/github.com/Winter0rbit/yamler.svg)](https://pkg.go.dev/github.com/Winter0rbit/yamler)
[![Go Report Card](https://goreportcard.com/badge/github.com/Winter0rbit/yamler)](https://goreportcard.com/report/github.com/Winter0rbit/yamler)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Tests](https://github.com/Winter0rbit/yamler/actions/workflows/test.yml/badge.svg)](https://github.com/Winter0rbit/yamler/actions/workflows/test.yml)

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

**With Yamler** the file keeps its comments, key order, indentation and flow/block styles. Yamler parses the document with [`gopkg.in/yaml.v3`](https://gopkg.in/yaml.v3), remembers how the original text was laid out, and restores that layout when the modified document is written back.

## ✨ Key Features

- 🎨 **Format preservation** - comments, key order, indentation (2/4/6/8 spaces, zero-indent lists), blank lines, flow/block styles, literal and folded scalars, document markers
- 🔒 **Type-safe operations** - typed getters and setters with automatic conversion
- 🧩 **Document merging** - merge documents while keeping the target's formatting
- 🎯 **Wildcard patterns** - bulk reads and updates with `*`, `**` and `[*]`
- 🛠️ **Array operations** - append, insert, update and remove with style preservation
- 💬 **Comment alignment** - relative, absolute (column) or disabled
- 🎭 **Flexible boolean parsing** - `true/false`, `yes/no`, `on/off`, `1/0`
- ✅ **Schema validation** - JSON-Schema-like rules (`type`, `required`, `minimum`, `pattern`, `enum`, ...)
- 📊 **Array-root documents** - Ansible-style playbooks
- 📑 **Multi-document streams** - `---`-separated files via `LoadAll` / `SaveAll`
- 🔗 **Anchors and merge keys** - `&anchor`, `*alias` and `<<: *alias` survive round trips

## 📦 Installation

```bash
go get github.com/Winter0rbit/yamler
```

Requires Go 1.21 or newer. The only dependency is `gopkg.in/yaml.v3`.

## 🚀 Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/Winter0rbit/yamler"
)

func main() {
    // Load YAML with full format preservation
    doc, err := yamler.LoadFile("config.yaml")
    if err != nil {
        log.Fatal(err)
    }

    // Type-safe value retrieval
    appName, _ := doc.GetString("app.name")
    debug, _ := doc.GetBool("app.debug")
    servers, _ := doc.GetStringSlice("app.servers")
    port, _ := doc.GetInt("database.port") // int64

    fmt.Printf("App: %s, Debug: %v, Servers: %v, Port: %d\n", appName, debug, servers, port)

    // Modify values while preserving formatting
    doc.Set("app.version", "2.0")          // new keys are appended to their mapping
    doc.SetBool("app.debug", false)
    doc.AppendToArray("app.servers", "web3")
    doc.SetInt("database.port", 5433)

    // Save with the original formatting intact
    if err := doc.Save("config.yaml"); err != nil {
        log.Fatal(err)
    }
}
```

## 📂 Examples

Runnable examples live in the [`examples/`](examples/) directory, each as its own Go module:

- **[Basic Usage](examples/basic_usage/)** - Fundamental operations and type-safe getters
- **[Comment Alignment](examples/comment_alignment/)** - Comment formatting control
- **[Docker Compose](examples/docker_compose/)** - Service configuration updates
- **[Kubernetes](examples/kubernetes/)** - Multi-document manifest editing and scaling
- **[Ansible](examples/ansible/)** - Playbook management (array-root documents)
- **[Wildcard Patterns](examples/wildcard_patterns/)** - Bulk operations and pattern matching
- **[File Operations](examples/file_operations/)** - Loading, saving and merging files
- **[Advanced Performance](examples/advanced_performance/)** - Caching and bulk operations
- **[Real-World Use Cases](examples/real_world_use_cases/)** - CI/CD and multi-environment configs

```bash
cd examples && ./run_all.sh        # run everything
cd examples/docker_compose && go run main.go
```

## 📚 Guide

### 1. Format Preservation

**Original YAML:**
```yaml
# Production Configuration
app:
  name: myapp           # Application identifier
  version: "1.0"
  debug: yes            # Debug mode flag

  # Server configuration section
  servers: [web1, web2]  # Inline array style

  database:
    host: localhost      # Database host
    port: 5432          # Standard PostgreSQL port
    pools:
      - name: primary
        size: 10
      - name: replica
        size: 5

  # Feature flags
  features:
    - authentication
    - logging
    - metrics
```

**Modifications:**
```go
doc, _ := yamler.LoadFile("app.yaml")

doc.Set("app.version", "2.0")
doc.Set("app.debug", false)
doc.Set("app.database.port", 3306)
doc.Set("app.database.pools[0].size", 20)
doc.AppendToArray("app.servers", "web3")
doc.AppendToArray("app.features", "monitoring")

doc.Save("app.yaml")
```

**Result:**
```yaml
# Production Configuration
app:
  name: myapp           # Application identifier
  version: "2.0"
  debug: false            # Debug mode flag

  # Server configuration section
  servers: [web1, web2, web3]  # Inline array style

  database:
    host: localhost      # Database host
    port: 3306          # Standard PostgreSQL port
    pools:
      - name: primary
        size: 20
      - name: replica
        size: 5

  # Feature flags
  features:
    - authentication
    - logging
    - metrics
    - monitoring
```

Inline comments keep the number of spaces that separated them from the value (relative alignment, the default); use `SetAbsoluteCommentAlignment` to line them up in a column instead.

### 2. Type-Safe Operations

```go
// Strings
name, err := doc.GetString("app.name")
doc.SetString("app.name", "newapp")

// Numbers (integers are int64)
port, err := doc.GetInt("database.port")
timeout, err := doc.GetFloat("app.timeout")
doc.SetInt("database.port", 5432)
doc.SetFloat("app.timeout", 30.5)

// Booleans: true/false, yes/no, on/off and 1/0 are all accepted
debug, err := doc.GetBool("app.debug")
doc.SetBool("app.debug", true)

// Slices
servers, err := doc.GetStringSlice("app.servers")
ports, err := doc.GetIntSlice("app.ports")       // []int64
ratios, err := doc.GetFloatSlice("app.ratios")
flags, err := doc.GetBoolSlice("app.flags")
items, err := doc.GetSlice("app.items")          // []interface{}
pools, err := doc.GetMapSlice("database.pools")  // []map[string]interface{}
doc.SetStringSlice("app.servers", []string{"web1", "web2"})

// Maps
config, err := doc.GetMap("database")
doc.Set("database", map[string]interface{}{
    "host": "localhost",
    "port": 5432,
})

// Anything
value, err := doc.Get("app.database")   // interface{}
doc.Set("app.database.host", "db.local")
```

Paths use dots for keys and `[i]` for array indices: `services.web.ports[0]`. `Set` creates missing intermediate mappings. Keys of new map values are written in sorted order.

### 3. Array Operations

```go
length, _ := doc.GetArrayLength("servers")
first, _ := doc.GetArrayElement("servers", 0)
port, _ := doc.GetTypedArrayElement("ports", 0, "int") // "string", "int", "float", "bool"

doc.AppendToArray("servers", "web3")
doc.InsertIntoArray("servers", 1, "web1.5")
doc.UpdateArrayElement("servers", 0, "web1-updated")
doc.RemoveFromArray("servers", 2)

// Elements are also reachable through paths
doc.Set("servers[0]", "web1")
```

Flow arrays (`[a, b]`, `[ a , b ]`, `[a,b]` and multi-line flow arrays) keep their style when modified; block arrays keep their indentation, including the zero-indent style used by kubectl and GitHub Actions.

### 4. Array-Root Documents (Ansible)

```go
doc, _ := yamler.LoadFile("playbook.yaml")   // root is a list of plays

name, _ := doc.GetString("[0].name")
hosts, _ := doc.GetArrayDocumentElement(0, "hosts")

doc.SetArrayElement(0, "vars.http_port", 8080) // element index + path inside it
doc.Set("vars.http_port", 8080)                // same thing, shorthand for element 0
doc.AddArrayElement(map[string]interface{}{"name": "Extra play", "hosts": "all"})
```

### 5. Wildcard Patterns

```go
// Read all matching values (map of path -> value)
debugFlags, _ := doc.GetAll("**.debug")          // any depth
appPorts, _ := doc.GetAll("apps.*.port")          // one level
firstPorts, _ := doc.GetAll("services.*.ports[0]") // with an index
envNames, _ := doc.GetAll("**.env[*].name")       // any index

// Bulk updates (existing paths only; SetAll does not create keys)
doc.SetAll("**.debug", false)
doc.SetAll("services.*.replicas", 3)

// Matching paths without values
keys, _ := doc.GetKeys("apps.*")     // ["apps.web", "apps.api"]
all, _ := doc.GetPathsRecursive()    // every leaf path in the document
```

### 6. Document Merging

```go
base, _ := yamler.LoadFile("base.yaml")
override, _ := yamler.LoadFile("override.yaml")

err := base.Merge(override)              // whole document
err = base.MergeAt("database", override) // into a sub-tree
```

Values from the source override the target, sequences are replaced as a whole, new keys are appended, and comments of the target are kept where a key already existed.

### 7. Multi-Document Streams

`Load` accepts a single document and returns an error for a `---`-separated stream so that nothing is silently dropped. Use `LoadAll` for streams:

```go
docs, _ := yamler.LoadAllFile("manifests.yaml")   // e.g. Service + Deployment
docs[1].Set("spec.replicas", 5)
yamler.SaveAll("manifests.yaml", docs)             // separators and formatting kept

out, _ := yamler.DocumentsToBytes(docs)            // or get the bytes
```

### 8. Schema Validation

Rules are written as YAML or JSON. Both the library's own type names (`string`, `int`, `float`, `bool`, `array`, `map`, `any`) and the JSON Schema names (`object`, `integer`, `number`, `boolean`) are accepted.

```go
schema := `{
  "type": "object",
  "properties": {
    "app": {
      "type": "object",
      "properties": {
        "name": {"type": "string", "minLength": 1},
        "port": {"type": "integer", "minimum": 1, "maximum": 65535},
        "env":  {"type": "string", "enum": ["dev", "prod"]}
      },
      "required": ["name", "port"]
    }
  }
}`

rule, err := yamler.LoadSchemaFromString(schema) // or LoadSchemaFromFile
if err := doc.Validate(rule); err != nil {
    log.Println("invalid config:", err) // "path app.port: value 0 is less than minimum 1.000000"
}
```

Supported keywords: `type`, `minLength`, `maxLength`, `pattern`, `minimum`, `maximum`, `exclusiveMinimum`, `exclusiveMaximum`, `minItems`, `maxItems`, `uniqueItems`, `items`, `required`, `properties`, `additionalProperties`, `enum`, `nullable`.

### 9. Comment Alignment

```go
doc.EnableRelativeCommentAlignment()   // default: keep original spacing
doc.SetAbsoluteCommentAlignment(20)    // align all inline comments to column 20
doc.DisableCommentAlignment()          // drop inline comments
```

```yaml
# Before
name: myapp    # App name
port: 8080        # Port number
debug: true # Debug flag

# After SetAbsoluteCommentAlignment(20)
name: myapp         # App name
port: 8080          # Port number
debug: true         # Debug flag
```

## 🔧 Error Handling

Every operation returns an error with the path that caused it:

```go
doc, err := yamler.LoadFile("config.yaml") // file or parse errors
value, err := doc.GetString("app.name")    // "path app.name: key app not found"
port, err := doc.GetInt("app.name")        // "path app.name: invalid integer value: ..."
err = doc.Set("items[abc]", 1)             // "path items[abc]: invalid array index: abc"
```

## ⚡ Performance

Formatting information is detected once at load time and cached; parsed paths are cached in a bounded map; serialization reuses pooled buffers. `benchmark_test.go` contains benchmarks for loading, serialization, getters, setters and wildcard operations:

```bash
go test -bench . -benchmem
```

Every mutation re-serializes the document to keep the formatting snapshot current, so batch many changes with `SetAll` or apply them before a single `Save` rather than saving after each one.

## ⚠️ Known Limitations

- Tab indentation is rejected by the YAML parser (the YAML spec forbids tabs for indentation).
- Some formatting hints are keyed by key name rather than by full path, so two keys with the same name but different styles (for example `branches: [main, develop]` and `branches: [main]`) can influence each other's spacing or blank lines.
- Comments after a key without an inline value (`pools:  # note`) or after list items are kept but re-spaced to a single space.
- A comment on a `---` separator line is moved to the following line.
- New arrays created by `Set`/`AppendToArray` use two-space block style; existing arrays keep their own style.

See [FORMATTING_SUPPORT.md](FORMATTING_SUPPORT.md) for the detailed compatibility matrix and [FEATURES.md](FEATURES.md) for a feature walkthrough.

## 📋 API Reference

### Loading and Saving
- `Load(string)`, `LoadBytes([]byte)`, `LoadFile(filename)` - load a single document
- `LoadAll(string)`, `LoadAllBytes([]byte)`, `LoadAllFile(filename)` - load a `---`-separated stream
- `(*Document) String()`, `ToBytes()`, `Save(filename)` - serialize with formatting preserved
- `DocumentsToBytes(docs)`, `SaveAll(filename, docs)` - serialize a stream

### Basic Operations
- `Get(path)`, `Set(path, value)`

### Typed Getters
- `GetString`, `GetInt` (int64), `GetFloat`, `GetBool`
- `GetSlice`, `GetStringSlice`, `GetIntSlice`, `GetFloatSlice`, `GetBoolSlice`, `GetMapSlice`
- `GetMap`

### Typed Setters
- `SetString`, `SetInt` (int64), `SetFloat`, `SetBool`
- `SetStringSlice`, `SetIntSlice`, `SetFloatSlice`, `SetBoolSlice`, `SetMapSlice`

### Arrays
- `GetArrayLength(path)`, `GetArrayElement(path, index)`, `GetTypedArrayElement(path, index, type)`
- `AppendToArray(path, value)`, `InsertIntoArray(path, index, value)`, `UpdateArrayElement(path, index, value)`, `RemoveFromArray(path, index)`

### Array-Root Documents
- `GetArrayDocumentElement(index, path)`, `SetArrayElement(index, path, value)`, `AddArrayElement(value)`

### Wildcards
- `GetAll(pattern)`, `SetAll(pattern, value)`, `GetKeys(pattern)`, `GetPathsRecursive()`
- `FilterByPattern(map, pattern)` - filter a `GetAll` result further

### Merging and Validation
- `Merge(other)`, `MergeAt(path, other)`
- `LoadSchemaFromString(string)`, `LoadSchemaFromFile(filename)`, `(*Document) Validate(rule)`

### Comment Alignment
- `SetCommentAlignment(mode)`, `SetAbsoluteCommentAlignment(column)`, `EnableRelativeCommentAlignment()`, `DisableCommentAlignment()`

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Run `gofmt -s -l .`, `go vet ./...` and `go test -race ./...`
4. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
5. Push to the branch (`git push origin feature/AmazingFeature`)
6. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built on top of the excellent [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) library
- Inspired by the need for format-preserving YAML operations in DevOps workflows

---

**Made with ❤️ for the Go and DevOps communities**
