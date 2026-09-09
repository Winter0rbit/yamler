# Changelog

All notable changes to this project are documented in this file. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.0] - 2026-09-09

The first release since v1.2.4. It fixes several ways in which the library
changed or lost data in exactly the file formats it advertises support for,
adds multi-document and structural APIs, and makes editing large documents
orders of magnitude cheaper.

### Fixed

- **Merge keys were rewritten.** `<<: *defaults` was written back as
  `!!merge <<: *defaults`, breaking every docker-compose file that uses `x-*`
  extension anchors. (A `gopkg.in/yaml.v3` v3.0.1 quirk, now worked around.)
- **Multi-document streams lost data.** `Load` silently kept only the first
  document of a `---`-separated stream and dropped the rest on save — a
  Kubernetes manifest with a Service and a Deployment came back with only the
  Service. `Load` now returns an error for such input and `LoadAll` handles it
  properly.
- **Zero-indent lists were re-indented.** The kubectl / GitHub Actions /
  Ansible style
  ```yaml
  containers:
  - name: web
  ```
  was rewritten with the items indented, producing noisy diffs on every edit.
- **A leading `---` was dropped** unless the document had been modified with
  `Set`, and the array-document methods removed it as well.
- **A trailing `...` was invented** whenever the last line happened to end with
  three dots (for example `run: go test ./...`).
- **Flow collections were split on every comma**, so `['a, b']` could become
  `['a', 'b']` and `{'a: b': 1}` could be split on the colon inside the key.
- **Blank lines inside a multi-line flow collection were dropped**, although
  they are part of a plain scalar's value.
- **`key: # comment`** was read as having the value `#`, which made the block
  below the key look like a scalar continuation.
- **A quoted key followed by a space before its colon** (`'k' : v`) was not
  recognised as a key.
- **Continuation lines of multi-line plain scalars** were taken for structure.
- **An empty quoted key** (`"": v`) was indistinguishable from a line without
  a key.
- **Formatting hints no longer leak between same-named keys.** Blank lines,
  flow styles, comment spacing and indentation are recorded per path, so a
  `branches: [main, develop]` in one job no longer reformats a
  `branches: [main]` in another.
- Wildcards: `**` now matches zero or more segments, so `**.debug` also matches
  a top-level `debug`; patterns combining wildcards with array indices
  (`services.*.ports[0]`) match at all, which they previously did not.
- `GetBool` accepts the numeric `1`/`0` forms the documentation promised.
- `Get` and the typed getters work on array-root documents (`[0].name`).
- `Validate` accepts JSON Schema type names (`object`, `integer`, `number`,
  `boolean`) alongside the library's own.

### Added

- **Multi-document streams**: `LoadAll`, `LoadAllBytes`, `LoadAllFile`,
  `DocumentsToBytes`, `SaveAll`. Each document keeps its own formatting and
  separators.
- **Structural operations**: `Has`, `Keys` (in document order), `Copy` (deep
  copy including formatting), `Delete` and `DeleteAll` (wildcard, returns the
  number of removed entries).
- **Sentinel errors** for `errors.Is`: `ErrNotFound`, `ErrType`, `ErrIndex`,
  `ErrPath`, `ErrRoot`, `ErrParse`, `ErrIO`, `ErrMultiDocument`,
  `ErrValidation`, `ErrUnsupported`. Errors from `os` and `yaml.v3` remain
  reachable through `errors.Is` / `errors.As`.
- `OrderedMap` values passed to `Set` now keep their key order (a plain
  `map[string]interface{}` is still written with sorted keys, since Go maps
  have no order).
- A round-trip test corpus of real-world files (docker-compose, multi-document
  Kubernetes, GitHub Actions, Ansible, Helm values) that must survive
  `Load` → `ToBytes` byte for byte, and a `FuzzRoundTrip` fuzz test.

### Changed

- **Serialization is lazy.** Mutations only change the node tree; the document
  is rendered by `String`, `ToBytes`, `Save` and `DocumentsToBytes`. Editing
  100 keys of a 100-service document and saving once went from 301 ms and
  295 MB of allocations to 5.4 ms and 3.7 MB, and rendering is now idempotent.
- `Load` returns `ErrMultiDocument` for a `---`-separated stream instead of
  silently discarding all but the first document. This is the one behaviour
  change that can break existing code; use `LoadAll` for such input.
- `OrderedMap.MarshalYAML` now actually preserves key order. It previously
  returned a plain map, which lost the order the type exists to keep.
- Documents whose root is a scalar, keys longer than 128 characters (which
  `yaml.v3` writes in explicit `? key` form) and block scalars whose content
  starts with a whitespace-only line are written by the encoder without
  formatting restoration; their data is preserved, their layout may change.
- `document.go` was split into files by concern, and ~830 lines of unreachable
  code were removed. `coverage.out`, `test_results.log` and a duplicated
  `example/` directory are no longer part of the repository.
- Documentation now matches the implementation: the previous README documented
  methods that did not exist (`GetIntArrayElement`, `LoadSchema`) and claimed
  100% test coverage.

## [1.2.4] - 2026-02-03

- Fix array-root `Set` and validation edge cases.

## [1.2.3] - 2025-07-02

- GitHub Actions fixes and formatting improvements.

## [1.2.2] - 2025-07-01

- Formatting preservation improvements.

## [1.2.1] - 2025-07-01

- Fix formatting preservation issues.

## [1.2.0] - 2025-06-25

- Comment alignment modes, array style preservation, performance work.

## [1.1.1] - 2025-06-09

- Move the package to the repository root for a clean import path.

[1.3.0]: https://github.com/Winter0rbit/yamler/compare/v1.2.4...v1.3.0
[1.2.4]: https://github.com/Winter0rbit/yamler/compare/v1.2.3...v1.2.4
[1.2.3]: https://github.com/Winter0rbit/yamler/compare/v1.2.2...v1.2.3
[1.2.2]: https://github.com/Winter0rbit/yamler/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/Winter0rbit/yamler/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/Winter0rbit/yamler/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/Winter0rbit/yamler/releases/tag/v1.1.1
