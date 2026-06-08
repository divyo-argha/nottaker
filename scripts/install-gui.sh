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
    darwin)  
        CLI_BINARY="octonote-darwin-$ARCH"
        GUI_BINARY="octonote-gui-darwin-$ARCH"
        ;;
    linux)   
        CLI_BINARY="octonote-linux-$ARCH"
        GUI_BINARY="octonote-gui-linux-$ARCH"
        ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

CLI_URL="https://github.com/$REPO/releases/download/v$VERSION/$CLI_BINARY"
GUI_URL="https://github.com/$REPO/releases/download/v$VERSION/$GUI_BINARY"

INSTALL_DIR="/usr/local/bin"

echo "Downloading octonote CLI & GUI..."

# Download CLI
curl -sSfL "$CLI_URL" -o octonote
chmod +x octonote

# Download GUI
curl -sSfL "$GUI_URL" -o octonote-gui
chmod +x octonote-gui

echo "Installing to $INSTALL_DIR (you may be prompted for your password)..."
if [ -w "$INSTALL_DIR" ]; then
    mv octonote "$INSTALL_DIR/octonote"
    mv octonote-gui "$INSTALL_DIR/octonote-gui"
else
    sudo mv octonote "$INSTALL_DIR/octonote"
    sudo mv octonote-gui "$INSTALL_DIR/octonote-gui"
fi

echo "✓ Installation complete!"
echo "Run 'octonote' for the terminal interface."
echo "Run 'octonote-gui' for the desktop application."
