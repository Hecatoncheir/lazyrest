# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with
code in this repository. `AGENTS.md` holds the full contribution conventions;
this file covers what is easiest to get wrong.

## Commands

`AGENTS.md` carries the everyday commands — build, test, race test, vet, lint,
format, `go mod tidy` — and is the place to change them. These are the ones it
does not cover:

- Test (one package): `go test -v ./parser/http`
- Test (one test): `go test ./ui -run TestTUIChainsRequestsThroughAnEarlierResponse`
- Terminal integration tests: `go test ./ui -run TUI`
- Fuzz one target: `go test ./parser/http -run=^$ -fuzz=^FuzzResolveVariables$ -fuzztime=2s`
- Fuzz everything the way CI does: `.github/scripts/fuzz.sh 10s 2`
- Run against the bundled samples: `go run . example`

The app is a TUI and cannot be driven from a shell directly. To see a change
working, follow `.claude/skills/run-lazyrest/SKILL.md`: it wraps the app in
tmux and bundles the servers a stream needs.

CI enforces more than `go test -race ./...`. It also fails on anything
`gofmt -l .` prints, on a `go mod tidy` that leaves a diff, on `shellcheck`
over every tracked `*.sh`, on `govulncheck`, on a ten second smoke run of every
fuzz target, and on a `CGO_ENABLED=0` build of all five release targets. Both
fuzz jobs call
`.github/scripts/fuzz.sh <fuzztime> <parallel>`, which discovers the targets
with `go test ./... -list='^Fuzz'` rather than listing them, so a new target
needs no CI change; a discovery that comes back empty fails the run instead of
quietly fuzzing nothing.

## Invariants

- **No CGO.** The build must stay pure Go; CI compiles every released target
  with `CGO_ENABLED=0` and fails otherwise. Tree-sitter was removed for this
  reason, so do not reintroduce a C dependency to parse or highlight anything.
- **Nothing secret reaches disk.** Private environment values, sensitive
  headers, and captured responses are redacted or kept in memory only. The
  cookie jar and the response store live for the session; persisted history is
  redacted and bounded before writing. Explicit response export writes the same
  redacted body shown by Producer to a user-selected `0600` file.
- Every package holding logic has tests. `ui/symbols`, which is four glyph
  constants, is the only package without them. `golangci-lint` reports zero
  issues; keep it that way rather than adding exclusions. The single exclusion
  that exists exempts test files from `gocognit`, where a table is long by
  design.
- **A function stays under a cognitive complexity of 25**, which `gocognit`
  enforces. The usual fix is to split a long dispatch into named steps, one per
  thing it decides. What sits nearest the bound is a hand written scanner, the
  flag parsing in `main`, and one read loop: branchy by nature, where splitting
  would add indirection rather than clarity.

## Architecture

A terminal UI for reading and running `.http`, `.hurl` and `.socket` files. The
flow is: file discovery → parsing → suite selection → execution → response
rendering. A request that opens a long lived connection — a WebSocket, an MQTT
session, a raw socket — takes a different path after selection: `runner/stream`
holds the connection open and `ui/stream` shows the frames.

### Tech stack

- **Language:** Go 1.25, no CGO
- **TUI:** `tview` and `tcell`
- **Parsing:** a hand written scanner

### Packages

- `main.go`: flags, configuration layering, and the client built once per
  session. `config.LoadFiles` merges the user file, then the project file, then
  an explicit `--config`, onto `DefaultDocument()`, and validates the keymap,
  locale, and theme while building `Settings`. A configuration error surfaces
  there, not in the UI.
- `parser/http/`: reads `.http` files. `document.go` splits a file into comment,
  variable, and request blocks; `request_text.go` reads one request into a
  request line, headers, and a body; `body_type.go` names the format of a body;
  `variables.go` substitutes `{{variable}}`; `response_reference.go` fills in
  `{{name.response…}}` when a request runs; `graphql.go` splits a GraphQL
  document from its variables; `redact.go` hides secrets.
- `parser/hurl/`: splits a `.hurl` file into its entries.
- `runner/`: executes a request or shells out to `hurl`.
- `runner/stream/`: the connections that outlive a single request. `tcp.go`,
  `websocket.go` and `mqtt.go` each satisfy `Session`; `frame.go` holds the
  `Frame` they all produce.
- `finder/`: walks a directory tree for request files.
- `environment/`, `config/`, `keymap/`, `locale/`, `color/`: environment
  profiles, layered YAML configuration, key bindings, translations, and the
  hexadecimal `Color` that themes are written in.
- `ui/`: the root package owns the `Application`, the `Model`, the overlays, and
  the callbacks that join the widgets. The widgets themselves live one level
  down — `tree/`, `suites/`, `suite/`, `producer/`, `footer/`, `workspace/`,
  `layout/` — beside `theme/`, `syntax/`, `progress/`, `symbols/`.

### How the UI is wired

- `ui.Run` calls `BuildApplication` in `ui/ui.go`, and that function is the one
  place where widgets are joined together. A widget never reaches back for the
  `Application`; it is handed `On…Callback` functions and calls them. Adding a
  widget interaction means adding a callback to its `Parameters`, not an import.
- Every widget package follows the same shape: `New()` returns the widget, then
  `Build(parameters)` constructs the `tview` primitive and stores it on
  `Element`. `Parameters` carries the theme, keybindings, locale, and callbacks.
- `Model` guards `State` with an `RWMutex`. Read with `Snapshot()`, which hands
  back a deep clone, and write with `Model.update(func(*State))`. Do not keep a
  `State` around and expect it to track the model.
- `tview` is not concurrency safe. Anything touching a widget from a goroutine —
  startup, the file watcher, footer progress, a running request — goes through
  `Element.QueueUpdateDraw`.
- `Application.loadEnvironment` and `Application.scanFiles` are function fields
  precisely so tests can replace them. Keep new I/O seams in that form.
- Sizing happens in one place: `SetBeforeDrawFunc` calls `updateResponsiveUI`,
  which resizes the workspace, the footer hints, and every overlay from the
  screen size and the focused pane. Do not compute widths in a widget.

### How the terminal tests work

`ui/integration_test.go` holds the harness the `TestTUI…` tests share.
`runTestApplication` puts a `tcell.SimulationScreen` behind an application built
by `BuildApplication`, runs it on its own goroutine, and registers the cleanup
that cancels the watchers and stops it. Drive the UI with `screen.InjectKey`,
and assert with `waitFor` and `waitForScreenText` against `applicationText`,
which flattens the screen to text. These tests poll to a deadline; never add a
sleep, and never assert on a single draw.

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
- **Which file a request belongs in is decided by the session, not the
  protocol.** A WebSocket and an MQTT session live in `.http` because they want
  the variables, cookies and captured responses around them — a broker password
  is commonly a token an HTTP login returned. A raw socket authenticates inside
  its own protocol, so it lives in `.socket` and gives up chaining, which a
  reference keyed by source file cannot cross. STOMP needs no file of its own:
  it rides on a raw socket or on a WebSocket, which is how brokers serve it.
- Only the dial of a stream is bounded by the timeout; the session that follows
  is not. `runner.Runner` bounds a whole run and returns one response, which is
  why a stream is not one. It refuses a stream transport rather than failing
  obscurely.
- Frames never reach `ui.Model`. A snapshot deep copies every suite in the file
  and `refreshStatus` takes one on each call, so a busy connection would clone
  the request list per frame. `ui/stream.Log` carries its own lock, and the
  pane repaints on a ticker rather than per frame.
- A `Frame` carries `Attributes` for what the payload does not say — an MQTT
  topic, a quality of service. Redact them as you redact a body: a topic is
  built from the same variables.
