# lazyrest

`lazyrest` is a terminal UI for discovering and running requests from `.http` and `.hurl` files.

![lazyrest preview](preview.png)

## Features

- Recursive `.http` and `.hurl` discovery with common dependency directories ignored.
- Tree-sitter parsing with visible diagnostics and named requests.
- Variables declared as `@name = "value"` and referenced as `{{name}}`.
- Cancellable HTTP and Hurl execution with a timeout and bounded response bodies.
- File/request/response search and an in-memory history of the last 50 runs.
- Mouse and Vim-style keyboard navigation.

## Requirements

- Go 1.25 or newer when building from source.
- The [`hurl`](https://hurl.dev/docs/installation.html) executable for `.hurl` files. It is not required for ordinary `.http` requests.

## Installation

```sh
go install github.com/Hecatoncheir/lazyrest@latest
```

Or build the repository locally:

```sh
go build -o lazyrest .
```

Tree-sitter uses CGO. Cross-compilation therefore requires a C cross-compiler; for example, from macOS to Windows/amd64:

```sh
brew install mingw-w64
env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build
```

## Usage

```sh
lazyrest [flags] [directory]
```

When the directory is omitted, the current working directory is used.

```text
-timeout duration         request and Hurl timeout (default 30s)
-max-response-bytes int   maximum response bytes kept in memory (default 10485760)
-hurl string              Hurl executable name or path (default "hurl")
```

Example `.http` file:

```http
@host = "api.example.com"
@token = "development-token"

# @name List users
GET https://{{host}}/users
Authorization: Bearer {{token}}
Accept: application/json
```

Sensitive headers such as `Authorization`, cookies, API keys, tokens, and secrets are redacted from the response pane.

## Navigation

- `j` / `k` or arrows: move and scroll.
- `Enter`: select a file/request or execute the selected request.
- `Esc`: go back; in the response pane it also cancels the active run.
- `Ctrl+h/j/k/l`: move between areas.
- `/`: search in the focused Files, Suites, or Producer area; `Enter` finishes entering the query.
- `n` / `N`: next/previous matching file.
- `[` / `]`: previous/next response history entry.
- `q` or `Ctrl+C`: quit.

## Development

```sh
go test ./...
go vet ./...
go build ./...
```
