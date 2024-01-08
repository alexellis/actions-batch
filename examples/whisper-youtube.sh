#!/bin/bash

set -e -x -o pipefail

# Example by Alex Ellis

# Download a video from YouTube, convert it to audio, then
# transcribe it using Whisper's tiny model. When CUDA and a GPU are
# available, it'll run significantly faster, otherwise it'll revert to
# CPU mode.

## Stage 1 - download the video

TARGET=https://www.youtube.com/watch?v=igv9LRPzZbE

# If running on cuda, set DEVICE="cuda"
DEVICE="cpu"

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

time yt-dlp -o "./videos/video.flv" "$TARGET" && \
  du -h -d 0 ./videos/*

mkdir -p audio

## Stage 2 - convert the video to mp3
cd videos
time ffmpeg -i * -b:a 128K -vn ../audio/track.mp3
cd ../

## Stage 3 Run Whisper

# Install openai-whisper

time pip install -U openai-whisper

# Download the tiny model

cat << EOF > ./download_models.py
#!/bin/python3
import sys, os
from whisper import _download, _MODELS

models = ["tiny.en"]

home = os.path.expanduser('~')

for model in models:
    _download(_MODELS[model], os.path.join(home, ".cache/whisper"), False)
EOF

chmod +x ./download_models.py
time ./download_models.py

# Transcribe using the tiny model from the cache
time whisper --device $DEVICE --language English ./audio/*.mp3 --model tiny > ./uploads/track.txt

