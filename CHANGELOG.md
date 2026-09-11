# Changelog

All notable changes to this project are documented in this file. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.3.2] - 2026-09-10

Restores `OrderedMap`, which v1.3.0 dropped by accident, and completes the
round of data-preservation fixes. This is the release to use: v1.3.0 and
v1.3.1 both point at a commit that predates all of it.

### Fixed

- **`OrderedMap` is back.** It is part of the public API of v1.2.4 and was
  removed from v1.3.0 unintentionally; code using it did not compile against
  that release. It now also keeps key order, which it never did: its
  `MarshalYAML` used to return a plain map, losing the order the type exists
  to preserve.
- **Flow collections were split on every comma**, so `['a, b']` could become
  `['a', 'b']`, and `{'a: b': 1}` could be split on the colon inside its key.
  Splitting is now aware of quoting.
- **Blank lines inside a multi-line flow collection were dropped**, although
  they are part of a plain scalar's value.
- **`key: # comment`** was read as having the value `#`, which made the block
  below the key look like the continuation of a scalar.
- **A quoted key followed by a space before its colon** (`'k' : v`) was not
  recognised as a key, so its line kept the encoder's indentation.
- **Continuation lines of a multi-line plain scalar** were taken for
  structure; a line such as `- "` below `a: 0` is content, not a list item.
- **An empty quoted key** (`"": v`) was indistinguishable from a line without
  a key.
- Documents whose root is a scalar, keys longer than 128 characters (which
  `yaml.v3` writes in explicit `? key` form) and block scalars whose content
  starts with a whitespace-only line are now written by the encoder without
  formatting restoration, instead of being re-indented into a different
  document. Their data is preserved; their layout may change.

### Added

- `Set` accepts an `OrderedMap` and writes its keys in the given order. A
  plain `map[string]interface{}` is still written sorted, since Go maps have
  no order of their own.

### Changed

- CI now runs the linter that the repository has been configured for but
  never ran, a short fuzzing round on every pull request and a longer one
  nightly, and tests on both the oldest supported Go version and the current
  one. A tag push is checked against the changelog and against `main`, so a
  release cannot again point at a commit that predates it.
- **Serialization is lazy.** Mutations only change the node tree; the document
  is rendered by `String`, `ToBytes`, `Save` and `DocumentsToBytes`. Editing
  100 keys of a 100-service document and saving once went from 301 ms and
  295 MB of allocations to 5.4 ms and 3.7 MB, and rendering is now idempotent:
  the result no longer depends on how often the document was rendered along
  the way.
- A document built from scratch (`Load("")`) ends with a newline like any
  other, instead of only growing one on the second render.

## [1.3.1] - 2026-09-09

Published by mistake: the tag was placed on the same commit as v1.3.0 before
the release branch had been merged, so this version is byte for byte
identical to v1.3.0 and contains none of the changes listed above. Use
v1.3.2.

## [1.3.0] - 2026-09-09

The first release since v1.2.4. It fixes several ways in which the library
changed or lost data in exactly the file formats it advertises support for.

### Fixed

- **Merge keys were rewritten.** `<<: *defaults` was written back as
  `!!merge <<: *defaults`, breaking every docker-compose file that uses `x-*`
  extension anchors. (A `gopkg.in/yaml.v3` v3.0.1 quirk, now worked around.)
- **Multi-document streams lost data.** `Load` silently kept only the first
  document of a `---`-separated stream and dropped the rest on save — a
  Kubernetes manifest with a Service and a Deployment came back with only the
  Service.
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
- A round-trip test corpus of real-world files (docker-compose, multi-document
  Kubernetes, GitHub Actions, Ansible, Helm values) that must survive
  `Load` → `ToBytes` byte for byte, and a `FuzzRoundTrip` fuzz test.

### Changed

- `Load` returns `ErrMultiDocument` for a `---`-separated stream instead of
  silently discarding all but the first document. This is the one intended
  behaviour change that can break existing code; use `LoadAll` for such input.
- `document.go` was split into files by concern, and ~830 lines of unreachable
  code were removed. `coverage.out`, `test_results.log` and a duplicated
  `example/` directory are no longer part of the repository.
- Documentation now matches the implementation: the previous README documented
  methods that did not exist (`GetIntArrayElement`, `LoadSchema`) and claimed
  100% test coverage.

### Removed

- `OrderedMap` — unintentionally, as part of a dead-code cleanup. It is
  restored in v1.3.2; prefer that release.

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

[1.3.2]: https://github.com/Winter0rbit/yamler/compare/v1.3.1...v1.3.2
[1.3.1]: https://github.com/Winter0rbit/yamler/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/Winter0rbit/yamler/compare/v1.2.4...v1.3.0
[1.2.4]: https://github.com/Winter0rbit/yamler/compare/v1.2.3...v1.2.4
[1.2.3]: https://github.com/Winter0rbit/yamler/compare/v1.2.2...v1.2.3
[1.2.2]: https://github.com/Winter0rbit/yamler/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/Winter0rbit/yamler/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/Winter0rbit/yamler/compare/v1.1.1...v1.2.0
[1.1.1]: https://github.com/Winter0rbit/yamler/releases/tag/v1.1.1
