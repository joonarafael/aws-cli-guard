#!/bin/sh
# Build release archives for OP-3 (ADR 0025).
set -eu

error() {
	printf '%s\n' "aws-cli-guard: error: $1" >&2
	exit 3
}

usage() {
	error 'usage: release-assets.sh X.Y.Z outdir'
}

validate_version() {
	version=$1
	old_ifs=$IFS
	IFS=.
	# shellcheck disable=SC2086
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
}

sha256_file() {
	path=$1
	if command -v sha256sum >/dev/null 2>&1; then
		sha256sum "$path" | awk '{print $1}'
	elif command -v shasum >/dev/null 2>&1; then
		shasum -a 256 "$path" | awk '{print $1}'
	else
		error 'sha256sum or shasum is required'
	fi
}

if [ "$#" -ne 2 ]; then
	usage
fi

version=$1
outdir=$2

case "$version" in
v*)
	error 'version must match X.Y.Z'
	;;
*-*)
	error 'version must match X.Y.Z'
	;;
esac

validate_version "$version"

script_dir=$(
	cd -- "$(dirname "$0")" && pwd
)
repo_root=$(
	cd -- "$script_dir/.." && pwd
)
cd "$repo_root" || error 'cannot enter repository root'

mkdir -p "$outdir"

ldflags="-X aws-cli-guard/internal/version.Version=${version}"
disable_src="${repo_root}/scripts/aws-cli-guard-disable"

if [ ! -f "$disable_src" ]; then
	error 'missing aws-cli-guard-disable script'
fi

build_platform() {
	goos=$1
	goarch=$2

	stage=$(mktemp -d)

	CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build \
		-ldflags "$ldflags" \
		-o "${stage}/aws" \
		./cmd/aws
	CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build \
		-ldflags "$ldflags" \
		-o "${stage}/aws-guard" \
		./cmd/aws-guard
	cp "$disable_src" "${stage}/aws-cli-guard-disable"
	chmod 755 "${stage}/aws" "${stage}/aws-guard" "${stage}/aws-cli-guard-disable"

	archive="aws-cli-guard_${version}_${goos}_${goarch}.tar.gz"
	(
		cd "$stage" || exit 1
		tar --format=ustar -czf "${outdir}/${archive}" \
			aws aws-guard aws-cli-guard-disable
	)

	rm -rf "$stage"
}

build_platform linux amd64
build_platform linux arm64
build_platform darwin amd64
build_platform darwin arm64

checksums_path="${outdir}/checksums.txt"
: >"$checksums_path"

for archive in \
	"aws-cli-guard_${version}_linux_amd64.tar.gz" \
	"aws-cli-guard_${version}_linux_arm64.tar.gz" \
	"aws-cli-guard_${version}_darwin_amd64.tar.gz" \
	"aws-cli-guard_${version}_darwin_arm64.tar.gz"
do
	path="${outdir}/${archive}"
	if [ ! -f "$path" ]; then
		error "missing archive ${archive}"
	fi
	digest=$(sha256_file "$path")
	printf '%s  %s\n' "$digest" "$archive" >>"$checksums_path"
done

cp "${repo_root}/install.sh" "${outdir}/install.sh"
chmod 755 "${outdir}/install.sh"
