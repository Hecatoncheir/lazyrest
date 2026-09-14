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

## Driving a stream pane

Three transports reach the pane, and each needs something listening. Do not
point a run at a public broker or echo service: it then depends on the network
and on somebody else's uptime.

### A raw socket

Start the echo server bundled with this skill and use a `.socket` file.

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

### A WebSocket

The bundled WebSocket server uses the same library lazyrest does, so it needs
no module of its own.

```bash
go build -o /tmp/lazyrest-ws ./.claude/skills/run-lazyrest/ws-echo
/tmp/lazyrest-ws &            # listens on ws://127.0.0.1:7799
mkdir -p /tmp/lazyrest-requests
printf '@url = "ws://127.0.0.1:7799"\n\n# @name Watch an echo stream\nWEBSOCKET {{url}}\n\n{"hello":"lazyrest"}\n' \
  > /tmp/lazyrest-requests/live.http
```

Frames say `text` rather than `bytes` here, because a WebSocket declares what
its messages are:

```
→ text 20B {"hello":"lazyrest"}
← text 29B {"echo":{"hello":"lazyrest"}}
← text 10B {"tick":1}
```

### MQTT

mosquitto refuses anonymous clients by default, so even a throwaway broker
needs a configuration file.

```bash
printf 'listener 1883 127.0.0.1\nallow_anonymous true\n' > /tmp/mosquitto.conf
mosquitto -c /tmp/mosquitto.conf &
mkdir -p /tmp/lazyrest-requests
printf '@broker = "mqtt://127.0.0.1:1883"\n\n# @name Watch the sensors\nMQTT {{broker}}\nClient-Id: lazyrest-run\nSubscribe: sensors/+/temperature; qos=1\nTopic: commands/reboot\n\n{"announce":"watching"}\n' \
  > /tmp/lazyrest-requests/live.http
```

Drive both directions with the clients that ship with mosquitto — an outside
publisher and an outside subscriber prove more than an echo does. Start the
subscriber **before** running the request, or it misses the body, which is
published the moment the session connects.

```bash
mosquitto_sub -h 127.0.0.1 -t 'commands/#' -v &    # sees what the pane sends
# ... now launch lazyrest and run the request ...
mosquitto_pub -h 127.0.0.1 -t sensors/7/temperature -m '{"celsius":23.5}'
```

The pane reports the connection and each subscription as events of their own,
and puts the topic before the payload:

```
← event 0B packet=connack client-id=lazyrest-run
← event 0B packet=suback topic=sensors/+/temperature qos=1
→ text 23B topic=commands/reboot {"announce":"watching"}
← text 16B topic=sensors/7/temperature {"celsius":23.5}
```

A quality of service of zero is not shown: it is what a broker assumes anyway,
and the row has little space to spare.

## Typing into the composer

`tmux send-keys -l` delivers a whole string faster than the field takes focus,
so the opening characters land on the pane behind it and the frame goes out
truncated — silently, because what is left still looks like a frame. Waiting
for the composer label is not enough on its own: the frame is drawn before the
field is focused.

Type one character at a time, and read the field back before pressing Enter.

```bash
field() {
  tmux capture-pane -t lazyrest -p | grep "кадр:" | sed 's/.*\(кадр:.*\)/\1/'
}
type_slowly() {
  for (( i = 0; i < ${#1}; i++ )); do
    tmux send-keys -t lazyrest -l "${1:$i:1}"
    sleep 0.08
  done
}

tmux send-keys -t lazyrest s
wait_for "кадр:"          # the label; "frame:" when the UI is in English
sleep 0.4                 # the field takes focus after the frame is drawn
type_slowly "SUBSCRIBE prices"
tmux send-keys -t lazyrest Enter
```

`Up` and `Down` walk the frames sent earlier and return to an empty draft.
The opening message from the request body is not among them: the history holds
what was typed here, not everything the connection has sent.

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
pkill -f lazyrest-ws 2>/dev/null || true
pkill -f "mosquitto -c" 2>/dev/null || true
pkill -f mosquitto_sub 2>/dev/null || true
```

## Direct run, for a human

```bash
go run . example
```
