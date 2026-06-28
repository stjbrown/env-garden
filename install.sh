#!/usr/bin/env sh
# env-garden installer — downloads the latest `eg` release binary.
#   curl -fsSL https://raw.githubusercontent.com/stjbrown/env-garden/main/install.sh | sh
set -eu

REPO="stjbrown/env-garden"
BIN="eg"
PREFIX="${PREFIX:-$HOME/.local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac
case "$os" in
  darwin|linux) ;;
  *) echo "unsupported os: $os" >&2; exit 1 ;;
esac

tag="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name":' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')"
if [ -z "$tag" ]; then
  echo "could not determine latest release" >&2; exit 1
fi
ver="${tag#v}"

url="https://github.com/${REPO}/releases/download/${tag}/env-garden_${ver}_${os}_${arch}.tar.gz"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $BIN $tag ($os/$arch)…"
curl -fsSL "$url" | tar -xz -C "$tmp"

mkdir -p "$PREFIX"
install -m 0755 "$tmp/$BIN" "$PREFIX/$BIN"

echo "Installed $BIN to $PREFIX/$BIN"
echo
echo "Next steps:"
echo "  1. Ensure $PREFIX is on your PATH."
echo "  2. Add the shell integration:"
echo "       echo 'eval \"\$(eg init zsh)\"'  >> ~/.zshrc   # or bash"
echo "  3. Create a profile:  eg add claude-code bedrock"
