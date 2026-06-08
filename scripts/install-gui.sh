#!/bin/sh
set -e

REPO="divyo-argha/nottaker"
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
        CLI_BINARY="nottaker-darwin-$ARCH"
        GUI_BINARY="nottaker-gui-darwin-$ARCH"
        ;;
    linux)   
        CLI_BINARY="nottaker-linux-$ARCH"
        GUI_BINARY="nottaker-gui-linux-$ARCH"
        ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

CLI_URL="https://github.com/$REPO/releases/download/v$VERSION/$CLI_BINARY"
GUI_URL="https://github.com/$REPO/releases/download/v$VERSION/$GUI_BINARY"

INSTALL_DIR="/usr/local/bin"

echo "Downloading nottaker CLI & GUI..."

# Download CLI
curl -sSfL "$CLI_URL" -o nottaker
chmod +x nottaker

# Download GUI
curl -sSfL "$GUI_URL" -o nottaker-gui
chmod +x nottaker-gui

echo "Installing to $INSTALL_DIR (you may be prompted for your password)..."
if [ -w "$INSTALL_DIR" ]; then
    mv nottaker "$INSTALL_DIR/nottaker"
    mv nottaker-gui "$INSTALL_DIR/nottaker-gui"
else
    sudo mv nottaker "$INSTALL_DIR/nottaker"
    sudo mv nottaker-gui "$INSTALL_DIR/nottaker-gui"
fi

echo "✓ Installation complete!"
echo "Run 'nottaker' for the terminal interface."
echo "Run 'nottaker-gui' for the desktop application."
