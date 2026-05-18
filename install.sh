#!/bin/sh
# Vessel installer (Linux).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/hanifanggawi/vessel/main/install.sh | sh
#
# Downloads the latest release binary from GitHub, verifies its checksum, and
# installs it to /usr/local/bin. It does NOT enable the startup service; run
# `sudo vessel daemon install` afterwards for that.
set -eu

REPO="hanifanggawi/vessel"
BIN="vessel"
INSTALL_DIR="${VESSEL_INSTALL_DIR:-/usr/local/bin}"

err() { echo "error: $*" >&2; exit 1; }

# --- detect platform -------------------------------------------------------
os="$(uname -s)"
[ "$os" = "Linux" ] || err "this installer supports Linux only (detected: $os). On Windows, download the .zip from the releases page."

case "$(uname -m)" in
  x86_64|amd64)  arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) err "unsupported architecture: $(uname -m)" ;;
esac

# --- pick a downloader -----------------------------------------------------
if command -v curl >/dev/null 2>&1; then
  dl() { curl -fsSL "$1" -o "$2"; }
  fetch() { curl -fsSL "$1"; }
elif command -v wget >/dev/null 2>&1; then
  dl() { wget -qO "$2" "$1"; }
  fetch() { wget -qO - "$1"; }
else
  err "need curl or wget to download"
fi

# --- resolve latest version ------------------------------------------------
echo "Resolving latest release..."
tag="$(fetch "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name" *: *"([^"]+)".*/\1/')"
[ -n "$tag" ] || err "could not determine latest release tag"
version="${tag#v}"

archive="${BIN}_${version}_linux_${arch}.tar.gz"
base="https://github.com/${REPO}/releases/download/${tag}"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# --- download + verify -----------------------------------------------------
echo "Downloading ${archive}..."
dl "${base}/${archive}" "${tmp}/${archive}"
dl "${base}/checksums.txt" "${tmp}/checksums.txt"

echo "Verifying checksum..."
if command -v sha256sum >/dev/null 2>&1; then
  expected="$(grep " ${archive}\$" "${tmp}/checksums.txt" | awk '{print $1}')"
  actual="$(sha256sum "${tmp}/${archive}" | awk '{print $1}')"
else
  expected="$(grep " ${archive}\$" "${tmp}/checksums.txt" | awk '{print $1}')"
  actual="$(shasum -a 256 "${tmp}/${archive}" | awk '{print $1}')"
fi
[ -n "$expected" ] || err "no checksum found for ${archive}"
[ "$expected" = "$actual" ] || err "checksum mismatch for ${archive}"

# --- install ---------------------------------------------------------------
tar -xzf "${tmp}/${archive}" -C "${tmp}"
[ -f "${tmp}/${BIN}" ] || err "binary not found in archive"
chmod +x "${tmp}/${BIN}"

if [ -w "$INSTALL_DIR" ]; then
  mv "${tmp}/${BIN}" "${INSTALL_DIR}/${BIN}"
else
  echo "Installing to ${INSTALL_DIR} (requires sudo)..."
  sudo mv "${tmp}/${BIN}" "${INSTALL_DIR}/${BIN}"
fi

echo
echo "vessel ${version} installed to ${INSTALL_DIR}/${BIN}"
echo
echo "Next steps:"
echo "  vessel init                  # create your per-user config"
echo "  sudo vessel daemon install   # run the reconcile daemon at startup"
