#!/bin/sh
set -e

# Install script for West
# Usage: 
#   curl -fsSL https://raw.githubusercontent.com/sprisa/west/main/install.sh | sh
#   curl -fsSL https://raw.githubusercontent.com/sprisa/west/main/install.sh | sh -s -- --version v1.2.3
#   curl -fsSL https://raw.githubusercontent.com/sprisa/west/main/install.sh | sh -s -- --dir /opt/bin

REPO="sprisa/west"
INSTALL_DIR="/usr/local/bin"
VERSION=""

# Parse flags
while [ $# -gt 0 ]; do
  case "$1" in
    --version)
      VERSION="$2"
      shift 2
      ;;
    --dir)
      INSTALL_DIR="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: sh install.sh [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --version VERSION    Install specific version (e.g., v1.2.3)"
      echo "  --dir DIRECTORY      Installation directory (default: /usr/local/bin)"
      echo "  -h, --help           Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Detect OS and architecture
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  OS="Linux" ;;
  Darwin) OS="Darwin" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get version (use --version flag if set, otherwise get latest)
if [ -z "$VERSION" ]; then
  LATEST_URL="https://api.github.com/repos/$REPO/releases/latest"
  VERSION=$(curl -fsSL "$LATEST_URL" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
  
  if [ -z "$VERSION" ]; then
    echo "Failed to get latest version"
    exit 1
  fi
fi

echo "Installing West ${VERSION}..."

# Download URL
BINARY="west"
ARCHIVE="west_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/${VERSION}/${ARCHIVE}"

# Create temp directory
TMP_DIR="$(mktemp -d)"
trap "rm -rf '$TMP_DIR'" EXIT

# Download and extract
echo "Downloading from $DOWNLOAD_URL..."
curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$ARCHIVE"

echo "Extracting..."
tar -xzf "$TMP_DIR/$ARCHIVE" -C "$TMP_DIR"

# Install
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
  chmod +x "$INSTALL_DIR/$BINARY"
else
  sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/$BINARY"
  sudo chmod +x "$INSTALL_DIR/$BINARY"
fi

echo "✓ West ${VERSION} installed successfully!"
echo "Run 'west --version' to verify the installation"
