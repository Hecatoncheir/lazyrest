# Changelog

All notable changes to this project will be documented in this file.

## [v0.4.0] - 2026-08-24
### Added
- Cancellable background suite parsing and file-tree reloads with stale-result protection.
- Public/private HTTP environment profiles with recursive variables and secret redaction.
- HTTP response headers, protocol metadata, and Pretty/Raw JSON or XML rendering.

### Changed
- Parser diagnostics now report recursive variable cycles.
- Project licensing changed from GPL-3.0 to MIT.

## [v0.3.2] - 2026-08-24
### Changed
- Updated `Ctrl+h/j/k/l` focus navigation between Files, Suites, Suite, and Producer.

## [v0.3.1] - 2026-08-24
### Changed
- README preview now includes the screenshot, animated demo, and full MOV recording.

### Removed
- Unused root-level Hurl scratch files and the unused Hurl source submodule.
- Empty footer source files.

## [v0.3.0] - 2026-08-24
### Added
- Cancellable request execution with configurable timeouts and response-size limits.
- File, request, and response search plus an in-memory response history.
- `.http` variable substitution and visible parser diagnostics.

### Changed
- Hurl files are represented and executed as complete Hurl sessions.
- Tree-sitter parser and tree ownership is explicit and no longer copied by value.
- Sensitive request headers are redacted in the response pane.
- CI, module metadata, and usage documentation have been updated.

### Fixed
- Hurl execution now receives the selected file path.
- Stale background requests can no longer overwrite newer UI results.
- Footer breadcrumbs retain the selected request name.

## [v0.2.9] - 2026-08-21
### Fixed
- Resolved Hurl test failures.
- Cleaned up repository by removing untracked `.cache` files.

## [v0.2.8] - 2024-05-23
### Fixed
- Resolved build failure in `ui/footer` caused by an invalid type assertion.
- Corrected argument types in `finder` and `ui/tree` tests to match function signatures.

## [v0.2.7] - 2024-05-23
### Fixed
- Build failure in `ui/footer` due to incorrect type assertion.
- Test failures in `finder` and `ui/tree` due to incorrect argument types.

## [v0.2.6] - 2024-05-23
### Added
- Enhanced navigation using `Ctrl+h/j/k/l` keys for area switching.
- Status color indication in the Producer (green for 2xx, yellow for 3xx, red for 4xx/5xx).
- Improved visual separation and request details display in the Producer area.
- Automatic footer update when switching suites.

## [v0.2.5] - 2026-08-21
### Added
- Support for `.hurl` files in the file tree and parser.

## [v0.2.4] - 2026-08-21
### Added
- Progress bar indicator in the Producer area during network requests to provide visual feedback.

## [v0.2.3] - 2026-08-21
### Fixed
- HTTP parser tests to match tree-sitter grammar requirements.

### Added
- Enhanced test coverage for `finder` and `runner` packages.
- Improved `Content-Type` handling in `runner` package.
- Unit tests for `runner.Response`.

## [v0.2.2] - 2026-08-21
### Added
- Asynchronous HTTP requests in the Producer component to prevent UI freezing.

## [v0.2.1] - 2026-08-21
### Fixed
- Case-insensitive extension matching in `finder` package

## [v0.2.0] - 2026-08-20
### Added
- Case-insensitive extension matching in `finder` package

### Fixed
- None

## [0.1.0] - 2026-08-21
### Added
- File discovery subsystem (`finder`)
- HTTP request parser (`parser/http`)
- File tree UI component (`ui/tree`)
- HTTP request runner (`runner`)
- Comprehensive test suite for core components

### Fixed
- Compilation errors in `ui/suite`, `ui/suites`, and `runner` packages
- Inconsistent file discovery logic in `finder` package
