#!/bin/bash

set -e -x -o pipefail

# Example by Alex Ellis

# AT&T Archives: The UNIX Operating System
TARGET=https://www.youtube.com/watch?v=tc4ROCJYbm0

export DEBIAN_FRONTEND=noninteractive

# Add ffmpeg to convert to mp4 later
sudo -E apt update -qqqy && \
  time sudo -E apt install -qqqy ffmpeg \
      --no-install-recommends

DL_URL=https://github.com/yt-dlp/yt-dlp/releases/download/2023.11.16/yt-dlp_linux

if [ "$(uname -m)" == "aarch64" ]; then
  DL_URL="https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux_aarch64"
fi

sudo curl -LSls -o /usr/local/bin/yt-dlp \
  $DL_URL && \
  sudo chmod +x /usr/local/bin/yt-dlp

mkdir -p videos
mkdir -p uploads

yt-dlp -o "./videos/video.flv" "$TARGET" && \
  du -h -d 0 ./videos/*

# Convert to mp4
cd videos
for i in *; do time ffmpeg -i "$i" ../uploads/"$i".mp4; done

