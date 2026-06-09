#!/bin/sh
set -e

REPO="divyo-argha/octonote"
REPO_URL="https://github.com/$REPO.git"
INSTALL_DIR="/usr/local/bin"
TMP_DIR="$(mktemp -d)"

cleanup() { rm -rf "$TMP_DIR"; }
trap cleanup EXIT

# Detect OS and Arch
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
    darwin)  CLI_BINARY="octonote-darwin-$ARCH" ;;
    linux)   CLI_BINARY="octonote-linux-$ARCH" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Try fetching latest version from Github releases
VERSION=$(curl -sSfL "https://api.github.com/repos/$REPO/releases/latest" | grep -Po '"tag_name":\s*"\K[^"]+' || true)
VERSION="${VERSION#v}"

# Attempt pre-built binary download if version was found
if [ -n "$VERSION" ]; then
    CLI_URL="https://github.com/$REPO/releases/download/v$VERSION/$CLI_BINARY"
    echo "→ Attempting to download pre-built CLI from $CLI_URL…"
    if curl -sSfL "$CLI_URL" -o "$TMP_DIR/octonote"; then
        echo "→ Installing pre-built CLI to $INSTALL_DIR…"
        chmod +x "$TMP_DIR/octonote"
        if [ -w "$INSTALL_DIR" ]; then
            mv "$TMP_DIR/octonote" "$INSTALL_DIR/octonote"
        else
            sudo mv "$TMP_DIR/octonote" "$INSTALL_DIR/octonote"
        fi
        echo "✓ octonote CLI installed. Run: octonote"
        exit 0
    fi
    echo "⚠ Pre-built binary not found. Falling back to build from source."
fi

# Fallback: Build from source
# Check Go
if ! command -v go >/dev/null 2>&1; then
    echo "✗ Go is not installed. Install it from https://go.dev/dl/ to build from source."
    exit 1
fi

if [ -d ".git" ] && git remote -v 2>/dev/null | grep -q "$REPO"; then
    echo "→ Copying local repository files…"
    cp -R . "$TMP_DIR/octonote"
else
    echo "→ Cloning octonote…"
    if ! git clone --depth=1 "$REPO_URL" "$TMP_DIR/octonote" >/dev/null 2>&1; then
        echo "→ HTTPS clone failed. Retrying with SSH…"
        git clone --depth=1 "git@github.com:$REPO.git" "$TMP_DIR/octonote" >/dev/null 2>&1
    fi
fi

echo "→ Building CLI…"
cd "$TMP_DIR/octonote"
go build -trimpath -ldflags "-s -w" -o octonote ./tui

echo "→ Installing to $INSTALL_DIR…"
if [ -w "$INSTALL_DIR" ]; then
    mv octonote "$INSTALL_DIR/octonote"
else
    sudo mv octonote "$INSTALL_DIR/octonote"
fi

echo "✓ octonote CLI installed. Run: octonote"
