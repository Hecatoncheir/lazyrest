#!/usr/bin/env bash
# Regenerates the stream preview assets from preview.tape.
#
# Usage: scripts/preview.sh
#
# VHS records the frames and this script assembles them. VHS can encode a GIF
# on its own, but it drives ffmpeg with flags a current ffmpeg no longer
# accepts, and its encode step fails without reporting anything — it simply
# leaves the frames behind. Assembling them here keeps the pipeline working
# whichever ffmpeg is installed.

set -euo pipefail

frames="preview-frames"
still="site/assets/preview-stream.png"
animation="preview-stream.gif"

for tool in vhs ffmpeg go; do
  command -v "$tool" > /dev/null || { echo "$tool is not on PATH" >&2; exit 1; }
done

rm -rf "$frames"
vhs preview.tape

shopt -s nullglob
captured=("$frames"/frame-text-*.png)
if [[ ${#captured[@]} -eq 0 ]]; then
  echo "vhs recorded no frames; the tape did not reach the interface" >&2
  exit 1
fi
echo "recorded ${#captured[@]} frames"

# The tape ends on the state worth keeping, so the last frame is the still.
# Indexed from the front, because the bash macOS ships has no negative index.
cp "${captured[${#captured[@]} - 1]}" "$still"

ffmpeg -y -loglevel error \
  -framerate 24 -i "$frames/frame-text-%05d.png" \
  -vf "fps=12,split[a][b];[a]palettegen=stats_mode=diff[p];[b][p]paletteuse=dither=bayer" \
  "$animation"

rm -rf "$frames"
echo "wrote $still and $animation"
