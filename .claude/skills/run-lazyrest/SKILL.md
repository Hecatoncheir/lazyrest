---
name: run-lazyrest
description: Launch lazyrest against a directory of request files and drive it — open a file, run a request, inspect the response or the live stream pane. Use when asked to run the app, screenshot it, or confirm a change works in the real TUI rather than only in tests.
---

# Running lazyrest

lazyrest is a `tview`/`tcell` TUI: it takes over the terminal, so it cannot be
driven by the bash tool directly. Run it inside tmux and read the screen with
`capture-pane`.

## Build first

The repository root is the module root.

```bash
go build -o /tmp/lazyrest .
```

## Launch and wait

There is **no `timeout` on macOS**, so the poll loop the generic tmux recipe
uses does not work here. Define this helper instead:

```bash
wait_for() {
  for i in $(seq 1 60); do
    tmux capture-pane -t lazyrest -p | grep -qF "$1" && return 0
    sleep 0.3
  done
  echo "НЕ ДОЖДАЛСЯ: $1" >&2
  return 1
}
```

```bash
tmux kill-session -t lazyrest 2>/dev/null
tmux new-session -d -s lazyrest -x 120 -y 24 "/tmp/lazyrest ./example"
wait_for "basic.http"
```

Ready means a file name from the directory appears in the Files pane.

## Drive it to a response

The tree cursor starts on the **directory root**, not on the first file, so the
first `j` is not optional.

```bash
tmux send-keys -t lazyrest j       # move off the root onto the first file
tmux send-keys -t lazyrest Enter   # open the file -> Requests pane fills
tmux send-keys -t lazyrest Enter   # select a request -> Request pane
tmux send-keys -t lazyrest Enter   # run it
tmux capture-pane -t lazyrest -p
```

Use `wait_for` between the steps rather than `sleep`; a request name from the
file is a good marker.

## Key reference

| Key | Action | Pane |
|---|---|---|
| `j` / `k` | move | any list |
| `Enter` | open / select / run | tree, requests, request |
| `Esc` | back; on a stream, close the connection | any |
| `ctrl+h` / `ctrl+l` | move focus left / right | any |
| `s` | send a frame | stream |
| `f` | follow or pause | stream |
| `c` | clear the log, keeping the connection | stream |
| `?` | help | any |
| `:` | command palette | any |

## Driving the stream pane

Do not point a test at an external WebSocket service: it makes the run depend
on the network and on somebody else's uptime. Start the echo server bundled
with this skill and use a `.socket` file.

```bash
go build -o /tmp/lazyrest-echo ./.claude/skills/run-lazyrest/echo-server
/tmp/lazyrest-echo &          # listens on 127.0.0.1:7788
mkdir -p /tmp/lazyrest-requests
printf '# @name Live cache stream\nSOCKET tcp://127.0.0.1:7788\n\nPING\\r\\n\n' \
  > /tmp/lazyrest-requests/live.socket
```

Then launch against `/tmp/lazyrest-requests` and run the request. The pane
shows both directions:

```
07:07:17.944 → bytes 6B  PING\r\n
07:07:17.995 ← bytes 7B  +PONG\r\n
07:07:18.695 ← bytes 26B {"tick":1,"at":"07:07:18"}
```

A `.socket` body needs its terminator written as `\r\n`: the parser trims the
trailing newline, so a file cannot end a message any other way.

## Things that will confuse you

- **The UI follows the configured language.** `capture-pane` may come back in
  Russian (`Файлы`, `Запросы`, `Поток — слежение`). Match on request names and
  payloads rather than on pane titles.
- **Below 80 columns the layout is compact** and shows only the focused pane.
  That is deliberate — check the narrow layout with `-x 60`.
- **A payload is drawn with its control characters visible** (`+PONG\r\n`), so
  one frame is always one row. Do not expect a real line break.

## Clean up

```bash
tmux send-keys -t lazyrest q
tmux kill-session -t lazyrest 2>/dev/null || true
pkill -f lazyrest-echo 2>/dev/null || true
```

## Direct run, for a human

```bash
go run . example
```
