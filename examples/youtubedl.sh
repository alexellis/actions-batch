#!/bin/bash

set -e -x -o pipefail

# Example by Alex Ellis

# https://www.youtube.com/watch?v=zcq1LQq08lk

sudo curl -LSls -o /usr/local/bin/yt-dlp \
  https://github.com/yt-dlp/yt-dlp/releases/download/2023.11.16/yt-dlp_linux && \
  sudo chmod +x /usr/local/bin/yt-dlp

yt-dlp -o "video.flv" "https://www.youtube.com/watch?v=zcq1LQq08lk" && \
    du -h -d 0 video.flv
