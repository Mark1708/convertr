#!/usr/bin/env bash
# Install convertr backend dependencies on macOS via Homebrew.
set -euo pipefail

if ! command -v brew &>/dev/null; then
  echo "Homebrew not found. Install from https://brew.sh" >&2
  exit 1
fi

echo "Installing convertr dependencies..."

brew install \
  pandoc \
  ffmpeg \
  imagemagick \
  jq \
  yq \
  tesseract \
  tesseract-lang \
  csvkit \
  asciidoctor \
  figlet

# LibreOffice is distributed as a cask.
if ! command -v soffice &>/dev/null; then
  brew install --cask libreoffice
fi

echo "All dependencies installed."
