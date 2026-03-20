#!/bin/sh
set -e

REPO="smilemilks2021/easy-web"
INSTALL_DIR="/usr/local/bin"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case $ARCH in
  x86_64)         ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

# Use grep -oE for macOS/Linux portable version extraction
VERSION=$(curl -sSf "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' \
  | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' \
  | head -1)

if [ -z "$VERSION" ]; then
  echo "ERROR: Could not determine latest version"; exit 1
fi

echo "Installing easy-web v${VERSION} (${OS}/${ARCH})..."
URL="https://github.com/${REPO}/releases/download/v${VERSION}/easy-web_${VERSION}_${OS}_${ARCH}.tar.gz"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
curl -sSfL "$URL" -o "$TMP/easy-web.tar.gz"
tar -xz -C "$TMP" -f "$TMP/easy-web.tar.gz"
[ -f "$TMP/easy-web" ] || { echo "ERROR: binary not found in archive"; exit 1; }

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/easy-web" "${INSTALL_DIR}/easy-web"
  chmod +x "${INSTALL_DIR}/easy-web"
else
  sudo mv "$TMP/easy-web" "${INSTALL_DIR}/easy-web"
  sudo chmod +x "${INSTALL_DIR}/easy-web"
fi
echo "easy-web ${VERSION} installed. Run: easy-web --help"
