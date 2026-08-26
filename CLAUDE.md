# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.
`AGENTS.md` holds the full contribution conventions; this file covers what is
easiest to get wrong.

## Commands

- Build: `go build ./...`
- Test: `go test ./...`
- Test (race, what CI runs): `go test -race ./...`
- Test (one package): `go test -v ./parser/http`
- Terminal integration tests: `go test ./ui -run TUI`
- Lint: `golangci-lint run ./...` (the set is pinned in `.golangci.yml`)
- Vet: `go vet ./...`
- Format: `gofmt -w .`
- Run: `go run . example`

## Invariants

- **No CGO.** The build must stay pure Go; CI compiles every released target
  with `CGO_ENABLED=0` and fails otherwise. Tree-sitter was removed for this
  reason, so do not reintroduce a C dependency to parse or highlight anything.
- **Nothing secret reaches disk.** Private environment values, sensitive
  headers, and captured responses are redacted or kept in memory only. The
  cookie jar and the response store live for the session; persisted history is
  redacted and bounded before writing. Explicit response export writes the same
  redacted body shown by Producer to a user-selected `0600` file.
- Every package has tests. `golangci-lint` reports zero issues; keep it that
  way rather than adding exclusions.

## Architecture

A terminal UI for reading and running `.http` and `.hurl` files. The flow is:
file discovery → parsing → suite selection → execution → response rendering.

### Tech stack

- **Language:** Go 1.25, no CGO
- **TUI:** `tview` and `tcell`
- **Parsing:** a hand written scanner

### Packages

- `main.go`: flags, configuration layering, and the client built once per
  session.
- `parser/http/`: reads `.http` files. `document.go` splits a file into comment,
  variable, and request blocks; `request_text.go` reads one request into a
  request line, headers, and a body; `body_type.go` names the format of a body;
  `variables.go` substitutes `{{variable}}`; `response_reference.go` fills in
  `{{name.response…}}` when a request runs; `graphql.go` splits a GraphQL
  document from its variables; `redact.go` hides secrets.
- `parser/hurl/`: splits a `.hurl` file into its entries.
- `runner/`: executes a request or shells out to `hurl`.
- `finder/`: walks a directory tree for request files.
- `environment/`, `config/`, `keymap/`, `locale/`: environment profiles, layered
  YAML configuration, key bindings, and translations.
- `ui/`: the widgets — `tree/`, `suites/`, `suite/`, `producer/`, `footer/`,
  `workspace/`, `layout/` — plus `theme/`, `syntax/`, `progress/`, `symbols/`.

### Things worth knowing before editing

- `HttpSuite.Header` is a `net/http.Header`, so a name can carry several values.
  Do not reduce it to a map of strings.
- A `{{name.response…}}` reference is resolved when the request runs, not while
  the file is parsed, because it depends on what has already been run. The
  parser deliberately leaves those references alone. Captured responses are
  keyed by both `HttpSuite.SourceFilePath` and the request name; do not let a
  name resolve across files.
- Highlighting colours come from `theme.Theme.Syntax` and `theme.Theme.Methods`,
  which every pane reads; they are not per-widget.
- `ui/syntax` escapes its own output. Text that goes into a `tview` widget with
  markup enabled must be escaped exactly once — `tview.List` reads style tags in
  item text by default.
- A `.hurl` entry is run with `--to-entry`, never `--from-entry`: an entry may
  use what an earlier one captured.
