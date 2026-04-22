#!/bin/bash
set -e

REPO="symtalha14/tapr"
BINARY_NAME="tapr"
INSTALL_DIR="/usr/local/bin"

# --- Detect OS ---
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [[ "$OS" != "linux" && "$OS" != "darwin" ]]; then
  echo "❌ Unsupported OS: $OS"
  echo "   Windows users: download the binary from https://github.com/$REPO/releases"
  exit 1
fi

# --- Detect Arch ---
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)           ARCH="amd64" ;;
  aarch64 | arm64)  ARCH="arm64" ;;
  *)
    echo "❌ Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# --- Fetch latest release tag ---
echo "🔍 Fetching latest release..."
LATEST=$(curl -sSf "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' \
  | cut -d '"' -f 4)

if [[ -z "$LATEST" ]]; then
  echo "❌ Could not determine latest release. Check your internet connection."
  exit 1
fi

# --- Build download URL ---
BINARY="${BINARY_NAME}-${OS}-${ARCH}"
URL="https://github.com/$REPO/releases/download/$LATEST/$BINARY"

# --- Download ---
echo "⬇️  Downloading $BINARY_NAME $LATEST for $OS/$ARCH..."
TMP=$(mktemp)
if ! curl -sSfL "$URL" -o "$TMP"; then
  echo "❌ Download failed. No binary found for $OS/$ARCH in release $LATEST."
  echo "   Check available releases at: https://github.com/$REPO/releases"
  rm -f "$TMP"
  exit 1
fi

# --- Install ---
chmod +x "$TMP"
echo "🔐 Installing to $INSTALL_DIR/$BINARY_NAME (may prompt for password)..."
sudo mv "$TMP" "$INSTALL_DIR/$BINARY_NAME"

echo ""
echo "✅ tapr $LATEST installed successfully!"
echo "   Run: tapr --help"