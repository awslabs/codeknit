#!/usr/bin/env bash
# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
#
# Ensure a codeknit binary is available and print its absolute path on the last
# line of stdout. Downloads the GitHub release for this OS/arch, verifies it
# against the release checksums.txt, and installs it when missing.
#
# Environment:
#   CODEKNIT_VERSION      Release to install, e.g. 0.6.0. Defaults to latest.
#   CODEKNIT_INSTALL_DIR  Install directory. Defaults to ~/.local/bin.
#   CODEKNIT_FORCE=1      Reinstall even when codeknit is already available.

set -euo pipefail

repo="awslabs/codeknit"
install_dir="${CODEKNIT_INSTALL_DIR:-$HOME/.local/bin}"

log() { echo "codeknit-install: $*" >&2; }
die() { log "error: $*"; exit 1; }

case "$(uname -s)" in
	Darwin) os=darwin ;;
	Linux) os=linux ;;
	MINGW* | MSYS* | CYGWIN*) os=windows ;;
	*) die "unsupported OS: $(uname -s)" ;;
esac
case "$(uname -m)" in
	x86_64 | amd64) arch=amd64 ;;
	arm64 | aarch64) arch=arm64 ;;
	*) die "unsupported architecture: $(uname -m)" ;;
esac
[[ "$os/$arch" == windows/arm64 ]] && die "no prebuilt binary for windows/arm64"

exe="codeknit"
[[ "$os" == windows ]] && exe="codeknit.exe"

if [[ "${CODEKNIT_FORCE:-0}" != 1 ]]; then
	if existing="$(command -v codeknit 2>/dev/null)"; then
		echo "$existing"
		exit 0
	fi
	if [[ -x "$install_dir/$exe" ]]; then
		echo "$install_dir/$exe"
		exit 0
	fi
fi

command -v curl >/dev/null 2>&1 || die "curl is required"

version="${CODEKNIT_VERSION:-}"
if [[ -z "$version" ]]; then
	# Follow the /releases/latest redirect instead of the API to avoid rate limits.
	latest_url="$(curl -fsSLI -o /dev/null -w '%{url_effective}' "https://github.com/$repo/releases/latest")" ||
		die "cannot resolve latest release"
	version="${latest_url##*/}"
fi
version="${version#v}"
[[ "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]] || die "invalid version: $version"

ext="tar.gz"
[[ "$os" == windows ]] && ext="zip"
archive="codeknit_${version}_${os}_${arch}.${ext}"
base_url="https://github.com/$repo/releases/download/v$version"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

log "downloading $archive"
curl -fsSL -o "$tmp/$archive" "$base_url/$archive" || die "download failed: $base_url/$archive"
curl -fsSL -o "$tmp/checksums.txt" "$base_url/checksums.txt" || die "download failed: checksums.txt"

expected="$(awk -v f="$archive" '$2 == f { print $1 }' "$tmp/checksums.txt")"
[[ -n "$expected" ]] || die "no checksum listed for $archive"
if command -v sha256sum >/dev/null 2>&1; then
	actual="$(sha256sum "$tmp/$archive" | awk '{ print $1 }')"
elif command -v shasum >/dev/null 2>&1; then
	actual="$(shasum -a 256 "$tmp/$archive" | awk '{ print $1 }')"
else
	die "sha256sum or shasum is required to verify the download"
fi
[[ "$actual" == "$expected" ]] || die "checksum mismatch for $archive"

mkdir -p "$tmp/x"
if [[ "$ext" == zip ]]; then
	unzip -q "$tmp/$archive" "$exe" -d "$tmp/x" || die "extract failed"
else
	tar -xzf "$tmp/$archive" -C "$tmp/x" "$exe" || die "extract failed"
fi

mkdir -p "$install_dir"
install -m 0755 "$tmp/x/$exe" "$install_dir/$exe"
log "installed codeknit $version to $install_dir/$exe"
case ":$PATH:" in
	*":$install_dir:"*) ;;
	*) log "note: $install_dir is not on PATH; call the printed path directly or add it to PATH" ;;
esac
echo "$install_dir/$exe"
