#!/bin/sh
set -eu

SNAPBACK_BASE_URL_SET="${SNAPBACK_BASE_URL:+1}"
SNAPBACK_BASE_URL="${SNAPBACK_BASE_URL:-https://github.com/adeelahmad/snapback/releases/latest/download}"

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

os_release_value() {
	sed -n "s/^$1=//p" "$2" | tr -d "\"'" | tr '[:upper:]' '[:lower:]'
}

# fuse_command prints the command that installs fuse3 on this distribution, or
# fails when the distribution is unknown, so we never name a command that would
# not work there.
fuse_command() {
	release="${SNAPBACK_OS_RELEASE:-/etc/os-release}"
	[ -r "$release" ] || return 1
	for name in $(os_release_value ID "$release") $(os_release_value ID_LIKE "$release"); do
		case "$name" in
			debian | ubuntu | raspbian)
				echo 'sudo apt install fuse3'
				return 0
				;;
			fedora | rhel | centos | rocky | almalinux | alma)
				echo 'sudo dnf install fuse3'
				return 0
				;;
			arch | manjaro)
				echo 'sudo pacman -S fuse3'
				return 0
				;;
			alpine)
				echo 'sudo apk add fuse3'
				return 0
				;;
			opensuse* | sles | sled | suse)
				echo 'sudo zypper install fuse3'
				return 0
				;;
		esac
	done
	return 1
}

next_steps() {
	printf '\nNext steps:\n'
	case "$1" in
		darwin) printf '  1. Install macFUSE: https://macfuse.github.io/\n' ;;
		linux)
			if fuse_cmd=$(fuse_command); then
				printf '  1. Install fuse3: %s\n' "$fuse_cmd"
			else
				printf '  1. Install the fuse3 package with your distribution package manager\n'
			fi
			;;
	esac
	printf '  2. Run: snapback config to create your configuration\n'
	case "$1" in
		darwin)
			printf '  3. Run: snapback run to start the daemon in the foreground\n'
			printf '     (starting Snapback at login is Linux-only for now; it needs systemd)\n'
			;;
		linux) printf '  3. Run: snapback install service to start Snapback at login\n' ;;
	esac
	printf '  4. Run: snapback doctor to check your setup\n'
	printf '  5. Run: snapback version to confirm the install\n'
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

path_hint() {
	if on_path "$1"; then
		return 0
	fi
	printf 'warning: %s is not on your PATH; add it to your PATH:\n' "$1" >&2
	shell="${SNAPBACK_SHELL:-${SHELL:-}}"
	case "${shell##*/}" in
		bash) printf '  export PATH="%s:%s"  # add this line to ~/.bashrc\n' "$1" "\$PATH" >&2 ;;
		zsh) printf '  export PATH="%s:%s"  # add this line to ~/.zshrc\n' "$1" "\$PATH" >&2 ;;
		fish) printf '  fish_add_path %s\n' "$1" >&2 ;;
		*) printf '  export PATH="%s:%s"\n' "$1" "\$PATH" >&2 ;;
	esac
}

usage() {
	printf '%s\n' \
		'Usage: install.sh [options]' \
		'' \
		'Downloads the release binary for this OS and architecture,' \
		'verifies its checksum and copies it into an install directory.' \
		'' \
		'Options:' \
		'  --dry-run       print the download and install plan, then exit' \
		'  --dir DIR       install into DIR instead of the default location' \
		'  --version VER   install release tag VER instead of the latest release' \
		'  --yes           assume yes; reserved for a later confirmation step' \
		'  -h, --help      print this help and exit' \
		'' \
		'Environment:' \
		'  SNAPBACK_INSTALL_DIR   same as --dir' \
		'  SNAPBACK_DRY_RUN=1     same as --dry-run' \
		'  SNAPBACK_BASE_URL      base URL to download the release assets from' \
		'  SNAPBACK_OS            override the detected OS' \
		'  SNAPBACK_ARCH          override the detected architecture' \
		'  SNAPBACK_OS_RELEASE    read distribution facts from this file instead of /etc/os-release'
}

parse_args() {
	version=
	while [ "$#" -gt 0 ]; do
		case "$1" in
			-h | --help)
				usage
				exit 0
				;;
			--dry-run) SNAPBACK_DRY_RUN=1 ;;
			--dir)
				[ "$#" -ge 2 ] || die "--dir needs a directory"
				SNAPBACK_INSTALL_DIR="$2"
				shift
				;;
			--version)
				[ "$#" -ge 2 ] || die "--version needs a release tag"
				version="$2"
				shift
				;;
			--yes) ;;
			*)
				printf 'unknown argument: %s\n' "$1" >&2
				usage >&2
				exit 2
				;;
		esac
		shift
	done
	if [ -n "$version" ] && [ -z "${SNAPBACK_BASE_URL_SET:-}" ]; then
		SNAPBACK_BASE_URL="https://github.com/adeelahmad/snapback/releases/download/$version"
	fi
}

main() {
	parse_args "$@"

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
	if [ "$os" = darwin ]; then
		printf 'warning: the macOS binary is a self-contained executable; its linkage and runtime are not verified, and macOS support is a follow-up\n' >&2
	fi

	install_dir="${SNAPBACK_INSTALL_DIR:-/usr/local/bin}"
	if [ -z "${SNAPBACK_INSTALL_DIR:-}" ] && { [ ! -w "$install_dir" ] || ! on_path "$install_dir"; }; then
		install_dir="$HOME/.local/bin"
	fi
	printf 'installing snapback %s for %s/%s to %s\n' "${version:-latest}" "$os" "$arch" "$install_dir"
	printf 'set SNAPBACK_INSTALL_DIR or pass --dir to change the location\n'

	if [ "${SNAPBACK_DRY_RUN:-0}" = 1 ]; then
		printf 'asset: %s\n' "$asset"
		printf 'url: %s/%s\n' "$SNAPBACK_BASE_URL" "$asset"
		printf 'checksums: %s/checksums.txt\n' "$SNAPBACK_BASE_URL"
		printf 'install dir: %s\n' "$install_dir"
		next_steps "$os"
		path_hint "$install_dir"
		exit 0
	fi

	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT
	download "$SNAPBACK_BASE_URL/$asset" "$tmp/$asset" || die "failed to download $SNAPBACK_BASE_URL/$asset"
	download "$SNAPBACK_BASE_URL/checksums.txt" "$tmp/checksums.txt" || die "failed to download checksums.txt"
	want=$(awk -v a="$asset" '$2 == a { print $1 }' "$tmp/checksums.txt")
	[ -n "$want" ] || die "no checksum for $asset in checksums.txt"
	[ "$(sha256 "$tmp/$asset")" = "$want" ] || die "checksum mismatch for $asset; refusing to install"
	printf 'checksum verified: %s\n' "$asset"
	tar -xzf "$tmp/$asset" -C "$tmp" snapback
	mkdir -p "$install_dir"
	cp "$tmp/snapback" "$install_dir/snapback"
	chmod 755 "$install_dir/snapback"
	printf 'installed snapback to %s\n' "$install_dir/snapback"
	next_steps "$os"
	path_hint "$install_dir"
}

main "$@"
