# Yamler Development Roadmap

## Current State

**Project stats** (see `go test -cover ./...`):
- Core code: ~4 000 lines across topic files (`document.go`, `formatting_*.go`, `flow_styles.go`, `array.go`, ...)
- Tests: 93 top-level tests with ~290 sub-cases, all passing under `-race`; ~77% statement coverage
- Dependencies: only `gopkg.in/yaml.v3`

**Done:**
- ✅ Formatting preservation: indentation (including zero-indent lists), comments, blank lines, flow/block styles, literal/folded scalars, document markers
- ✅ Type-safe getters/setters, array CRUD, array-root documents
- ✅ Wildcards (`*`, `**`, `[*]`, indices), document merging, schema validation (JSON Schema type names accepted)
- ✅ Multi-document streams (`LoadAll` / `SaveAll`)
- ✅ Anchors, aliases and merge keys survive round trips
- ✅ Path-based exact indentation (keys and list items keyed by full path)
- ✅ Benchmarks for load, serialize, get, set, wildcards and arrays (`benchmark_test.go`)
- ✅ Formatting-info and path caches, pooled buffers
- ✅ CI: tests with `-race`, `gofmt`, `go vet`, `goimports`

## 🎯 Priorities

### Phase 1: Path-keyed formatting hints
**Goal**: no more cross-talk between same-named keys.

- [ ] Blank lines (`EmptyLines`) keyed by path via `lineWalker`
- [ ] Flow array / flow object styles keyed by path
- [ ] Inline comment spacing keyed by path; keep spacing for comments after bare keys and list items
- [ ] Inherit the document's dominant list style (zero-indent vs indented) for newly created arrays

### Phase 2: API completeness
- [ ] `Has(path)` / `Delete(path)` / `Keys(path)` / `Copy()`
- [ ] Typed error values (`ErrNotFound`, `ErrType`, `ErrIndex`) usable with `errors.Is`
- [ ] Array slices in paths (`items[1:3]`)
- [ ] Keep insertion position for new keys (e.g. insert after a given key)

### Phase 3: Performance
- [ ] Avoid re-serializing the whole document on every mutation (serialize lazily on `ToBytes`)
- [ ] Profile and benchmark 1 MB+ and 10 MB+ documents; target memory < 2x file size
- [ ] Performance regression benchmarks in CI

### Phase 4: Format details
- [ ] Preserve `%YAML` / `%TAG` directives
- [ ] Keep comments on `---` separator lines in place
- [ ] Preserve CRLF line endings
- [ ] Preserve key order of `map[string]interface{}` values (ordered input type)

### Phase 5: Tooling
- [ ] Configuration diff between two documents
- [ ] Environment variable substitution helpers
- [ ] JSON ↔ YAML conversion with formatting

## 🔧 Technical Debt
- [ ] Reduce the remaining key-name-based heuristics in `flow_styles.go` and `comments.go`
- [ ] Property-based / fuzz tests for round-trip stability (`Load` → `ToBytes` must be identity for untouched documents)
- [ ] `examples/advanced_performance` lacks a `go.sum` (`go mod tidy`)

## 📈 Success Metrics
- [ ] `Load` → `ToBytes` round trip is byte-identical for all test fixtures
- [ ] Handle 10 MB+ YAML files
- [ ] Zero panics on arbitrary (fuzzed) input
