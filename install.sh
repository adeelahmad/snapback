#!/bin/sh
set -eu

SNAPBACK_BASE_URL="${SNAPBACK_BASE_URL:-https://example.invalid/snapback/releases/latest/download}"

die() {
	printf 'error: %s\n' "$1" >&2
	exit 1
}

detect_os() {
	os="${SNAPBACK_OS:-$(uname -s)}"
	case "$os" in
		Linux) echo linux ;;
		Darwin) echo darwin ;;
		*) die "unsupported OS: $os" ;;
	esac
}

detect_arch() {
	arch="${SNAPBACK_ARCH:-$(uname -m)}"
	case "$arch" in
		x86_64 | amd64) echo amd64 ;;
		aarch64 | arm64) echo arm64 ;;
		armv6l | armv7l) echo arm ;;
		mips) echo mips ;;
		mipsel) echo mipsle ;;
		*) die "unsupported architecture: $arch" ;;
	esac
}

asset_name() {
	case "$2" in
		arm | mips | mipsle) suffix=_unverified ;;
		*) suffix= ;;
	esac
	echo "snapback_$1_$2$suffix.tar.gz"
}

next_steps() {
	printf '\nNext steps:\n'
	case "$1" in
		darwin) printf '  1. Install macFUSE: https://macfuse.github.io/\n' ;;
		linux) printf '  1. Install the fuse3 package with your distribution package manager\n' ;;
	esac
	printf '  2. Run: snapback config\n'
}

download() {
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL -o "$2" "$1"
	elif command -v wget >/dev/null 2>&1; then
		wget -q -O "$2" "$1"
	else
		die "curl or wget is required"
	fi
}

sha256() {
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$1" | cut -d ' ' -f 1
	else
		shasum -a 256 "$1" | cut -d ' ' -f 1
	fi
}

on_path() {
	case ":$PATH:" in
		*":$1:"*) return 0 ;;
		*) return 1 ;;
	esac
}

main() {
	os=$(detect_os)
	arch=$(detect_arch)
	case "$os/$arch" in
		darwin/amd64 | darwin/arm64 | linux/*) ;;
		*) die "unsupported platform: $os/$arch" ;;
	esac
	asset=$(asset_name "$os" "$arch")
	case "$arch" in
		arm | mips | mipsle) printf 'warning: %s/%s is an unverified target\n' "$os" "$arch" >&2 ;;
	esac

	install_dir="${SNAPBACK_INSTALL_DIR:-/usr/local/bin}"
	if [ -z "${SNAPBACK_INSTALL_DIR:-}" ] && { [ ! -w "$install_dir" ] || ! on_path "$install_dir"; }; then
		install_dir="$HOME/.local/bin"
	fi
	if ! on_path "$install_dir"; then
		printf 'warning: %s is not on your PATH; add it to use snapback\n' "$install_dir" >&2
	fi

	if [ "${SNAPBACK_DRY_RUN:-0}" = 1 ]; then
		printf 'asset: %s\n' "$asset"
		printf 'url: %s/%s\n' "$SNAPBACK_BASE_URL" "$asset"
		printf 'checksums: %s/checksums.txt\n' "$SNAPBACK_BASE_URL"
		printf 'install dir: %s\n' "$install_dir"
		next_steps "$os"
		exit 0
	fi

	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT
	download "$SNAPBACK_BASE_URL/$asset" "$tmp/$asset" || die "failed to download $SNAPBACK_BASE_URL/$asset"
	download "$SNAPBACK_BASE_URL/checksums.txt" "$tmp/checksums.txt" || die "failed to download checksums.txt"
	want=$(awk -v a="$asset" '$2 == a { print $1 }' "$tmp/checksums.txt")
	[ -n "$want" ] || die "no checksum for $asset in checksums.txt"
	[ "$(sha256 "$tmp/$asset")" = "$want" ] || die "checksum mismatch for $asset; refusing to install"
	tar -xzf "$tmp/$asset" -C "$tmp" snapback
	mkdir -p "$install_dir"
	cp "$tmp/snapback" "$install_dir/snapback"
	chmod 755 "$install_dir/snapback"
	printf 'installed snapback to %s\n' "$install_dir/snapback"
	next_steps "$os"
}

main "$@"
