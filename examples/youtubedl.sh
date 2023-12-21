#!/bin/bash

set -e -x -o pipefail

# Example by Alex Ellis

# https://www.youtube.com/watch?v=tPEE9ZwTmy0

sudo curl -LSls -o /usr/local/bin/yt-dlp \
  https://github.com/yt-dlp/yt-dlp/releases/download/2023.11.16/yt-dlp_linux && \
  sudo chmod +x /usr/local/bin/yt-dlp

mkdir -p uploads

yt-dlp -o "uploads/video.flv" "https://www.youtube.com/watch?v=tPEE9ZwTmy0" && \
  du -h -d 0 uploads/video.flv
