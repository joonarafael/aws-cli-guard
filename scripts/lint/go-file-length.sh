#!/bin/sh
# POSIX compliant, ready to run on Linux or macOS

set -eu

# Script to check that no .go file (including _test.go files) exceeds a maximum line count.
#
# Scans tracked and untracked worktree files; gitignored paths are skipped.
#
# golangci-lint has no native linter for maximum file length: `funlen` caps function length,
# `lll` caps line length, but neither caps the number of lines in a whole file. If golangci-lint
# ever ships a linter for this, prefer it and delete this script (plus its CI wiring).

REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT" || exit 1

MAX_LINES=300

echo "Checking that no .go file exceeds ${MAX_LINES} lines (test files included, gitignored paths skipped)..."

violations=$(
	git ls-files --cached --others --exclude-standard -- '*.go' | while IFS= read -r file; do
		[ -f "$file" ] || continue

		lines=$(wc -l < "$file" | tr -d '[:space:]')
		if [ "$lines" -gt "$MAX_LINES" ]; then
			printf '%s %s\n' "$lines" "$file"
		fi
	done
)

if [ -n "$violations" ]; then
	echo "The following .go files exceed ${MAX_LINES} lines:" >&2
	echo "$violations" | sort -k2 >&2
	echo "" >&2
	echo "Please split the above files into smaller, more focused files." >&2
	exit 1
fi

echo "No .go files exceed ${MAX_LINES} lines!"
