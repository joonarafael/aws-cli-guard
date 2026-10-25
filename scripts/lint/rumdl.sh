#!/bin/sh
# POSIX compliant, ready to run on Linux or macOS

set -eu

# Downloads the rumdl release named by RUMDL_VERSION into .cache and execs it.
# Make targets pass the version. CI runs those targets, so Markdown lint does
# not depend on a global install. Linux and macOS, amd64 and arm64, match the
# hosts in ADR 0001.

: "${RUMDL_VERSION:?RUMDL_VERSION is required (rumdl release, without a leading v)}"

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT" || exit 1

os=$(uname -s)
arch=$(uname -m)

case "$os" in
Linux) platform=unknown-linux-gnu ;;
Darwin) platform=apple-darwin ;;
*)
	echo "rumdl: unsupported operating system: ${os}" >&2
	exit 1
	;;
esac

case "$arch" in
x86_64 | amd64) cpu=x86_64 ;;
aarch64 | arm64) cpu=aarch64 ;;
*)
	echo "rumdl: unsupported architecture: ${arch}" >&2
	exit 1
	;;
esac

target="${cpu}-${platform}"
dest_dir="${REPO_ROOT}/.cache/rumdl/${RUMDL_VERSION}/${target}"
dest="${dest_dir}/rumdl"

version_matches() {
	[ -x "$1" ] || return 1
	got=$("$1" --version 2>/dev/null || true)
	[ "$got" = "rumdl ${RUMDL_VERSION}" ]
}

if ! version_matches "$dest"; then
	echo "Installing rumdl ${RUMDL_VERSION} (${target})..." >&2

	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT

	archive="rumdl-v${RUMDL_VERSION}-${target}.tar.gz"
	base_url="https://github.com/rvben/rumdl/releases/download/v${RUMDL_VERSION}"

	curl -fsSL -o "${tmp}/${archive}" "${base_url}/${archive}"
	curl -fsSL -o "${tmp}/${archive}.sha256" "${base_url}/${archive}.sha256"

	if command -v sha256sum >/dev/null 2>&1; then
		(cd "$tmp" && sha256sum -c "${archive}.sha256")
	elif command -v shasum >/dev/null 2>&1; then
		(cd "$tmp" && shasum -a 256 -c "${archive}.sha256")
	else
		echo "rumdl: sha256sum or shasum is required to verify the download" >&2
		exit 1
	fi

	tar -xzf "${tmp}/${archive}" -C "$tmp" rumdl
	mkdir -p "$dest_dir"
	mv "${tmp}/rumdl" "$dest"
	chmod 755 "$dest"
	rm -rf "$tmp"
	trap - EXIT

	if ! version_matches "$dest"; then
		echo "rumdl: installed binary did not report version ${RUMDL_VERSION}" >&2
		exit 1
	fi
fi

exec "$dest" "$@"
