#!/bin/sh
set -e

REPO="divyo-argha/octonote"
VERSION="1.0.0"

# Detect OS and Arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin)  BINARY="octonote-darwin-$ARCH" ;;
    linux)   BINARY="octonote-linux-$ARCH" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

URL="https://github.com/$REPO/releases/download/v$VERSION/$BINARY"
INSTALL_DIR="/usr/local/bin"

echo "Downloading octonote CLI from $URL..."
curl -sSfL "$URL" -o octonote
chmod +x octonote

if [ -w "$INSTALL_DIR" ]; then
    mv octonote "$INSTALL_DIR/octonote"
else
    echo "Requires sudo permission to install to $INSTALL_DIR:"
    sudo mv octonote "$INSTALL_DIR/octonote"
fi

echo "octonote CLI successfully installed to $INSTALL_DIR/octonote"
