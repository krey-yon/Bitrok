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

case "$VERSION" in
	latest|v[0-9]*.[0-9]*.[0-9]*) ;;
	*) echo "bitrok: invalid version: $VERSION" >&2; exit 1 ;;
esac

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

ARCHIVE_NAME="bitrok_${OS}_${ARCH}.tar.gz"
if [ "$VERSION" = "latest" ]; then
	BASE_URL="https://github.com/${GITHUB_REPO}/releases/latest/download"
else
	BASE_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}"
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "bitrok: downloading ${VERSION} (${OS}/${ARCH})..."
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
