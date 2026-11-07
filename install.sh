#!/bin/sh
set -eu

usage() {
	printf '%s\n' 'aws-cli-guard: error: usage: install.sh --version X.Y.Z [--prefix dir] [--base-url url]' >&2
	exit 3
}

error() {
	printf '%s\n' "aws-cli-guard: error: $1" >&2
	exit 3
}

version=''
prefix=''
base_url=''

while [ "$#" -gt 0 ]; do
	case "$1" in
	--version)
		shift
		[ "$#" -gt 0 ] || usage
		version=$1
		;;
	--prefix)
		shift
		[ "$#" -gt 0 ] || usage
		prefix=$1
		;;
	--base-url)
		shift
		[ "$#" -gt 0 ] || usage
		base_url=$1
		;;
	-*)
		usage
		;;
	*)
		usage
		;;
	esac
	shift
done

[ -n "$version" ] || usage

old_ifs=$IFS
IFS=.
set -- $version
IFS=$old_ifs
if [ "$#" -ne 3 ]; then
	error 'version must match X.Y.Z'
fi
for part in "$1" "$2" "$3"; do
	case "$part" in
	'' | *[!0-9]*)
		error 'version must match X.Y.Z'
		;;
	esac
done

if ! command -v curl >/dev/null 2>&1; then
	error 'curl is required'
fi

os_name=$(uname -s)
arch_name=$(uname -m)

case "$os_name" in
Linux)
	goos=linux
	;;
Darwin)
	goos=darwin
	;;
*)
	error 'unsupported operating system'
	;;
esac

case "$arch_name" in
x86_64)
	goarch=amd64
	;;
aarch64 | arm64)
	goarch=arm64
	;;
*)
	error 'unsupported CPU architecture'
	;;
esac

if [ -z "$base_url" ]; then
	base_url="https://github.com/joonarafael/aws-cli-guard/releases/download/v${version}"
fi

archive="aws-cli-guard_${version}_${goos}_${goarch}.tar.gz"
tmp=$(mktemp -d)
cleanup() {
	rm -rf "$tmp"
}
trap cleanup EXIT

curl -fsSL "${base_url}/${archive}" -o "${tmp}/${archive}"
curl -fsSL "${base_url}/checksums.txt" -o "${tmp}/checksums.txt"

checksum_line=$(grep "  ${archive}$" "${tmp}/checksums.txt" || true)
[ -n "$checksum_line" ] || error 'checksum for archive not found'

expected=$(printf '%s' "$checksum_line" | awk '{print $1}')

if command -v sha256sum >/dev/null 2>&1; then
	actual=$(sha256sum "${tmp}/${archive}" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
	actual=$(shasum -a 256 "${tmp}/${archive}" | awk '{print $1}')
else
	error 'sha256sum or shasum is required'
fi

[ "$actual" = "$expected" ] || error 'archive checksum mismatch'

members=$(tar -tzf "${tmp}/${archive}")
member_count=0
for member in $members; do
	case "$member" in
	aws | aws-guard | aws-cli-guard-disable) ;;
	.. | ../* | */.. | */../*)
		error 'archive contains parent path'
		;;
	/*)
		error 'archive contains absolute path'
		;;
	*/*)
		error 'archive contains directory members'
		;;
	*)
		error 'archive has unexpected members'
		;;
	esac
	member_count=$((member_count + 1))
done

[ "$member_count" -eq 3 ] || error 'archive must contain exactly three members'

tar -xzf "${tmp}/${archive}" -C "$tmp"

if [ -n "$prefix" ]; then
	"${tmp}/aws-guard" install --prefix "$prefix"
else
	"${tmp}/aws-guard" install
fi
