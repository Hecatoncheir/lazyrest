# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [v0.28.0] - 2026-09-14

### Added
- The frame composer remembers what has been sent. `Up` and `Down` walk back through earlier frames and return to the draft, so a subscription can be resent after a reconnect without retyping it. The history is bounded, drops consecutive repeats, keeps what was typed rather than what went on the wire, and is never written to disk.

### Changed
- The fuzz smoke run in CI lasts ten seconds per target rather than two. The deadline used to arrive while a target was still gathering baseline coverage, which failed the run with `context deadline exceeded` and blocked a release.
- The file watcher debounce window is a variable so its test can widen it. The test writes a burst and expects one report, which depends on the burst landing inside one window; the production window cannot guarantee that on a loaded machine.

## [v0.27.1] - 2026-09-14

### Fixed
- The frame composer leaves the screen once the frame is sent, instead of staying drawn over the log.
- A frame occupies exactly one row in the log. A payload ending in CRLF used to add a blank row, so the rows no longer matched the frames.
- A `.socket` body expands `\r`, `\n`, and `\t`. The parser trims the trailing newline off a body, so a file could not otherwise end a message the way a line oriented protocol requires; the bundled Redis example depended on it.

## [v0.27.0] - 2026-09-14

### Added
- WebSocket requests run from `.http` files, addressed by a `ws://` or `wss://` URL, and carry the headers and session cookies of the requests around them.
- Raw TCP requests run from a new `.socket` file. A socket carries no message boundary, so a burst is joined into one frame and a pause ends it.
- A live pane replaces the response while a stream is open: frames appear with direction, kind, size, and time; JSON is highlighted, unprintable payloads are shown as hex, and secrets are redacted. `f` follows or pauses, `c` clears without hanging up, and `Esc` closes the connection.
- The body of a stream request is sent as its opening message on connecting, which is where a subscription belongs.
- Frames are sent from the pane with `s`. A raw socket expands `\r`, `\n`, and `\t`, which a line oriented protocol needs and an input field cannot hold; a WebSocket message is sent exactly as typed, so JSON keeps its own escapes.
- The frame log is bounded by both count and size, and reports how many earlier frames it discarded rather than pretending to be complete.

### Fixed
- The footer no longer offers help twice when no pane holds focus.
- Focus moves into and out of the stream pane, instead of navigating to the response view that the stream had replaced.
- The stream pane follows a theme, language, or keybinding change like every other pane.
- Picking a history entry while a stream is open brings the response pane back and closes the connection, instead of drawing the entry off screen behind a live socket.

## [v0.26.1] - 2026-09-14

### Changed
- CI discovers its fuzz targets with `go test -list` instead of two hand-written lists, so a target added to the tree can no longer go unfuzzed, and a discovery that finds nothing fails the run instead of passing empty.
- Contributor documentation now records the full set of CI gates, how the terminal UI is wired, and how the terminal tests work; the everyday command list lives in `AGENTS.md` alone.

No application code changed since v0.26.0. The binaries differ only in the stamped version, so there is nothing to gain by upgrading.

## [v0.26.0] - 2026-09-13

### Added
- Request files are watched recursively and refreshed automatically after debounced writes, creates, renames, and removals.
- Hurl runs now render the selected exchange's HTTP status, headers, protocol, body, and assertion failures instead of exposing only the raw JSON report.
- Theme selections from the command palette are saved to the owning configuration layer and restored on the next launch.

### Fixed
- Truncated responses now keep the complete declared size separately from the bytes retained for display and history; chunked responses report a lower bound when their full size is unknowable.

## [v0.25.0] - 2026-09-13

### Changed
- Empty History and Captured responses windows now explain how to populate them and hide the unavailable clear action.
- Help and Diagnostics overlays now expose their toggle shortcuts in the contextual footer hints.
- Theme, Environment, and Save Response overlays now expose their primary Enter action in the contextual footer hints.
- Save Response now switches its footer hints to confirm/cancel when an existing file needs overwrite confirmation.
- Files and Suites search titles now show the current match and total, such as `[2/7]`.
- Added responsive screenshot regression coverage for 60, 80, 120, and 160-column terminals.
- Added a first-run footer hint for Help and the command palette until the first request starts.

## [v0.24.0] - 2026-09-13

### Changed
- Clearing persistent history or session captures now requires a second `c` confirmation; `Esc` cancels the pending destructive action.

## [v0.23.0] - 2026-09-13

### Added
- The command palette supports live filtering with `/`, match counts, and a guided no-results state; `Esc` clears an active filter before closing the palette.
- Theme and environment pickers mark the active choice explicitly instead of relying on cursor position alone.

## [v0.22.0] - 2026-09-13

### Added
- Producer can independently reveal response headers with `h` and resolved request details with `i`, while its title shows the current detail state.
- Focused panes now include a visible `▶` marker so keyboard focus does not rely on colour alone.

### Changed
- Help, history, diagnostics, pickers, and save dialogs now shrink to remain usable in small terminals.
- Built-in themes meet automated contrast targets for primary, muted, selected, status, breadcrumb, and focused-border colour pairs.

## [v0.21.0] - 2026-09-12

### Added
- A warm, responsive GitHub Pages landing page presents lazyrest's core workflow, installation command, and terminal interface, with a new application icon shared by the site and README.
- Context-aware footer hints expose the most useful keybindings for the focused pane, and empty searches now explain why no rows are visible.

### Changed
- The workspace adapts from three panes to two-pane and focused-stage layouts as the terminal narrows, while the request preview gives more room to the request list.
- Producer puts a compact response summary and body before response headers and request details, making the result faster to scan.

## [v0.20.0] - 2026-08-27
### Added
- Seeded fuzz targets cover HTTP document splitting, headers, variables, response references, secret redaction, Hurl entries, dotenv values, and POSIX shell quoting; CI runs short mutations on every change and a deeper weekly fuzz job.
- CI scans reachable Go dependencies with `govulncheck`, Dependabot monitors Go modules and GitHub Actions weekly, and every release archive includes a target-specific CycloneDX SBOM that is also published as a separate release asset.

### Changed
- Releases are now explicitly dispatched by version from `main`. The workflow validates the version and changelog, runs all checks, builds every archive, and only then publishes the tag and GitHub Release, cleaning up a partially published tag if release creation fails.
- The minimum Go toolchain is now 1.25.13 so released binaries include the latest standard-library security fixes from the Go 1.25 line.

## [v0.19.1] - 2026-08-27
### Fixed
- Clearing history now registers its background persistence snapshot before exposing the empty in-memory state, preventing `WaitForHistory` from racing with a concurrent write.

## [v0.19.0] - 2026-08-27
### Added
- Persistent history now defaults to metadata-only storage through `history: metadata`; `history: full` opts into restoring redacted request and response details.

### Changed
- Loading detailed history while metadata mode is active removes request and response details from memory and rewrites the project history without them.

### Fixed
- Persistent history now redacts GraphQL variables, GraphQL error messages, and other stored strings, and serializes an explicit allowlist of request and response fields so new runtime-only fields cannot leak to disk accidentally.
- Tokens, credentials, and session values created by responses are now treated as secrets when they come from sensitive response references, headers, or common JSON fields, keeping them out of rendering and persisted full history.
- Corrupted history files and failed background writes are now retained as deduplicated Diagnostics entries instead of being silently ignored.
- Parser diagnostics now distinguish warnings from blocking errors, and requests with missing external bodies, undefined variables, or cyclic variable references are rejected before any network call.

## [v0.18.0] - 2026-08-27
### Added
- File discovery honours `.gitignore` files from the project root and nested directories, including Git globs, directory-only rules, and negation; invalid patterns appear in diagnostics.

## [v0.17.0] - 2026-08-27
### Added
- A project-root `.env` file is loaded automatically as a private base environment, including for Hurl; `-dotenv-file` selects another filename, and JSON profiles plus request-local declarations retain their precedence.
- The command palette can open a separate project History window that lists safe run metadata newest first, opens an entry in Producer with `Enter` / `l`, and clears memory and persisted entries with `c`.
- Producer repeats the visible current-session request with `R`, including an entry selected from the History window.
- Producer copies the resolved request as a quoted, multiline cURL command with `C` or **Copy as cURL**; GraphQL uses the same encoded payload as the runner.
- **Choose environment** in the command palette switches between base `.env` values and JSON profiles, clears stale captures and cookies, and reparses the open request file without restarting.

### Changed
- Producer history state is synchronized across UI reads, active requests, and background persistence.
- Runtime theme regression coverage now verifies that existing content in Files and Suites is recolored immediately.

## [v0.16.0] - 2026-08-26
### Added
- Producer can save the complete plain-text response with `S`, matching what `Y` copies and respecting the active Pretty/Raw body mode.
- Producer response search supports cyclic `n` / `N` navigation and shows the current/total match count.
- Vim-style viewport navigation across all selectable and scrollable areas: `Ctrl+d` / `Ctrl+u` move by half a page, `Ctrl+f` / `Ctrl+b` by a full page, `gg` / `G` jump to the boundaries, and `zt` / `zz` / `zb` align the current item or scroll anchor.
- Persistent response history is isolated by canonical project root instead of mixing every project in one file.
- The command palette can open a captured-responses window that shows safe named-response summaries and clears session captures with `c`.

### Fixed
- `q` closes the active help, diagnostics, palette, theme, or save window instead of quitting lazyrest; from the main interface it still quits, while `Ctrl+C` remains an unconditional exit.
- Runtime theme changes now update the text background, border, and title state in filled Request and Result panes instead of leaving colours from the previous theme.

## [v0.15.0] - 2026-08-26
### Added
- Producer can copy the visible Pretty/Raw response body with `y`, copy the complete plain-text response with `Y`, or save the unformatted body with `s`. Export follows the selected history entry, keeps secret redaction, reports truncated data, and requires confirmation before overwriting a private response file.

## [v0.14.2] - 2026-08-26
### Fixed
- Captured responses are scoped to their request file, so two files that use the same request name can no longer resolve a reference with each other's answer.

## [v0.14.1] - 2026-08-26
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
