#!/bin/sh
set -e

REPO="nottaker/nottaker"
VERSION="0.1.0"

# Detect OS and Arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin)  BINARY="nottaker-darwin-$ARCH" ;;
    linux)   BINARY="nottaker-linux-$ARCH" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

URL="https://github.com/$REPO/releases/download/v$VERSION/$BINARY"
INSTALL_DIR="/usr/local/bin"

echo "Downloading nottaker CLI from $URL..."
curl -sSfL "$URL" -o nottaker
chmod +x nottaker

if [ -w "$INSTALL_DIR" ]; then
    mv nottaker "$INSTALL_DIR/nottaker"
else
    echo "Requires sudo permission to install to $INSTALL_DIR:"
    sudo mv nottaker "$INSTALL_DIR/nottaker"
fi

echo "nottaker CLI successfully installed to $INSTALL_DIR/nottaker"
