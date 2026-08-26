# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]
### Fixed
- A request held back because a response reference could not be resolved now settles as a failed run. It was left in the running state, so the footer animated a request that was never sent.

### Changed
- In Files, `l` opens the selected file, while `Ctrl+l` only moves focus to Suites instead of opening the file as a side effect.

## [v0.14.0] - 2026-08-26
### Added
- `ignore` in the configuration names directories the file tree skips, on top of a built-in list that now also covers `.venv`, `.tox`, `target`, `dist`, and `build`. The lists of the configuration layers add up.
- Symbolic links to directories are followed, each one once, so a tree that links back into itself still terminates. They were skipped entirely before.

### Fixed
- An unknown key in the configuration is refused instead of being dropped without a word, so a typo such as `keybinding` for `keybindings` no longer looks accepted.
- `color.Color.ToRGB` returned 16-bit channels where 0 to 255 was meant. The colours on screen were right only because tcell masks each channel, so a theme rendered correctly by coincidence.
- The summary of a response is translated. It was built with English baked in, so a Russian or Chinese interface still read "Response code" and "Content length".

### Changed
- CI runs `golangci-lint` and compiles every released target on each pull request, so a break on another platform is caught before the tag rather than during a release.
- A scan stops at 32 directories deep and says so in the diagnostics window, rather than descending without limit.

## [v0.13.0] - 2026-08-26
### Added
- A request can use what an earlier one answered: `{{login.response.body.$.token}}` reads a value out of a JSON body, `{{login.response.body}}` takes it whole, and `{{login.response.headers.X-Token}}` takes a response header. The path supports member names and array indices under a `$` root. References are resolved when the request runs, from the last answer of each named request in the session, and nothing is written to disk. A request whose references cannot be resolved is not sent, and the pane says which one was waiting and why.
- A `.hurl` file is listed one entry at a time instead of as a single opaque session. Selecting an entry runs the file up to it, because an entry may use what an earlier one captured; the last entry runs the whole file as before.
- Cookies set by a response are carried into the requests that follow, so a session survives a whole run. The jar is held in memory only; `-cookies=false` turns it off.
- `-max-redirects` bounds how many redirects a request follows, and `-follow-redirects=false` returns the redirect itself so that its `Location` can be read.
- `-insecure` accepts any server certificate, for a host serving a self-signed one.
- The Suites list leads every row with its HTTP method, coloured by what the method does: reads take the success colour of the theme, a create its progress colour, an update its accent, and a delete its failure colour. The row carrying the selection is drawn plain, so the selection stays readable.
- A `-version` flag that reports the build. Release archives carry the tag, and `go install` builds report the module version.

### Fixed
- `lazyrest -h` now prints the usage and exits with `0` instead of reporting `flag: help requested` as a fatal error.

## [v0.12.0] - 2026-08-26
### Added
- Syntax highlighting for JSON, XML, and GraphQL across the Suites, Suite, and Producer panes, covering the body preview of the list, the selected request, the GraphQL variables, and the response. The colours are derived from the active theme, so every preset and override matches. Raw view stays unhighlighted, and bodies over 256 KiB are shown plain so that drawing stays responsive.
- GraphQL requests are sent as `application/json` with a `{"query": …, "variables": …}` body, the form the GraphQL over HTTP specification requires. A request is recognized from its body or from `X-REQUEST-TYPE: GraphQL`, variables are written as a JSON object after the query, and the operation name is sent when the document names exactly one.
- The response pane lists the `errors` of a GraphQL response and marks the run as failed, which a `200` status alone would hide.

### Changed
- `.http` files are read by a hand written scanner instead of Tree-sitter. The build is now pure Go: `go install` and cross-compilation need no C compiler, and the release archives for every platform are built on one runner.
- A GraphQL request previously went out as `application/graphql` with the raw query as its body, which most servers reject. Declare `Content-Type: application/graphql` to keep that form.
- The Suite pane starts the body on its own line, so that a formatted body is readable.

### Fixed
- A body type declared through `Content-Type` is now honoured, so a single-line JSON body is recognized and `application/graphql` is no longer read as JSON.
- A malformed variable declaration is reported as one instead of discarding the request that follows it.
- A request name or body containing brackets, such as `["a"]`, is no longer swallowed by the Suites list, which read list item text as style tags.

## [v0.11.0] - 2026-08-25
### Fixed
- Repeated headers such as `Cookie` are kept instead of discarding every header of the request and sending the first one as the body.
- A header value containing `*`, for example `Accept: */*`, no longer swallows the headers that follow it.
- Content the grammar cannot read is no longer appended to the request body, and requests folded into it are recovered instead of disappearing from the list.
- A `###` separator or a naming comment that follows an inline body is no longer sent as part of that body.
- A `Host` header declared by a request is now sent instead of being dropped by `net/http`.
- Secret values are now redacted from Hurl output, errors, and history, which previously kept them in clear text.
- Persisted history no longer stalls the interface after every request and no longer grows without bound: bodies are limited to 64 KiB per entry in the file and the write happens in the background. The response pane keeps the full body.

### Added
- A request body can be read from a file with `< ./payload.json`, resolved next to the `.http` file and limited to 10 MiB.
- The selected environment is passed to Hurl, so `.hurl` files resolve the same variables as `.http` files. The values go through a private file rather than the command line.
- Releases now carry prebuilt archives with checksums for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64, so installing no longer requires a Go toolchain and a C compiler.

### Changed
- Request headers are stored as `net/http.Header`, so a request can carry several values for the same name. Persisted history is written as version 2; version 1 files are still read.

## [v0.10.1] - 2026-08-25
### Fixed
- `Ctrl` keybindings now support Unicode characters for non-Latin keyboard layouts.

## [v0.10.0] - 2026-08-25
### Added
- Layered project configuration through `.lazyrest.yml` and an explicit `--config` file.
- Configuration CLI commands: `--generate-config`, `--print-config`, and `--validate-config`.
- Persistent, size-limited request history in `~/.config/lazyrest/history.json` with secret redaction and secure file permissions.
- Built-in Simplified Chinese (`zh`) localization.
- Built-in `gruvbox`, `catppuccin-mocha`, `tokyo-night`, `dracula`, `nord`, and `monokai` theme presets.

### Changed
- Keybinding validation now rejects conflicts within the same UI context while allowing reuse across independent panels.
- The command palette now provides an interactive picker for switching built-in themes at runtime.

## [v0.9.0] - 2026-08-25
### Added
- Configurable semantic UI colors under `theme` in `~/.config/lazyrest/config.yml`.
- Runtime configuration reload through `Ctrl+r`, a configurable `reload_config` action, or the command palette.
- A localized command palette for configuration reload, file reload, diagnostics, help, and quit.

### Changed
- Localization now covers byte progress, execution errors, palette commands, and language-aware diagnostic plural forms.

## [v0.8.0] - 2026-08-25
### Added
- Configurable UI localization through `language` and `languages` in `~/.config/lazyrest/config.yml`.
- Built-in English, Russian, and Spanish translations with English fallback and per-string overrides.

## [v0.7.0] - 2026-08-25
### Added
- Configurable multi-key bindings loaded from `~/.config/lazyrest/config.yml`, with defaults preserved for actions omitted from the file.
- The help overlay now displays the active key bindings.

## [v0.6.2] - 2026-08-25
### Changed
- Footer request progress is shown as a separate yellow powerline segment and disappears when execution finishes.
- Completed request status uses a dedicated green or orange powerline segment while keeping the selected file breadcrumb visible.
- The redundant `Suite:` segment is no longer displayed in the footer.

## [v0.6.1] - 2026-08-24
### Changed
- Suite and footer progress indicators now use yellow while idle or running, green after successful requests, and orange after errors or unsuccessful responses.

## [v0.6.0] - 2026-08-24
### Added
- An animated footer progress bar for startup, parsing, and request execution, including byte or percentage progress when available.

### Changed
- Pressing `Ctrl+l` on a selected file now opens it before moving focus from Files to Suites.
- The active Suite in the footer is now a matching-color powerline segment with a leading arrow.

## [v0.5.1] - 2026-08-24
### Changed
- Producer now displays an animated progress bar while connecting and waiting for a response, then switches to byte or percentage progress while reading the body.

## [v0.5.0] - 2026-08-24
### Added
- Neovim and LazyVim terminal integration examples in the README.
- A unified application state model covering startup, file discovery, parsing, execution, selection, diagnostics, and overlays.
- Dedicated diagnostics (`d`) and keyboard help (`?`) windows.
- Integration tests that drive the complete TUI through a simulated terminal.
- Ready-to-run `.http` and `.hurl` examples under `example/`.

### Changed
- Environment loading and initial file discovery now run after the TUI is visible, without blocking startup.
- Parser warnings are summarized in the Suites title and displayed in the diagnostics window.

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
