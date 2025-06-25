# Comment Alignment in Yamler

Yamler supports flexible management of inline comment alignment in YAML documents. This functionality allows maintaining code readability when automatically editing configuration files.

## Alignment Modes

### 1. Relative Alignment (Default)

Preserves the original spacing between value and comment for each line.

```go
doc.EnableRelativeCommentAlignment()
// or
doc.SetCommentAlignment(yamler.CommentAlignmentRelative)
```

**Example:**
```yaml
# Original file
database:
  host: localhost # Primary host
  port: 5432      # Standard port
  timeout: 30     # Connection timeout

# After changing host value
database:
  host: newhost # Primary host     # <- preserves 1 space
  port: 5432    # Standard port    # <- preserves 6 spaces  
  timeout: 30   # Connection timeout # <- preserves 5 spaces
```

### 2. Absolute Alignment

Aligns all comments to the specified column.

```go
doc.SetAbsoluteCommentAlignment(30) // Align to column 30
```

**Example:**
```yaml
# Original file
database:
  host: localhost # Primary host
  port: 5432      # Standard port
  timeout: 30     # Connection timeout

# After setting absolute alignment to column 30
database:
  host: localhost             # Primary host
  port: 5432                  # Standard port  
  timeout: 30                 # Connection timeout
```

### 3. Disabled Comments

Completely removes inline comments from output.

```go
doc.DisableCommentAlignment()
// or
doc.SetCommentAlignment(yamler.CommentAlignmentDisabled)
```

**Example:**
```yaml
# Original file
database:
  host: localhost # Primary host
  port: 5432      # Standard port

# After disabling comments
database:
  host: localhost
  port: 5432
```

## API

### Main Methods

```go
// Set alignment mode
func (d *Document) SetCommentAlignment(mode CommentAlignmentMode)

// Absolute alignment to specified column
func (d *Document) SetAbsoluteCommentAlignment(column int)

// Enable relative alignment
func (d *Document) EnableRelativeCommentAlignment()

// Disable comment processing
func (d *Document) DisableCommentAlignment()
```

### Mode Constants

```go
const (
    CommentAlignmentRelative  // Preserves original spacing
    CommentAlignmentAbsolute  // Aligns to specified column
    CommentAlignmentDisabled  // Disables comments
)
```

## Usage Examples

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/Winter0rbit/yamler"
)

func main() {
    yaml := `
database:
  host: localhost # Primary host
  port: 5432      # Standard port`

    doc, _ := yamler.Load(yaml)
    
    // Absolute alignment
    doc.SetAbsoluteCommentAlignment(25)
    doc.Set("database.host", "newhost")
    
    result, _ := doc.String()
    fmt.Println(result)
    // Output:
    // database:
    //   host: newhost      # Primary host
    //   port: 5432         # Standard port
}
```

### Working with Configuration Files

```go
// Load existing config
doc, err := yamler.LoadFile("config.yml")
if err != nil {
    log.Fatal(err)
}

// Configure alignment for readability
doc.SetAbsoluteCommentAlignment(40)

// Update values
doc.Set("app.version", "2.0")
doc.Set("database.timeout", 60)

// Save with preserved formatting
err = doc.Save("config.yml")
```

### Dynamic Mode Switching

```go
doc, _ := yamler.Load(yamlContent)

// For development - show all comments with alignment
doc.SetAbsoluteCommentAlignment(35)
devConfig, _ := doc.String()

// For production - remove comments
doc.DisableCommentAlignment()  
prodConfig, _ := doc.String()

// For documentation - preserve original formatting
doc.EnableRelativeCommentAlignment()
docConfig, _ := doc.String()
```

## Implementation Features

1. **Performance**: Alignment information is cached for reuse
2. **Compatibility**: Works with all existing Yamler functions
3. **Flexibility**: Alignment mode can be changed at any time
4. **Safety**: Incorrect settings don't cause errors

## Limitations

- Only processes inline comments (on the same line as values)
- Comments on separate lines are not affected
- For absolute alignment, if comment doesn't fit in specified column, at least one space is added

## Compatibility

Comment alignment functionality is compatible with all Yamler operating modes:
- ✅ Regular YAML documents
- ✅ Array-root documents (Ansible style)  
- ✅ Indentation preservation (2, 4, 6, 8 spaces)
- ✅ Flow and block styles
- ✅ Multiline values
- ✅ Nested structures

Full usage example is available in `examples/comment_alignment.go`. 