# YAML Formatting Support in Yamler

This document describes which formatting features Yamler preserves when a document is modified and written back, based on the current test suite (`go test ./...`, 93 top-level tests with ~290 sub-cases, all passing).

## How preservation works

Yamler parses the document with `gopkg.in/yaml.v3`, keeps the original text next to the node tree and records a formatting snapshot (indentation of every key and list item, flow/block styles, blank lines, comment spacing, document markers). After every modification the node tree is encoded again and the encoder output is post-processed to match the snapshot. Lines that existed in the original get their original layout back; lines added by a modification follow the layout of their parent.

## ✅ Fully Supported

### Structure
- **Indentation**: 2, 4, 6, 8 spaces and other consistent widths; keys keep their exact column
- **Zero-indent lists** (kubectl / GitHub Actions / Ansible style), including elements appended later:
  ```yaml
  containers:
  - name: web
    image: nginx
  ```
- **Key order**: never reordered; new keys are appended to the end of their mapping
- **Blank lines** between sections
- **Document markers**: leading `---` and trailing `...`
- **Multi-document streams**: `LoadAll` / `DocumentsToBytes` / `SaveAll` keep every document's own formatting and separators (`Load` rejects streams with more than one document instead of dropping the rest)
- **Array-root documents** (Ansible playbooks) with dedicated methods

### Scalars
- **String styles**: plain, `"double"`, `'single'`
- **Literal (`|`) and folded (`>`) blocks**, including chomping indicators
- **Loose booleans** on read: `true/false`, `yes/no`, `on/off`, `1/0`

### Collections
- **Flow arrays**: `[1, 2, 3]`, compact `[1,2,3]`, spaced `[ 1 , 2 , 3 ]` and multi-line flow arrays keep their style through append/update/remove
- **Flow objects**: `{key: value}` and multi-line flow objects keep their spacing and field order
- **Block arrays** keep their indentation
- **Anchors, aliases and merge keys**: `&anchor`, `*alias`, `<<: *alias` (docker-compose `x-*` extensions)

### Comments
- Head, foot and inline comments are never lost
- Inline comment spacing after `key: value` is preserved (relative mode), or all inline comments can be aligned to a column (`SetAbsoluteCommentAlignment`), or removed (`DisableCommentAlignment`)

### Real-world formats covered by tests
Docker Compose, Kubernetes manifests (single and multi-document), Ansible playbooks, GitHub Actions workflows, generic application configuration files.

## ⚠️ Partially Supported

### Formatting hints keyed by key name
Blank-line patterns, flow-array/flow-object styles and inline comment spacing are currently recorded per **key name**, not per full path. Two keys with the same name but different styles can influence each other:
```yaml
on:
  push:
    branches: [main, develop]   # spaced
  pull_request:
    branches: [main]            # the later definition wins for both
```
Indentation is already recorded per path and is not affected.

### Comments after bare keys and list items
```yaml
pools:            # kept, but re-spaced to "pools: # ..."
  - name: primary # kept, but re-spaced
```
Comments following `key: value` pairs keep their spacing.

### Newly created structures
Arrays and mappings created by `Set`/`AppendToArray` use two-space block style. Map values given as `map[string]interface{}` are written with their keys sorted.

### Separator comments
`--- # comment` is preserved, but the comment moves to the line after the separator.

## ❌ Not Supported

### Tab indentation
The YAML specification forbids tabs for indentation and `yaml.v3` rejects such input; this is not a library limitation.

### Directives
`%YAML` / `%TAG` directives are not preserved.

## 🚀 Recommendations

1. Use `LoadAll` for Kubernetes-style multi-document files.
2. When a document mixes styles for same-named keys, check the output for that key; keying hints by full path is on the roadmap.
3. Prefer one `Save` after a batch of modifications: every mutation re-serializes the document internally.
4. Run your own configuration files through the test in `examples/` (or a quick round-trip test) before relying on the library in production.

---

*Status as of the current test suite. See [ROADMAP.md](ROADMAP.md) for planned improvements.*
