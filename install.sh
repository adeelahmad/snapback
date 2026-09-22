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
		linux) printf '  1. Install fuse3 with your package manager (e.g. apt install fuse3)\n' ;;
	esac
	printf '  2. Run: snapback config\n'
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

	if [ "${SNAPBACK_DRY_RUN:-0}" = 1 ]; then
		printf 'asset: %s\n' "$asset"
		printf 'url: %s/%s\n' "$SNAPBACK_BASE_URL" "$asset"
		printf 'checksums: %s/checksums.txt\n' "$SNAPBACK_BASE_URL"
		printf 'install dir: %s\n' "$install_dir"
		next_steps "$os"
		exit 0
	fi

	die "installation is not implemented yet"
}

main "$@"
