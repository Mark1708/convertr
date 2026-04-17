#!/usr/bin/env bash
# Install convertr backend dependencies on Debian/Ubuntu.
set -euo pipefail

if ! command -v apt-get &>/dev/null; then
  echo "apt-get not found. This script targets Debian/Ubuntu." >&2
  exit 1
fi

echo "Installing convertr dependencies..."

sudo apt-get update -qq
sudo apt-get install -y \
  pandoc \
  ffmpeg \
  imagemagick \
  jq \
  tesseract-ocr \
  tesseract-ocr-rus \
  tesseract-ocr-eng \
  libreoffice \
  asciidoctor \
  figlet \
  python3-pip

# yq (Go version) — not in apt, install via GitHub releases.
if ! command -v yq &>/dev/null; then
  YQ_VER=$(curl -s https://api.github.com/repos/mikefarah/yq/releases/latest | grep '"tag_name"' | cut -d'"' -f4)
  curl -fsSL "https://github.com/mikefarah/yq/releases/download/${YQ_VER}/yq_linux_amd64" -o /usr/local/bin/yq
  chmod +x /usr/local/bin/yq
fi

# csvkit via pip.
pip3 install --quiet csvkit 2>/dev/null || true

echo "All dependencies installed."
