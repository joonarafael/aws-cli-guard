#!/bin/sh
# POSIX compliant, ready to run on Linux or macOS

set -eu

# Compare the newest AWS CLI v2 tag with the verified ceiling in
# internal/cliversion/version.go. Keeps one labeled issue up to date.
#
# Usage:
#   aws-cli-ceiling-watch.sh run [--dry-run]
#   aws-cli-ceiling-watch.sh --self-test
#
# run requires GH_TOKEN, GH_REPO, and RUN_URL.

MARKER_PREFIX='<!-- aws-cli-ceiling-watch'
LABEL_NAME='aws-cli-ceiling'
REPO_ROOT=${REPO_ROOT:-.}
CEILING_FILE="${REPO_ROOT}/internal/cliversion/version.go"
AWS_CLI_REPO='https://github.com/aws/aws-cli.git'
TAG_PATTERN='refs/tags/2.*'

usage() {
	echo "usage: aws-cli-ceiling-watch.sh run [--dry-run]" >&2
	echo "       aws-cli-ceiling-watch.sh --self-test" >&2
}

require_snippet() {
	file=$1
	snippet=$2

	if ! grep -F -q -- "$snippet" "$file"; then
		echo "aws-cli-ceiling-watch self-test: missing snippet: $snippet" >&2
		echo "file was:" >&2
		cat "$file" >&2
		exit 1
	fi
}

reject_snippet() {
	file=$1
	snippet=$2

	if grep -F -q -- "$snippet" "$file"; then
		echo "aws-cli-ceiling-watch self-test: unexpected snippet: $snippet" >&2
		echo "file was:" >&2
		cat "$file" >&2
		exit 1
	fi
}

valid_tag() {
	echo "$1" | grep -E '^2\.[0-9]+\.[0-9]+$' >/dev/null 2>&1
}

read_ceiling() {
	file=$1
	ceiling=$(
		awk '
			/^[[:space:]]*ceilingMajor = / { maj = $3 }
			/^[[:space:]]*ceilingMinor = / { min = $3 }
			/^[[:space:]]*ceilingPatch = / { patch = $3 }
			END {
				if (maj == "" || min == "" || patch == "") {
					exit 1
				}
				printf "%s.%s.%s", maj, min, patch
			}
		' "$file"
	) || {
		echo "aws-cli-ceiling-watch: could not read ceiling from $file" >&2
		exit 1
	}

	case "$ceiling" in
	*[!0-9.]* | *.*.*.*)
		echo "aws-cli-ceiling-watch: invalid ceiling: $ceiling" >&2
		exit 1
		;;
	esac

	echo "$ceiling"
}

pick_latest() {
	latest=$(
		sed -n 's/.*refs\/tags\///p' |
			grep -E '^2\.[0-9]+\.[0-9]+$' |
			sort -t. -k1,1n -k2,2n -k3,3n |
			tail -n 1
	)

	if [ -z "$latest" ]; then
		echo "aws-cli-ceiling-watch: no AWS CLI v2 tags in input" >&2
		exit 1
	fi

	echo "$latest"
}

version_cmp() {
	a=$1
	b=$2

	awk -v a="$a" -v b="$b" '
	function nparts(v, arr,    i, n) {
		n = split(v, arr, ".")
		for (i = 1; i <= n; i++) {
			arr[i] += 0
		}
		return n
	}
	BEGIN {
		na = nparts(a, ap)
		nb = nparts(b, bp)
		max = na > nb ? na : nb
		for (i = 1; i <= max; i++) {
			av = i <= na ? ap[i] : 0
			bv = i <= nb ? bp[i] : 0
			if (av < bv) {
				print -1
				exit
			}
			if (av > bv) {
				print 1
				exit
			}
		}
		print 0
	}
	'
}

normalize_options() {
	file=$1
	if ! jq -e '.options' "$file" >/dev/null 2>&1; then
		echo "aws-cli-ceiling-watch: missing .options in $file" >&2
		exit 1
	fi

	jq -S '.options | map_values(del(.help, .dest, .default))' "$file"
}

diff_options() {
	old_file=$1
	new_file=$2

	old_json=$(normalize_options "$old_file")
	new_json=$(normalize_options "$new_file")

	jq -n -r --argjson o "$old_json" --argjson n "$new_json" '
		def opt_keys(obj):
			obj | keys | sort;
		(opt_keys($o) + opt_keys($n) | unique)[] as $k |
		if ($o | has($k)) and ($n | has($k)) then
			if $o[$k] == $n[$k] then
				empty
			else
				"changed --\($k) \($o[$k] | tojson) -> \($n[$k] | tojson)"
			end
		elif $o | has($k) then
			"removed --\($k)"
		else
			"added --\($k) \($n[$k] | tojson)"
		end
	'
}

fingerprint_from_diff() {
	status=$1
	diff_text=$2

	if [ "$status" = "unavailable" ]; then
		echo unavailable
		return 0
	fi

	printf '%s' "$diff_text" | cksum | awk '{print $1}'
}

marker_value() {
	body=$1
	key=$2

	echo "$body" | grep -F 'aws-cli-ceiling-watch' | head -n 1 |
		sed -n "s/.*${key}=\\([^ ]*\\).*/\\1/p"
}

decide() {
	latest=$1
	ceiling=$2
	issue_num=$3
	prev_fp=$4
	new_fp=$5

	cmp=$(version_cmp "$latest" "$ceiling")

	if [ "$cmp" -lt 0 ]; then
		echo error
		return 0
	fi

	if [ "$cmp" -eq 0 ]; then
		if [ -z "$issue_num" ]; then
			echo none
		else
			echo close
		fi
		return 0
	fi

	if [ -z "$issue_num" ]; then
		echo create
		return 0
	fi

	if [ "$prev_fp" = "$new_fp" ]; then
		echo update
	else
		echo update_comment
	fi
}

options_section_from_diff() {
	status=$1
	diff_text=$2

	if [ "$status" = "unavailable" ]; then
		echo 'comparison unavailable'
		return 0
	fi

	if [ -z "$diff_text" ]; then
		echo 'no changes'
		return 0
	fi

	printf '%s\n' "$diff_text"
}

render_marker() {
	latest=$1
	ceiling=$2
	fp=$3

	printf '%s latest=%s ceiling=%s options=%s -->' "$MARKER_PREFIX" "$latest" "$ceiling" "$fp"
}

contrib_link() {
	if [ -n "${GH_REPO:-}" ]; then
		printf 'https://github.com/%s/blob/master/CONTRIBUTING.md#raise-the-aws-cli-verified-ceiling' "$GH_REPO"
	else
		echo 'https://github.com/example/aws-cli-guard/blob/master/CONTRIBUTING.md#raise-the-aws-cli-verified-ceiling'
	fi
}

render_body() {
	latest=$1
	ceiling=$2
	fp=$3
	options_section=$4
	run_url=$5
	contrib=$(contrib_link)

	{
		render_marker "$latest" "$ceiling" "$fp"
		echo ''
		echo "Newest AWS CLI v2 release is **${latest}**; verified ceiling is **${ceiling}**."
		echo ''
		echo "- Compare: https://github.com/aws/aws-cli/compare/${ceiling}...${latest}"
		echo "- Changelog: https://github.com/aws/aws-cli/blob/${latest}/CHANGELOG.rst"
		echo ''
		echo '### Global options in `awscli/data/cli.json`'
		echo ''
		printf '%s\n' "$options_section"
		echo ''
		echo "See [Raise the AWS CLI verified ceiling](${contrib}) in CONTRIBUTING.md."
		echo ''
		echo "Last checked: ${run_url}"
	}
}

render_title() {
	latest=$1
	ceiling=$2

	echo "AWS CLI ${latest} is newer than verified ceiling ${ceiling}"
}

render_comment() {
	status=$1
	diff_text=$2
	prev_diff_empty=$3

	if [ "$status" = "unavailable" ]; then
		echo 'Global options comparison is unavailable for this run.'
		return 0
	fi

	if [ -z "$diff_text" ]; then
		if [ "$prev_diff_empty" = "1" ]; then
			echo 'Global options in `awscli/data/cli.json` still match the verified ceiling.'
		else
			echo 'Global options in `awscli/data/cli.json` now match the verified ceiling release.'
		fi
		return 0
	fi

	echo 'Global options in `awscli/data/cli.json` changed:'
	echo ''
	printf '%s\n' "$diff_text"
}

fetch_cli_json() {
	tag=$1
	dest=$2

	if ! valid_tag "$tag"; then
		echo "aws-cli-ceiling-watch: invalid tag: $tag" >&2
		return 1
	fi

	url="https://raw.githubusercontent.com/aws/aws-cli/${tag}/awscli/data/cli.json"
	if ! curl -fsSL --retry 3 "$url" -o "$dest"; then
		echo "aws-cli-ceiling-watch: fetch failed for $tag" >&2
		return 1
	fi

	return 0
}

find_open_watch_issue() {
	repo=$1
	issues_json=$(
		gh issue list --repo "$repo" --label "$LABEL_NAME" --state open \
			--json number,body --limit 100
	)

	matches=$(
		echo "$issues_json" | jq '[.[] | select(.body | contains("aws-cli-ceiling-watch"))] | sort_by(.number)'
	)

	count=$(echo "$matches" | jq 'length')

	if [ "$count" -eq 0 ]; then
		echo ''
		return 0
	fi

	if [ "$count" -gt 1 ]; then
		numbers=$(echo "$matches" | jq -r '[.[].number] | join(", ")')
		echo "aws-cli-ceiling-watch: warning: multiple open watch issues: $numbers" >&2
	fi

	echo "$matches" | jq -r '.[0].number'
}

write_issue_body_file() {
	body_file=$1
	shift
	render_body "$@" > "$body_file"
}

run_watch() {
	dry_run=${1:-0}
	comparison_failed=0

	: "${GH_TOKEN:?GH_TOKEN is required}"
	: "${GH_REPO:?GH_REPO is required}"
	: "${RUN_URL:?RUN_URL is required}"

	export GH_TOKEN

	ceiling=$(read_ceiling "$CEILING_FILE")

	latest=$(
		git ls-remote --tags --refs "$AWS_CLI_REPO" "$TAG_PATTERN" | pick_latest
	)

	cmp=$(version_cmp "$latest" "$ceiling")
	if [ "$cmp" -lt 0 ]; then
		echo "aws-cli-ceiling-watch: latest $latest is below ceiling $ceiling" >&2
		exit 1
	fi

	diff_text=''
	cmp_status=ok
	tmp_dir=$(mktemp -d)
	trap 'rm -rf "$tmp_dir"' EXIT

	if [ "$cmp" -gt 0 ]; then
		ceiling_json="${tmp_dir}/ceiling.json"
		latest_json="${tmp_dir}/latest.json"

		if ! fetch_cli_json "$ceiling" "$ceiling_json" ||
			! fetch_cli_json "$latest" "$latest_json"; then
			cmp_status=unavailable
			comparison_failed=1
		else
			diff_text=$(diff_options "$ceiling_json" "$latest_json" | sort || true)
		fi
	fi

	new_fp=$(fingerprint_from_diff "$cmp_status" "$diff_text")
	options_section=$(options_section_from_diff "$cmp_status" "$diff_text")

	if [ "$dry_run" -eq 0 ]; then
		gh label create "$LABEL_NAME" --repo "$GH_REPO" --color bfd4f2 --force >/dev/null 2>&1 ||
			gh label create "$LABEL_NAME" --repo "$GH_REPO" --force >/dev/null
	fi

	issue_num=$(find_open_watch_issue "$GH_REPO")

	prev_fp=''
	if [ -n "$issue_num" ]; then
		issue_body=$(gh issue view "$issue_num" --repo "$GH_REPO" --json body -q .body)
		prev_fp=$(marker_value "$issue_body" options)
	fi

	action=$(decide "$latest" "$ceiling" "$issue_num" "$prev_fp" "$new_fp")

	echo "aws-cli-ceiling-watch: latest=${latest} ceiling=${ceiling} action=${action}"

	body_file=$(mktemp)
	comment_file=$(mktemp)
	title=$(render_title "$latest" "$ceiling")

	case "$action" in
	error)
		echo "aws-cli-ceiling-watch: decide returned error" >&2
		exit 1
		;;
	none)
		;;
	close)
		close_comment="Verified ceiling is now ${ceiling}; newest AWS CLI v2 release matches."
		if [ "$dry_run" -eq 1 ]; then
			echo "$close_comment"
		else
			gh issue close "$issue_num" --repo "$GH_REPO" --comment "$close_comment"
		fi
		;;
	create | update | update_comment)
		write_issue_body_file "$body_file" "$latest" "$ceiling" "$new_fp" "$options_section" "$RUN_URL"
		if [ "$dry_run" -eq 1 ]; then
			cat "$body_file"
		elif [ "$action" = create ]; then
			gh issue create --repo "$GH_REPO" --title "$title" --label "$LABEL_NAME" \
				--body-file "$body_file"
		else
			gh issue edit "$issue_num" --repo "$GH_REPO" --title "$title" --body-file "$body_file"
		fi

		if [ "$action" = update_comment ]; then
			prev_empty=0
			if [ -z "$prev_fp" ] || [ "$prev_fp" = unavailable ]; then
				prev_empty=1
			fi
			render_comment "$cmp_status" "$diff_text" "$prev_empty" > "$comment_file"
			if [ "$dry_run" -eq 1 ]; then
				echo '--- comment ---'
				cat "$comment_file"
			else
				gh issue comment "$issue_num" --repo "$GH_REPO" --body-file "$comment_file"
			fi
		fi
		;;
	*)
		echo "aws-cli-ceiling-watch: unknown action: $action" >&2
		exit 1
		;;
	esac

	rm -f "$body_file" "$comment_file"

	if [ "$comparison_failed" -eq 1 ]; then
		exit 1
	fi
}

self_test() {
	tmp=$(mktemp -d)
	trap 'rm -rf "$tmp"' EXIT

	ls_remote=${tmp}/ls-remote.txt
	cat > "$ls_remote" <<'EOF'
deadbeef	refs/tags/1.36.1
aaa	refs/tags/2.9.10
bbb	refs/tags/2.10.0
ccc	refs/tags/2.10.0-rc1
ddd	refs/tags/not-a-version
EOF

	got=$(pick_latest < "$ls_remote")
	if [ "$got" != "2.10.0" ]; then
		echo "aws-cli-ceiling-watch self-test: pick_latest want 2.10.0 got $got" >&2
		exit 1
	fi

	version_go=${tmp}/version.go
	cat > "$version_go" <<'EOF'
	ceilingMajor = 2
	ceilingMinor = 40
	ceilingPatch = 3
EOF
	got=$(read_ceiling "$version_go")
	if [ "$got" != "2.40.3" ]; then
		echo "aws-cli-ceiling-watch self-test: read_ceiling want 2.40.3 got $got" >&2
		exit 1
	fi

	if [ "$(version_cmp 2.9.10 2.10.0)" -ne -1 ] ||
		[ "$(version_cmp 2.10.0 2.10.0)" -ne 0 ] ||
		[ "$(version_cmp 2.11.0 2.10.0)" -ne 1 ]; then
		echo "aws-cli-ceiling-watch self-test: version_cmp failed" >&2
		exit 1
	fi

	old_json=${tmp}/old-cli.json
	new_json=${tmp}/new-cli.json
	cat > "$old_json" <<'EOF'
{
  "options": {
    "output": {
      "type": "string",
      "choices": ["json", "text", "table"],
      "help": "ignored",
      "dest": "ignored"
    },
    "region": {
      "type": "string",
      "help": "region help"
    },
    "debug": {
      "action": "store_true",
      "type": "string"
    }
  }
}
EOF
	cat > "$new_json" <<'EOF'
{
  "options": {
    "output": {
      "type": "string",
      "choices": ["json", "text", "table", "off"],
      "default": "json"
    },
    "region": {
      "type": "string"
    },
    "help": {
      "action": "store_true",
      "type": "string"
    }
  }
}
EOF

	diff_file=${tmp}/diff.txt
	diff_options "$old_json" "$new_json" | sort > "$diff_file"
	diff_out=$(cat "$diff_file")
	require_snippet "$diff_file" 'added --help'
	require_snippet "$diff_file" 'removed --debug'
	require_snippet "$diff_file" 'changed --output'

	empty_diff=$(diff_options "$old_json" "$old_json")
	if [ -n "$empty_diff" ]; then
		echo "aws-cli-ceiling-watch self-test: expected empty diff" >&2
		exit 1
	fi

	fp=$(fingerprint_from_diff ok "$diff_out")
	if [ -z "$fp" ]; then
		echo "aws-cli-ceiling-watch self-test: empty fingerprint" >&2
		exit 1
	fi
	if [ "$(fingerprint_from_diff unavailable '')" != unavailable ]; then
		echo "aws-cli-ceiling-watch self-test: unavailable fingerprint" >&2
		exit 1
	fi

	sample_body="${MARKER_PREFIX} latest=2.41.0 ceiling=2.40.3 options=${fp} -->"
	if [ "$(marker_value "$sample_body" latest)" != "2.41.0" ] ||
		[ "$(marker_value "$sample_body" ceiling)" != "2.40.3" ] ||
		[ "$(marker_value "$sample_body" options)" != "$fp" ]; then
		echo "aws-cli-ceiling-watch self-test: marker_value failed" >&2
		exit 1
	fi

	assert_decide() {
		want=$1
		shift
		got=$(decide "$@")
		if [ "$got" != "$want" ]; then
			echo "aws-cli-ceiling-watch self-test: decide want $want got $got ($*)" >&2
			exit 1
		fi
	}

	assert_decide error 2.39.0 2.40.3 '' '' ''
	assert_decide none 2.40.3 2.40.3 '' abc abc
	assert_decide close 2.40.3 2.40.3 12 abc abc
	assert_decide create 2.41.0 2.40.3 '' '' abc
	assert_decide update 2.41.0 2.40.3 12 abc abc
	assert_decide update_comment 2.41.0 2.40.3 12 abc def

	body_file=${tmp}/body.md
	render_body 2.41.0 2.40.3 abc 'no changes' 'https://example.test/run/1' > "$body_file"
	require_snippet "$body_file" "latest=2.41.0 ceiling=2.40.3 options=abc"
	require_snippet "$body_file" 'Newest AWS CLI v2 release is **2.41.0**'
	require_snippet "$body_file" 'compare/2.40.3...2.41.0'
	require_snippet "$body_file" 'blob/2.41.0/CHANGELOG.rst'
	require_snippet "$body_file" 'raise-the-aws-cli-verified-ceiling'
	require_snippet "$body_file" 'https://example.test/run/1'

	comment_file=${tmp}/comment.md
	render_comment ok "$diff_out" 0 > "$comment_file"
	require_snippet "$comment_file" 'Global options in `awscli/data/cli.json` changed:'
	render_comment unavailable '' 0 > "$comment_file"
	require_snippet "$comment_file" 'unavailable'
	render_comment ok '' 1 > "$comment_file"
	require_snippet "$comment_file" 'still match'

	echo "aws-cli-ceiling-watch self-test passed."
}

if [ "${1:-}" = "--self-test" ]; then
	if [ "$#" -ne 1 ]; then
		usage
		exit 1
	fi

	self_test
	exit 0
fi

if [ "${1:-}" = run ]; then
	shift
	dry_run=0
	if [ "${1:-}" = --dry-run ]; then
		dry_run=1
		shift
	fi
	if [ "$#" -ne 0 ]; then
		usage
		exit 1
	fi

	run_watch "$dry_run"
	exit 0
fi

usage
exit 1