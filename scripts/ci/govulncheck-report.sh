#!/bin/sh
# POSIX compliant, ready to run on Linux or macOS

set -eu

# Turn a govulncheck log into the sticky pull-request comment and job summary.
# The CI job stays green when the scanner finds vulnerabilities or throws.
# See ADR 0031.
#
# GO_VERSION and GOVULNCHECK_VERSION must be set. GOVULNCHECK_REPORT_LIMIT
# caps how many findings the comment lists (default 30).
#
# Usage:
#   govulncheck-report.sh EXIT_CODE LOG_FILE OUTPUT_FILE RUN_URL
#   govulncheck-report.sh --self-test

usage() {
	echo "usage: govulncheck-report.sh EXIT_CODE LOG_FILE OUTPUT_FILE RUN_URL" >&2
	echo "       govulncheck-report.sh --self-test" >&2
}

require_snippet() {
	file=$1
	snippet=$2

	if ! grep -F -q -- "$snippet" "$file"; then
		echo "govulncheck-report self-test: missing snippet: $snippet" >&2
		echo "file was:" >&2
		cat "$file" >&2
		exit 1
	fi
}

reject_snippet() {
	file=$1
	snippet=$2

	if grep -F -q -- "$snippet" "$file"; then
		echo "govulncheck-report self-test: unexpected snippet: $snippet" >&2
		echo "file was:" >&2
		cat "$file" >&2
		exit 1
	fi
}

require_integer() {
	value=$1
	label=$2

	case "$value" in
	'' | *[!0-9]*)
		echo "govulncheck-report: ${label} must be a non-negative integer" >&2
		exit 1
		;;
	esac
}

render() {
	exit_code=$1
	log_file=$2
	output_file=$3
	run_url=$4
	list_limit=${GOVULNCHECK_REPORT_LIMIT:-30}

	require_integer "$exit_code" "exit code"
	require_integer "$list_limit" "GOVULNCHECK_REPORT_LIMIT"

	if [ "$list_limit" -lt 1 ]; then
		echo "govulncheck-report: GOVULNCHECK_REPORT_LIMIT must be at least 1" >&2
		exit 1
	fi

	if [ ! -f "$log_file" ]; then
		echo "govulncheck-report: log does not exist: $log_file" >&2
		exit 1
	fi

	: "${GO_VERSION:?GO_VERSION is required}"
	: "${GOVULNCHECK_VERSION:?GOVULNCHECK_VERSION is required}"

	if [ -z "$run_url" ]; then
		echo "govulncheck-report: run URL is required" >&2
		exit 1
	fi

	awk \
		-v exit_code="$exit_code" \
		-v go_version="$GO_VERSION" \
		-v scanner_version="$GOVULNCHECK_VERSION" \
		-v run_url="$run_url" \
		-v list_limit="$list_limit" \
		-f - "$log_file" > "$output_file" <<'AWK'
function save_current() {
	if (cur_id == "") {
		return
	}

	n++
	ids[n] = cur_id
	urls[n] = cur_url
	summaries[n] = cur_summary
	founds[n] = cur_found
	fixeds[n] = cur_fixed
	cur_id = ""
	cur_url = ""
	cur_summary = ""
	cur_found = ""
	cur_fixed = ""
}

function append_field(current, value) {
	if (current == "") {
		return value
	}

	return current "; " value
}

function footer_start(line) {
	return line ~ /^Your code (is|may be) affected / ||
		line ~ /^This scan also found / ||
		line ~ /^This scan found no other /
}

function add_excerpt_line(line) {
	if (excerpt_n >= excerpt_limit) {
		return
	}

	excerpt = excerpt line "\n"
	excerpt_n++
}

function toolchain() {
	print "Pinned toolchain: Go " go_version ", govulncheck " scanner_version "."
	print ""
}

function run_link() {
	print "You can find the full log behind <a href=\"" run_url "\">this link</a>."
	print ""
}

function print_bullet(i,   line) {
	line = "- "

	if (urls[i] != "") {
		line = line "[" ids[i] "](" urls[i] ")"
	} else {
		line = line "`" ids[i] "`"
	}

	if (summaries[i] != "") {
		line = line ": " summaries[i]
	}

	if (founds[i] != "") {
		line = line ". Found in `" founds[i] "`"
	}

	if (fixeds[i] != "") {
		line = line ", fixed in `" fixeds[i] "`"
	}

	print line "."
}

BEGIN {
	excerpt_limit = 8
}

{
	sub(/\r$/, "")
}

/^panic:/ {
	threw = 1

	if (excerpt_n == 0) {
		capture_excerpt = 1
	}
}

capture_excerpt && excerpt_n < excerpt_limit {
	add_excerpt_line($0)

	if (excerpt_n >= excerpt_limit) {
		capture_excerpt = 0
	}
}

!threw && fallback_n < excerpt_limit && $0 !~ /^go: / && $0 != "" {
	fallback = fallback $0 "\n"
	fallback_n++
}

/^Vulnerability #[0-9]+: / {
	save_current()
	cur_id = $0
	sub(/^Vulnerability #[0-9]+: /, "", cur_id)
	next
}

/^  More info: / && cur_id != "" {
	cur_url = $0
	sub(/^  More info: /, "", cur_url)
	next
}

/^    Found in: / && cur_id != "" {
	value = $0
	sub(/^    Found in: /, "", value)
	cur_found = append_field(cur_found, value)
	next
}

/^    Fixed in: / && cur_id != "" {
	value = $0
	sub(/^    Fixed in: /, "", value)
	cur_fixed = append_field(cur_fixed, value)
	next
}

/^    (Example traces found:|Vulnerable symbols found:|Platforms:)/ {
	next
}

cur_id != "" && cur_url == "" && /^    [^ ]/ {
	value = $0
	sub(/^    /, "", value)

	if (cur_summary == "") {
		cur_summary = value
	} else {
		cur_summary = cur_summary " " value
	}

	next
}

footer_start($0) {
	saw_affected = saw_affected || $0 ~ /^Your code (is|may be) affected /
	in_footer = 1

	if (footer != "") {
		footer = footer "\n\n"
	}

	footer = footer $0
	next
}

in_footer && /^Use / {
	in_footer = 0
	next
}

in_footer && /^exit status / {
	in_footer = 0
	next
}

in_footer && /^$/ {
	in_footer = 0
	next
}

in_footer {
	footer = footer " " $0
	next
}

/^exit status 3$/ {
	child_exit = 3
}

END {
	save_current()

	if (exit_code == 0) {
		mode = "success"
	} else if (threw) {
		mode = "threw"
	} else if (n > 0 || child_exit == 3 || saw_affected) {
		mode = "findings"
	} else {
		mode = "failed"
	}

	print "### govulncheck Report"
	print ""

	if (mode == "success") {
		print "## ✅ No vulnerabilities were found."
		print ""
		print "Excellent, nothing to review here!"
		print ""
		toolchain()
		run_link()
		exit
	}

	if (mode == "findings") {
		print "## ❌ govulncheck reported vulnerabilities on the pinned Go toolchain."
	} else {
		print "## ❌ govulncheck failed while scanning the pinned Go toolchain."
	}

	print ""
	toolchain()

	if (mode == "findings") {
		print "**This check does not block merging.** Review the findings when you can."
	} else if (mode == "threw") {
		print "**This check does not block merging.** The scanner threw before it could finish."
	} else {
		print "**This check does not block merging.** The scanner did not finish."
	}

	print ""

	if (footer != "") {
		print footer
		print ""
	}

	if (mode == "findings") {
		shown = n

		if (shown > list_limit) {
			shown = list_limit
		}

		print "<details><summary>Click to expand findings (" n ")</summary>"
		print ""

		if (n == 0) {
			print "The log did not include vulnerability entries."
		}

		for (i = 1; i <= shown; i++) {
			print_bullet(i)
		}

		hidden = n - shown

		if (hidden == 1) {
			print ""
			print "1 more finding is in the log."
		} else if (hidden > 1) {
			print ""
			print hidden " more findings are in the log."
		}

		print ""
		print "</details>"
		print ""
	} else {
		print "<details><summary>Click to expand the scanner error</summary>"
		print ""
		print "````"

		text = excerpt

		if (text == "") {
			text = fallback
		}

		if (text != "") {
			printf "%s", text
		} else {
			print "govulncheck exited " exit_code " and produced no usable error text."
		}

		print "````"
		print ""
		print "</details>"
		print ""
	}

	run_link()
}
AWK
}

self_test() {
	log_file=$(mktemp)
	output_file=$(mktemp)
	trap 'rm -f "$log_file" "$output_file"' EXIT
	GO_VERSION=1.27.4
	GOVULNCHECK_VERSION=v1.1.4
	export GO_VERSION GOVULNCHECK_VERSION

	cat > "$log_file" <<'EOF'
=== Symbol Results ===

Vulnerability #1: GO-0000-0001
    Third-party vulnerability
  More info: https://pkg.go.dev/vuln/GO-0000-0001
  Module: golang.org/vmod
    Found in: golang.org/vmod@v0.0.1
    Fixed in: golang.org/vmod@v0.1.3
    Platforms: amd
    Example traces found:
      #1: main.main calls vmod.Vuln

Vulnerability #2: GO-2026-6613
    HTTP/1 server connection desynchronization after 2xx CONNECT response in
    net/http
  More info: https://pkg.go.dev/vuln/GO-2026-6613
  Standard library
    Found in: net/http@go1.26.3
    Fixed in: net/http@go1.26.9
    Example traces found:
      #1: internal/update/release.go:59:28: update.FetchLatest calls http.Client.Do

Your code is affected by 2 vulnerabilities from the Go standard library.
This scan also found 1 vulnerability in packages you import and 0
vulnerabilities in modules you require, but your code doesn't appear to call
these vulnerabilities.
Use '-show verbose' for more details.
exit status 3
EOF

	# go run reports the child status and exits 1.
	render 1 "$log_file" "$output_file" "https://example.test/run/1"
	require_snippet "$output_file" "### govulncheck Report"
	require_snippet "$output_file" "## ❌ govulncheck reported vulnerabilities on the pinned Go toolchain."
	require_snippet "$output_file" "Pinned toolchain: Go 1.27.4, govulncheck v1.1.4."
	require_snippet "$output_file" "**This check does not block merging.** Review the findings when you can."
	require_snippet "$output_file" "Your code is affected by 2 vulnerabilities from the Go standard library."
	require_snippet "$output_file" "This scan also found 1 vulnerability in packages you import and 0 vulnerabilities in modules you require, but your code doesn't appear to call these vulnerabilities."
	require_snippet "$output_file" "Click to expand findings (2)"
	require_snippet "$output_file" "- [GO-0000-0001](https://pkg.go.dev/vuln/GO-0000-0001): Third-party vulnerability. Found in \`golang.org/vmod@v0.0.1\`, fixed in \`golang.org/vmod@v0.1.3\`."
	require_snippet "$output_file" "- [GO-2026-6613](https://pkg.go.dev/vuln/GO-2026-6613): HTTP/1 server connection desynchronization after 2xx CONNECT response in net/http. Found in \`net/http@go1.26.3\`, fixed in \`net/http@go1.26.9\`."
	require_snippet "$output_file" "<a href=\"https://example.test/run/1\">this link</a>."
	reject_snippet "$output_file" "Platforms:"
	reject_snippet "$output_file" "update.FetchLatest calls http.Client.Do"
	reject_snippet "$output_file" "Use '-show verbose'"
	reject_snippet "$output_file" "Excellent, nothing to review here!"

	GOVULNCHECK_REPORT_LIMIT=1
	export GOVULNCHECK_REPORT_LIMIT
	render 3 "$log_file" "$output_file" "https://example.test/run/1"
	require_snippet "$output_file" "Click to expand findings (2)"
	require_snippet "$output_file" "- [GO-0000-0001](https://pkg.go.dev/vuln/GO-0000-0001): Third-party vulnerability. Found in \`golang.org/vmod@v0.0.1\`, fixed in \`golang.org/vmod@v0.1.3\`."
	require_snippet "$output_file" "1 more finding is in the log."
	reject_snippet "$output_file" "GO-2026-6613"
	unset GOVULNCHECK_REPORT_LIMIT

	printf '%s\n' "No vulnerabilities found." > "$log_file"
	render 0 "$log_file" "$output_file" "https://example.test/run/1"
	require_snippet "$output_file" "## ✅ No vulnerabilities were found."
	require_snippet "$output_file" "Excellent, nothing to review here!"
	require_snippet "$output_file" "Pinned toolchain: Go 1.27.4, govulncheck v1.1.4."
	reject_snippet "$output_file" "does not block merging"

	cat > "$log_file" <<'EOF'
go: downloading golang.org/x/vuln v1.1.4
panic: unexpected expr: *ast.KeyValueExpr

goroutine 1 [running]:
main.main()
	/tmp/govulncheck/main.go:1
	/tmp/govulncheck/main.go:2
	/tmp/govulncheck/main.go:3
	/tmp/govulncheck/main.go:4
	/tmp/govulncheck/main.go:5
SENTINEL-NOT-IN-EXCERPT
exit status 2
EOF

	render 1 "$log_file" "$output_file" "https://example.test/run/1"
	require_snippet "$output_file" "## ❌ govulncheck failed while scanning the pinned Go toolchain."
	require_snippet "$output_file" "**This check does not block merging.** The scanner threw before it could finish."
	require_snippet "$output_file" "panic: unexpected expr: *ast.KeyValueExpr"
	require_snippet "$output_file" "Click to expand the scanner error"
	reject_snippet "$output_file" "SENTINEL-NOT-IN-EXCERPT"
	reject_snippet "$output_file" "go: downloading"
	reject_snippet "$output_file" "Excellent, nothing to review here!"

	printf '%s\n' "flag provided but not defined: -bogus" > "$log_file"
	render 2 "$log_file" "$output_file" "https://example.test/run/1"
	require_snippet "$output_file" "**This check does not block merging.** The scanner did not finish."
	require_snippet "$output_file" "flag provided but not defined: -bogus"
	reject_snippet "$output_file" "The scanner threw"

	echo "govulncheck-report self-test passed."
}

if [ "${1:-}" = "--self-test" ]; then
	if [ "$#" -ne 1 ]; then
		usage
		exit 1
	fi

	self_test
	exit 0
fi

if [ "$#" -ne 4 ]; then
	usage
	exit 1
fi

render "$1" "$2" "$3" "$4"
