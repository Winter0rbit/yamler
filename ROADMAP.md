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
- ✅ All formatting hints keyed by full path (indentation, blank lines, flow styles, comments, block scalars)
- ✅ Round-trip corpus of real-world files and a fuzz test (`FuzzRoundTrip`)
- ✅ Benchmarks for load, serialize, get, set, wildcards and arrays (`benchmark_test.go`)
- ✅ Formatting-info and path caches, pooled buffers
- ✅ CI: tests with `-race`, `gofmt`, `go vet`, `goimports`

## 🎯 Priorities

### Phase 1: Remaining layout details
- [ ] Per-item layout records for lists whose items are laid out differently
- [ ] Preserve flow collections that contain comments
- [ ] Inherit the document's dominant list style (zero-indent vs indented) for newly created arrays
- [ ] Grow the round-trip corpus (Helm charts, Compose v2 profiles, OpenAPI specs)

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
- [ ] Run `FuzzRoundTrip` periodically in CI (nightly job)
- [ ] `examples/advanced_performance` lacks a `go.sum` (`go mod tidy`)

## 📈 Success Metrics
- [x] `Load` → `ToBytes` round trip is byte-identical for all corpus fixtures
- [ ] Handle 10 MB+ YAML files
- [x] Zero panics on arbitrary (fuzzed) input (no panic found by `FuzzRoundTrip` so far)
