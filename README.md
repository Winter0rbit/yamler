# Yamler

Yamler is a Go library for working with YAML files while preserving formatting, comments, and structure. Unlike many other YAML libraries, Yamler maintains the original file formatting during read and write operations.

## Features

- Preserves comments in YAML files
- Maintains original formatting and structure
- Type-safe operations with YAML content
- Easy to use API

## Installation

```bash
go get github.com/yourusername/yamler
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/yourusername/yamler"
)

func main() {
    // Load YAML file
    doc, err := yamler.LoadFile("config.yaml")
    if err != nil {
        panic(err)
    }

    // Get value preserving comments and formatting
    value, err := doc.Get("some.nested.key")
    if err != nil {
        panic(err)
    }

    // Set value preserving surrounding comments and formatting
    err = doc.Set("some.nested.key", "new_value")
    if err != nil {
        panic(err)
    }

    // Save back to file
    err = doc.Save("config.yaml")
    if err != nil {
        panic(err)
    }
}
```

## License

MIT License 