#!/bin/sh
# POSIX compliant, ready to run on Linux or macOS

set -eu

# Validates .golangci.yml against the JSON schema of the golangci-lint version
# pinned in go.mod. Lint targets and CI run this before golangci-lint.

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT" || exit 1

CONFIG_PATH="$REPO_ROOT/.golangci.yml"

echo "Validating .golangci.yml..."

if ! go tool golangci-lint config verify --config "$CONFIG_PATH"; then
	echo "golangci-lint configuration is invalid." >&2
	exit 1
fi

echo "golangci-lint configuration is valid."
