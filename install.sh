#!/bin/bash
set -e

# ding installer for macOS and Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/vishwaszadte/ding/main/install.sh | bash

REPO="vishwaszadte/ding"
BINARY="ding"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$OS" in
    linux*)  PLATFORM="linux" ;;
    darwin*) PLATFORM="darwin" ;;
    *)       echo "Unsupported operating system: $OS" && exit 1 ;;
esac

# Detect Architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *)             echo "Unsupported architecture: $ARCH" && exit 1 ;;
esac

echo "🔔 Installing ding for $PLATFORM/$ARCH..."

# Determine installation directory
INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

# Download latest release asset
LATEST_TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
if [ -z "$LATEST_TAG" ]; then
    LATEST_TAG="v1.0.0"
fi

DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST_TAG/ding_${PLATFORM}_${ARCH}.tar.gz"

TMP_DIR=$(mktemp -d)
curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/ding.tar.gz"
tar -xzf "$TMP_DIR/ding.tar.gz" -C "$TMP_DIR"
mv "$TMP_DIR/ding" "$INSTALL_DIR/ding"
chmod +x "$INSTALL_DIR/ding"
rm -rf "$TMP_DIR"

echo "✅ ding installed to $INSTALL_DIR/ding"

# Check if INSTALL_DIR is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "⚠️  Note: $INSTALL_DIR is not in your PATH."
    echo "Add it by running:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi

echo ""
echo "🎉 Run 'ding test' to verify your desktop notifications!"
