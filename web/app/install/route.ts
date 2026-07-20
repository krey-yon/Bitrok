const script = `#!/bin/sh
set -eu

GITHUB_REPO="krey-yon/Bitrok"
INSTALL_DIR="\${BITROK_INSTALL_DIR:-\${HOME}/.local/bin}"
VERSION="\${BITROK_VERSION:-latest}"

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

if [ "$VERSION" = "latest" ]; then
  URL="https://github.com/\${GITHUB_REPO}/releases/latest/download/bitrok_\${OS}_\${ARCH}.tar.gz"
else
  URL="https://github.com/\${GITHUB_REPO}/releases/download/\${VERSION}/bitrok_\${OS}_\${ARCH}.tar.gz"
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT
echo "bitrok: downloading $VERSION ($OS/$ARCH)..."
if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o "$TMP_DIR/bitrok.tar.gz" "$URL"
elif command -v wget >/dev/null 2>&1; then
  wget -q -O "$TMP_DIR/bitrok.tar.gz" "$URL"
else
  echo "bitrok: curl or wget is required" >&2
  exit 1
fi

tar -xzf "$TMP_DIR/bitrok.tar.gz" -C "$TMP_DIR"
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
