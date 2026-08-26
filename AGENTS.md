# Repository Guidelines

## Project Structure & Module Organization

`main.go` is the CLI entry point. `finder/` discovers `.http` and `.hurl` files, `parser/http/` owns `.http` parsing, `parser/hurl/` adapts Hurl files, and `runner/` executes requests. Terminal UI code lives in `ui/`, split by component (`tree/`, `suites/`, `suite/`, `producer/`, `footer/`, and layout/theme packages). Tests are colocated with production code as `*_test.go`. Documentation, changelog, license, and preview assets (`preview.png`, `.gif`, `.mov`) live at the repository root.

The request flow is broadly: file discovery → parsing → suite selection → execution → response rendering. Keep package boundaries aligned with that flow. The parser is a hand written scanner: `document.go` splits a file into blocks, `request_text.go` reads one request, and `body_type.go` names the format of a body. It uses no CGO, so keep it that way — the build must stay green with `CGO_ENABLED=0`.

## Build, Test, and Development Commands

- `go run . .` runs the TUI against the current directory.
- `go build ./...` compiles every package. The build is pure Go and needs no C toolchain.
- `go test ./...` runs the standard unit tests.
- `go test -race ./...` matches the CI test mode and checks concurrency safety.
- `go vet ./...` runs Go's static analysis.
- `golangci-lint run ./...` runs the linters CI enforces; the set is pinned in `.golangci.yml`.
- `gofmt -w .` formats Go sources.
- `go mod tidy` normalizes module metadata; verify it does not leave unintended changes.

Before submitting changes, run formatting, vet, the linters, race tests, and the full build.

## Coding Style & Naming Conventions

Follow idiomatic Go and accept `gofmt` output (tabs for indentation). Use exported `PascalCase` names, unexported `camelCase` names, and concise lowercase package names. Existing source filenames use descriptive snake case, such as `on_input_callback.go`. Keep UI components focused and prefer small functions over mixing parsing, execution, and rendering. Propagate cancellation with `context.Context`, wrap useful error context with `%w`, and avoid logging secrets or unredacted authentication headers.

## Testing Guidelines

Use Go's `testing` package and name tests `TestBehavior` or `TestFunction_Scenario`. Prefer table-driven tests for mappings and multiple cases, plus `t.TempDir` for filesystem fixtures. Add regression coverage beside the package being changed. There is no numeric coverage gate, but CI must pass under the race detector without external network access.

## Commit & Pull Request Guidelines

Use the repository's conventional prefixes: `feat:`, `fix:`, `refactor:`, `docs:`, and `chore:`. Keep subjects imperative and specific. Pull requests should explain behavior changes, list verification commands, link relevant issues, and update `CHANGELOG.md` for user-visible changes. Include a screenshot or GIF when modifying TUI layout, focus, colors, or navigation.
