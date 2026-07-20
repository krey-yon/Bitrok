const script = `#!/bin/sh
set -eu

GITHUB_REPO="krey-yon/Bitrok"
INSTALL_DIR="\${BITROK_INSTALL_DIR:-\${HOME}/.local/bin}"
VERSION="\${BITROK_VERSION:-latest}"

case "$VERSION" in
  latest|v[0-9]*.[0-9]*.[0-9]*) ;;
  *) echo "bitrok: invalid version: $VERSION" >&2; exit 1 ;;
esac

OS="$(uname -s)"
case "$OS" in
  Darwin) OS="darwin" ;;
  Linux) OS="linux" ;;
  *) echo "bitrok: unsupported OS: $OS" >&2; exit 1 ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "bitrok: unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

ARCHIVE_NAME="bitrok_\${OS}_\${ARCH}.tar.gz"
if [ "$VERSION" = "latest" ]; then
  BASE_URL="https://github.com/\${GITHUB_REPO}/releases/latest/download"
else
  BASE_URL="https://github.com/\${GITHUB_REPO}/releases/download/\${VERSION}"
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
echo "bitrok: downloading $VERSION ($OS/$ARCH)..."
download() {
  destination="$1"
  url="$2"
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$destination" "$url"
  elif command -v wget >/dev/null 2>&1; then
    wget -q -O "$destination" "$url"
  else
    echo "bitrok: curl or wget is required" >&2
    exit 1
  fi
}
download "$TMP_DIR/bitrok.tar.gz" "$BASE_URL/$ARCHIVE_NAME"
download "$TMP_DIR/checksums.txt" "$BASE_URL/checksums.txt"
EXPECTED="$(awk -v archive="$ARCHIVE_NAME" '$2 == archive { print $1; exit }' "$TMP_DIR/checksums.txt")"
if ! printf '%s' "$EXPECTED" | grep -Eq '^[0-9a-fA-F]{64}$'; then
  echo "bitrok: release checksum for $ARCHIVE_NAME was not found" >&2
  exit 1
fi
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL="$(sha256sum "$TMP_DIR/bitrok.tar.gz" | awk '{ print $1 }')"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL="$(shasum -a 256 "$TMP_DIR/bitrok.tar.gz" | awk '{ print $1 }')"
else
  echo "bitrok: sha256sum or shasum is required" >&2
  exit 1
fi
if [ "$ACTUAL" != "$EXPECTED" ]; then
  echo "bitrok: checksum verification failed for $ARCHIVE_NAME" >&2
  exit 1
fi

tar -xzf "$TMP_DIR/bitrok.tar.gz" -C "$TMP_DIR"
if [ ! -f "$TMP_DIR/bitrok" ]; then
  echo "bitrok: release archive did not contain the bitrok binary" >&2
  exit 1
fi
mkdir -p "$INSTALL_DIR"
mv "$TMP_DIR/bitrok" "$INSTALL_DIR/bitrok"
chmod +x "$INSTALL_DIR/bitrok"
case ":\${PATH}:" in
  *":\${INSTALL_DIR}:"*) ;;
  *) echo "bitrok: add \${INSTALL_DIR} to PATH: export PATH=\"\${INSTALL_DIR}:$PATH\"" ;;
esac
echo "bitrok: installed to \${INSTALL_DIR}/bitrok"
echo "bitrok: run 'bitrok --help' to get started"
`;

export function GET() {
  return new Response(script, {
    headers: {
      "Cache-Control": "public, max-age=300",
      "Content-Disposition": 'inline; filename="install.sh"',
      "Content-Type": "text/x-shellscript; charset=utf-8",
    },
  });
}
