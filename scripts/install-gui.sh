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
    darwin)  
        CLI_BINARY="octonote-darwin-$ARCH"
        GUI_BINARY="octonote-gui-darwin"
        ;;
    linux)   
        CLI_BINARY="octonote-linux-$ARCH"
        GUI_BINARY="octonote-gui-linux"
        ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

install_desktop_entry() {
    if [ "$OS" = "linux" ]; then
        echo "→ Installing desktop entry and icon…"
        DESKTOP_DIR="/usr/share/applications"
        ICON_DIR="/usr/share/pixmaps"
        
        # Check permissions
        USE_SUDO=""
        if [ ! -w "$DESKTOP_DIR" ] || [ ! -w "$ICON_DIR" ]; then
            if [ "$(id -u)" -ne 0 ]; then
                # Fallback to local user directory
                DESKTOP_DIR="$HOME/.local/share/applications"
                ICON_DIR="$HOME/.local/share/pixmaps"
                mkdir -p "$DESKTOP_DIR" "$ICON_DIR"
            else
                USE_SUDO="sudo"
            fi
        fi

        # Get the icon
        if [ -f "$TMP_DIR/octonote/assets/logo.png" ]; then
            $USE_SUDO cp "$TMP_DIR/octonote/assets/logo.png" "$ICON_DIR/octonote.png"
        else
            # Download from GitHub if running from pre-built download path
            $USE_SUDO curl -sSfL "https://raw.githubusercontent.com/$REPO/main/assets/logo.png" -o "$ICON_DIR/octonote.png" || true
        fi

        # Create desktop file
        TEMP_DESKTOP="$TMP_DIR/octonote.desktop"
        cat <<EOF > "$TEMP_DESKTOP"
[Desktop Entry]
Name=octoNote
Comment=Lightweight multi-tab auto-saving scratchpad
Exec=octonote-gui
Icon=octonote
Type=Application
Terminal=false
Categories=Utility;TextEditor;Development;
EOF
        $USE_SUDO mv "$TEMP_DESKTOP" "$DESKTOP_DIR/octonote.desktop"
        
        # Update desktop database if available
        if command -v update-desktop-database >/dev/null 2>&1; then
            $USE_SUDO update-desktop-database "$DESKTOP_DIR" 2>/dev/null || true
        fi
    fi
}

# Try fetching latest version from Github releases
VERSION=$(curl -sSfL "https://api.github.com/repos/$REPO/releases/latest" | grep -Po '"tag_name":\s*"\K[^"]+' || true)
VERSION="${VERSION#v}"

# Attempt pre-built binary download if version was found
if [ -n "$VERSION" ]; then
    CLI_URL="https://github.com/$REPO/releases/download/v$VERSION/$CLI_BINARY"
    GUI_URL="https://github.com/$REPO/releases/download/v$VERSION/$GUI_BINARY"
    
    echo "→ Attempting to download pre-built CLI & GUI..."
    
    # Try downloading both
    DOWNLOADED=true
    if ! curl -sSfL "$CLI_URL" -o "$TMP_DIR/octonote"; then
        DOWNLOADED=false
    fi
    
    if [ "$DOWNLOADED" = "true" ]; then
        if curl -sSfL "$GUI_URL" -o "$TMP_DIR/octonote-gui"; then
            # Success! Install them
            echo "→ Installing pre-built CLI & GUI to $INSTALL_DIR…"
            chmod +x "$TMP_DIR/octonote" "$TMP_DIR/octonote-gui"
            if [ -w "$INSTALL_DIR" ]; then
                mv "$TMP_DIR/octonote" "$INSTALL_DIR/octonote"
                mv "$TMP_DIR/octonote-gui" "$INSTALL_DIR/octonote-gui"
            else
                sudo mv "$TMP_DIR/octonote" "$INSTALL_DIR/octonote"
                sudo mv "$TMP_DIR/octonote-gui" "$INSTALL_DIR/octonote-gui"
            fi
            
            # Install launcher
            install_desktop_entry
            
            echo ""
            echo "✓ Installation complete!"
            echo "  Run 'octonote'     → Terminal TUI"
            echo "  Run 'octonote-gui' → Desktop GUI"
            exit 0
        else
            # Also try downloading with -amd64 / -arm64 suffix just in case
            GUI_URL_SUFFIXED="https://github.com/$REPO/releases/download/v$VERSION/${GUI_BINARY}-${ARCH}"
            if curl -sSfL "$GUI_URL_SUFFIXED" -o "$TMP_DIR/octonote-gui"; then
                echo "→ Installing pre-built CLI & GUI to $INSTALL_DIR…"
                chmod +x "$TMP_DIR/octonote" "$TMP_DIR/octonote-gui"
                if [ -w "$INSTALL_DIR" ]; then
                    mv "$TMP_DIR/octonote" "$INSTALL_DIR/octonote"
                    mv "$TMP_DIR/octonote-gui" "$INSTALL_DIR/octonote-gui"
                else
                    sudo mv "$TMP_DIR/octonote" "$INSTALL_DIR/octonote"
                    sudo mv "$TMP_DIR/octonote-gui" "$INSTALL_DIR/octonote-gui"
                fi
                
                # Install launcher
                install_desktop_entry
                
                echo ""
                echo "✓ Installation complete!"
                echo "  Run 'octonote'     → Terminal TUI"
                echo "  Run 'octonote-gui' → Desktop GUI"
                exit 0
            fi
        fi
    fi
    echo "⚠ Pre-built binaries not found. Falling back to build from source."
fi

# Fallback: Build from source
# Check Go
if ! command -v go >/dev/null 2>&1; then
    echo "✗ Go is not installed. Install it from https://go.dev/dl/ and retry."
    exit 1
fi

# Ensure Wails is installed
if ! command -v wails >/dev/null 2>&1; then
    echo "→ Wails not found. Installing Wails v2…"
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    export PATH="$PATH:$(go env GOPATH)/bin:$HOME/go/bin"
fi

# Linux GUI deps check (webkit2gtk)
if [ "$(uname -s)" = "Linux" ]; then
    if ! pkg-config --exists webkit2gtk-4.0 2>/dev/null && \
       ! pkg-config --exists webkit2gtk-4.1 2>/dev/null; then
        echo "⚠  webkit2gtk not found. Attempting to install…"
        if command -v apt-get >/dev/null 2>&1; then
            sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.1-dev || sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev
        elif command -v dnf >/dev/null 2>&1; then
            sudo dnf install -y gtk3-devel webkit2gtk4.0-devel
        elif command -v pacman >/dev/null 2>&1; then
            sudo pacman -S --noconfirm webkit2gtk
        else
            echo "✗ Please install webkit2gtk manually for your distro and retry."
            exit 1
        fi
    fi
fi

if [ -d ".git" ] && git remote -v 2>/dev/null | grep -q "$REPO"; then
    echo "→ Copying local repository files…"
    rm -rf "$TMP_DIR/octonote"
    mkdir -p "$TMP_DIR/octonote"
    cp -R ./* "$TMP_DIR/octonote"/
else
    echo "→ Cloning octonote…"
    if ! git clone --depth=1 "$REPO_URL" "$TMP_DIR/octonote" >/dev/null 2>&1; then
        echo "→ HTTPS clone failed. Retrying with SSH…"
        git clone --depth=1 "git@github.com:$REPO.git" "$TMP_DIR/octonote" >/dev/null 2>&1
    fi
fi

cd "$TMP_DIR/octonote"

echo "→ Building CLI…"
go build -trimpath -ldflags "-s -w" -o octonote ./tui

echo "→ Building GUI (this may take a moment)…"
cd gui
WAILS_TAGS=""
if pkg-config --exists webkit2gtk-4.1 2>/dev/null && ! pkg-config --exists webkit2gtk-4.0 2>/dev/null; then
    WAILS_TAGS="-tags webkit2_41"
fi
wails build -clean -ldflags "-s -w" $WAILS_TAGS
cd ..

echo "→ Installing to $INSTALL_DIR…"
if [ -w "$INSTALL_DIR" ]; then
    mv octonote "$INSTALL_DIR/octonote"
    mv gui/build/bin/octonote-gui "$INSTALL_DIR/octonote-gui" 2>/dev/null || \
    mv gui/build/bin/octonote "$INSTALL_DIR/octonote-gui"
else
    sudo mv octonote "$INSTALL_DIR/octonote"
    sudo mv gui/build/bin/octonote-gui "$INSTALL_DIR/octonote-gui" 2>/dev/null || \
    sudo mv gui/build/bin/octonote "$INSTALL_DIR/octonote-gui"
fi

# Install launcher
install_desktop_entry

echo ""
echo "✓ Installation complete!"
echo "  Run 'octonote'     → Terminal TUI"
echo "  Run 'octonote-gui' → Desktop GUI"
