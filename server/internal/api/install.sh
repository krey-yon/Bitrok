#!/bin/sh
# bitrok CLI installer — https://bitrok.tech
# Usage: curl -fsSL https://bitrok.tech/install | sh
#   Pin a version: curl -fsSL https://bitrok.tech/install | sh -s -- --version v1.0.0
set -eu

GITHUB_REPO="krey-yon/Bitrok"
BINARY_NAME="bitrok"
INSTALL_DIR="${BITROK_INSTALL_DIR:-${HOME}/.local/bin}"

VERSION="latest"
while [ $# -gt 0 ]; do
	case "$1" in
		--version) VERSION="$2"; shift 2 ;;
		--version=*) VERSION="${1#*=}"; shift ;;
		*) shift ;;
	esac
done

OS="$(uname -s)"
case "$OS" in
	Darwin) OS="darwin" ;;
	Linux)  OS="linux" ;;
	*) echo "bitrok: unsupported OS: $OS" >&2; exit 1 ;;
esac

ARCH="$(uname -m)"
case "$ARCH" in
	x86_64|amd64)   ARCH="amd64" ;;
	aarch64|arm64)  ARCH="arm64" ;;
	*) echo "bitrok: unsupported arch: $ARCH" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
	URL="https://github.com/${GITHUB_REPO}/releases/latest/download/bitrok_${OS}_${ARCH}.tar.gz"
else
	URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/bitrok_${OS}_${ARCH}.tar.gz"
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "bitrok: downloading ${VERSION} (${OS}/${ARCH})..."
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
mv "$TMP_DIR/bitrok" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

case ":${PATH}:" in
	*":${INSTALL_DIR}:"*) ;;
	*)
		echo ""
		echo "bitrok: installed to ${INSTALL_DIR}/${BINARY_NAME}"
		echo "bitrok: add ${INSTALL_DIR} to your PATH:"
		printf '  export PATH="%s:$PATH"\n' "$INSTALL_DIR"
		;;
esac

echo ""
echo "bitrok: run '${BINARY_NAME} --help' to get started"
