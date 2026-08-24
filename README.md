# lazyrest

[![CI](https://github.com/Hecatoncheir/lazyrest/actions/workflows/go-test.yml/badge.svg)](https://github.com/Hecatoncheir/lazyrest/actions/workflows/go-test.yml)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

`lazyrest` is a terminal UI for discovering and running requests from `.http` and `.hurl` files.

## Preview

![lazyrest screenshot](preview.png)

![lazyrest demo](preview.gif)

[Watch the full MOV preview](preview.mov)

## Features

- Recursive `.http` and `.hurl` discovery with background reloads and common dependency directories ignored.
- Tree-sitter parsing with visible diagnostics and named requests.
- Public/private environment profiles plus recursive `{{variable}}` substitution.
- Cancellable HTTP and Hurl execution with a timeout and bounded response bodies.
- Response headers, protocol metadata, and Pretty/Raw JSON or XML bodies.
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
-env string               environment profile name
-env-file string          public environment file (default "http-client.env.json")
-private-env-file string  private environment file (default "http-client.private.env.json")
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

## Environments

Select a profile with `lazyrest -env development .`. Public values come from `http-client.env.json`:

```json
{
  "development": {
    "host": "api.example.com",
    "api": { "version": "v1" },
    "baseUrl": "https://{{host}}/{{api.version}}"
  }
}
```

Put secrets in `http-client.private.env.json`, which is ignored by Git:

```json
{
  "development": { "token": "local-secret" }
}
```

Use the values as `{{baseUrl}}` or `{{token}}` in `.http` files. Private values override public ones, while declarations inside an `.http` file override both. Undefined variables and reference cycles appear as parser diagnostics. Private values are redacted from requests, responses, errors, and history output.

## Navigation

- `j` / `k` or arrows: move and scroll.
- `Enter`: select a file/request or execute the selected request.
- `Esc`: go back; in the response pane it also cancels the active run.
- `Ctrl+h/j/k/l`: move between areas according to the following map:

| Focused area | Shortcut | Destination |
| --- | --- | --- |
| Files | `Ctrl+l` | Suites |
| Suites | `Ctrl+h` | Files |
| Suites | `Ctrl+j` | Suite |
| Suites | `Ctrl+l` | Producer |
| Suite | `Ctrl+h` | Files |
| Suite | `Ctrl+k` | Suites |
| Suite | `Ctrl+l` | Producer |
| Producer | `Ctrl+h` | Suite |

- `/`: search in the focused Files, Suites, or Producer area; `Enter` finishes entering the query.
- `r`: reload the file tree in the background while Files is focused.
- `p`: toggle Pretty/Raw response bodies while Producer is focused.
- `n` / `N`: next/previous matching file.
- `[` / `]`: previous/next response history entry.
- `q` or `Ctrl+C`: quit.

## Development

```sh
go test ./...
go vet ./...
go build ./...
```
