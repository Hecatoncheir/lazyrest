# Changelog

All notable changes to this project will be documented in this file.

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
