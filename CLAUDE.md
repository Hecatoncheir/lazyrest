# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- Build: `go build`
- Test: `go test ./...`
- Test (verbose): `go test -v ./...`
- Test (package): `go test -v <package_path>`
- Run: `go run main.go [directory]`

## Architecture

A Terminal User Interface (TUI) application for parsing and executing `.http` and `.hurl` files.

### Tech Stack
- **Language:** Go
- **TUI Framework:** `tview` and `tcell`
- **Parsing:** Tree-sitter (via `go-tree-sitter`)

### Core Modules
- `main.go`: Entry point; initializes the UI with a target directory.
- `ui/`: Modular TUI implementation using a widget-based architecture:
    - `workspace/`: Orchestrates major components like the file tree, suites list, and producer.
    - `tree/`: Displays a navigable hierarchy of `.http` and `.hurl` files.
    - `suite/` & `suites/`: Manage selection and display of HTTP suites within a file.
    - `producer/`: Handles the execution and display of individual requests.
    - `layout/` & `footer/`: Manage the overall screen layout and status information.
- `finder/`: Implements file discovery logic, searching for relevant files in a directory tree.
- `parser/`: Wraps tree-sitter to parse HTTP request files.
- `runner/`: Executes the parsed HTTP requests.
- `color/`: Provides color management for the TUI elements.
