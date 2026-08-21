# Changelog

All notable changes to this project will be documented in this file.

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
