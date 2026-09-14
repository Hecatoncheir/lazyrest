# preview.tape — regenerates the stream preview assets.
#
#   vhs preview.tape
#
# Run it from the repository root. It builds lazyrest and the echo server the
# run skill bundles, drives the TUI against a socket request, and captures the
# pane while frames arrive. Everything before the TUI is hidden, so only the
# terminal a reader cares about is recorded.
#
# The waits are on what is actually on screen rather than on fixed pauses, so a
# slow machine makes the recording longer instead of wrong.

Require go

# Only frames. VHS can encode a GIF itself, but the ffmpeg it drives is older
# than the one installed here: its encode step fails without saying so and
# leaves nothing behind. scripts/preview.sh assembles these frames instead.
Output preview-frames/

Set Shell fish

# The terminal renders what lazyrest paints, and the pane draws its dimmer
# columns — the time and the size of a frame — in grey. On a background lighter
# than the theme's they disappear, so the background is set to match gruvbox,
# which is what lazyrest starts with.
Set Theme { "background": "#282828", "foreground": "#ebdbb2" }
Set FontSize 16
Set Width 1400
Set Height 760
Set Padding 24
Set TypingSpeed 55ms

Hide
Type "mkdir -p /tmp/lazyrest-vhs/bin /tmp/lazyrest-vhs/requests" Enter
Type "go build -o /tmp/lazyrest-vhs/bin/lazyrest . && go build -o /tmp/lazyrest-vhs/bin/echo-server ./.claude/skills/run-lazyrest/echo-server && echo BUILT" Enter
Wait+Screen@180s /BUILT/
# set -gx is fish's own; export survives only as a compatibility function
# whose presence differs between fish versions, and CI carries a different one.
Type "set -gx PATH /tmp/lazyrest-vhs/bin $PATH" Enter
Type "echo-server > /dev/null 2>&1 &" Enter
Sleep 1s
Type "printf '# @name Live cache stream\nSOCKET tcp://127.0.0.1:7788\n\nPING\n' > /tmp/lazyrest-vhs/requests/live.socket && echo READY" Enter
Wait+Screen@20s /READY/
Type "clear" Enter
Sleep 500ms
Show

Type "lazyrest /tmp/lazyrest-vhs/requests" Enter
Wait+Screen@30s /live.socket/
Sleep 1s

Type "j"
Sleep 700ms
Enter
Wait+Screen@20s /Live cache stream/
Sleep 900ms

Enter
Sleep 900ms
Enter
Wait+Screen@20s /PONG/
Sleep 4s

# The still is the last frame, so the tape ends on the picture worth keeping.
Sleep 2s

Hide
Escape
Type "q" Enter
Type "pkill -f echo-server; rm -rf /tmp/lazyrest-vhs" Enter
Sleep 1s
Show
