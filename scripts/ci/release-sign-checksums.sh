#!/bin/sh
# Download checksums.txt from a draft release, verify SHA-256, sign, upload .minisig.
# POSIX sh, set -eu in callers.

set -eu

REPO_ROOT=${REPO_ROOT:-.}
PUB_KEY_FILE="${REPO_ROOT}/internal/update/minisign.pub"

usage() {
	echo "usage: release-sign-checksums.sh run --tag TAG --checksums-sha256 HEX" >&2
	echo "       release-sign-checksums.sh --self-test" >&2
}

valid_tag() {
	tag=$1
	case "$tag" in
	v*) version=${tag#v} ;;
	*)
		echo "release-sign-checksums: tag must be vX.Y.Z, got ${tag}" >&2
		return 1
		;;
	esac
	printf '%s\n' "$version" | grep -Eq '^[0-9]+\.[0-9]+\.[0-9]+$'
}

valid_sha256() {
	echo "$1" | grep -Eiq '^[0-9a-f]{64}$'
}

normalize_sha256() {
	echo "$1" | tr '[:upper:]' '[:lower:]'
}

file_sha256() {
	sha256sum "$1" | awk '{print $1}'
}

run_sign() {
	tag=$1
	expected_sha256=$2

	if ! valid_tag "$tag"; then
		echo "release-sign-checksums: invalid tag ${tag}" >&2
		exit 1
	fi
	if ! valid_sha256 "$expected_sha256"; then
		echo "release-sign-checksums: checksums-sha256 must be 64 hex digits" >&2
		exit 1
	fi
	if [ -z "${GH_TOKEN:-}" ]; then
		echo "release-sign-checksums: GH_TOKEN is required" >&2
		exit 1
	fi
	if [ -z "${MINISIGN_SECRET_KEY:-}" ]; then
		echo "release-sign-checksums: MINISIGN_SECRET_KEY is required" >&2
		exit 1
	fi
	if [ ! -f "$PUB_KEY_FILE" ]; then
		echo "release-sign-checksums: missing ${PUB_KEY_FILE}" >&2
		exit 1
	fi

	expected_sha256=$(normalize_sha256 "$expected_sha256")

	is_draft=$(gh release view "$tag" --json isDraft --jq '.isDraft')
	if [ "$is_draft" != "true" ]; then
		echo "release-sign-checksums: ${tag} must be a draft release" >&2
		exit 1
	fi

	if gh release view "$tag" --json assets --jq '.assets[].name' | grep -qx 'checksums.txt.minisig'; then
		echo "release-sign-checksums: ${tag} already has checksums.txt.minisig" >&2
		exit 1
	fi

	workdir=$(mktemp -d)
	# shellcheck disable=SC2064
	trap 'rm -rf "$workdir"' EXIT INT HUP TERM

	gh release download "$tag" --pattern 'checksums.txt' --dir "$workdir" --clobber

	checksums="${workdir}/checksums.txt"
	if [ ! -f "$checksums" ]; then
		echo "release-sign-checksums: checksums.txt missing from ${tag}" >&2
		exit 1
	fi

	actual_sha256=$(file_sha256 "$checksums")
	actual_sha256=$(normalize_sha256 "$actual_sha256")
	if [ "$actual_sha256" != "$expected_sha256" ]; then
		echo "release-sign-checksums: checksums.txt SHA-256 mismatch" >&2
		echo "release-sign-checksums: expected ${expected_sha256}" >&2
		echo "release-sign-checksums: got      ${actual_sha256}" >&2
		exit 1
	fi

	secret_file="${workdir}/minisign.key"
	umask 077
	printf '%s' "$MINISIGN_SECRET_KEY" >"$secret_file"

	minisign -Sm "$checksums" -s "$secret_file" -x "${checksums}.minisig"

	public_key=$(sed -n '2p' "$PUB_KEY_FILE")
	if ! minisign -Vm "$checksums" -x "${checksums}.minisig" -P "$public_key" >/dev/null 2>&1; then
		echo "release-sign-checksums: signature does not match ${PUB_KEY_FILE}" >&2
		exit 1
	fi

	gh release upload "$tag" "${checksums}.minisig"
}

self_test() {
	workdir=$(mktemp -d)
	# shellcheck disable=SC2064
	trap 'rm -rf "$workdir"' EXIT INT HUP TERM

	if ! valid_tag 'v1.2.3'; then
		echo 'release-sign-checksums self-test: valid_tag v1.2.3' >&2
		exit 1
	fi
	if valid_tag 'v1.2'; then
		echo 'release-sign-checksums self-test: reject v1.2' >&2
		exit 1
	fi
	if ! valid_sha256 'abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789'; then
		echo 'release-sign-checksums self-test: valid_sha256' >&2
		exit 1
	fi
	if valid_sha256 'not-a-hash'; then
		echo 'release-sign-checksums self-test: reject bad sha256' >&2
		exit 1
	fi

	checksums="${workdir}/checksums.txt"
	printf 'sha256  sample\n' >"$checksums"
	expected=$(file_sha256 "$checksums")
	expected=$(normalize_sha256 "$expected")
	actual=$(normalize_sha256 "$(file_sha256 "$checksums")")
	if [ "$expected" != "$actual" ]; then
		echo 'release-sign-checksums self-test: sha256 round-trip' >&2
		exit 1
	fi

	minisign -G -W -p "${workdir}/pub.key" -s "${workdir}/sec.key" >/dev/null 2>&1 || {
		if ! command -v minisign >/dev/null 2>&1; then
			echo 'release-sign-checksums self-test: minisign not on PATH' >&2
			exit 1
		fi
		echo 'release-sign-checksums self-test: minisign -G failed' >&2
		exit 1
	}
	minisign -Sm "$checksums" -s "${workdir}/sec.key" -x "${checksums}.minisig"
	pub_b64=$(sed -n '2p' "${workdir}/pub.key")
	if ! minisign -Vm "$checksums" -x "${checksums}.minisig" -P "$pub_b64" >/dev/null 2>&1; then
		echo 'release-sign-checksums self-test: minisign verify' >&2
		exit 1
	fi
}

main() {
	case ${1:-} in
	run)
		shift
		tag=
		expected_sha256=
		while [ $# -gt 0 ]; do
			case $1 in
			--tag)
				tag=$2
				shift 2
				;;
			--checksums-sha256)
				expected_sha256=$2
				shift 2
				;;
			*)
				usage
				exit 1
				;;
			esac
		done
		if [ -z "$tag" ] || [ -z "$expected_sha256" ]; then
			usage
			exit 1
		fi
		run_sign "$tag" "$expected_sha256"
		;;
	--self-test)
		self_test
		;;
	-h | --help)
		usage
		;;
	*)
		usage
		exit 1
		;;
	esac
}

main "$@"