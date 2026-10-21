#!/usr/bin/env bash
# Builds the CLI, then formats every query in the samples file.
# Usage: scripts/run-samples.sh [samples-file]   (default: samples/queries.sql)
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
samples="${1:-$root/samples/queries.sql}"
bin="$root/bin/sqlfmt"

echo "Building $bin"
(cd "$root" && make build)

pattern=""
query=""
total=0
failed=0

run_query() {
	[ -z "$query" ] && return 0
	total=$((total + 1))
	printf '\n=== [%s] #%d\n' "${pattern:-none}" "$total"
	printf '%s\n' "$query"
	echo "---"
	if ! "$bin" --sql "$query"; then
		failed=$((failed + 1))
	fi
	query=""
}

while IFS= read -r line || [ -n "$line" ]; do
	case "$line" in
	"-- pattern:"*)
		run_query
		pattern="${line#-- pattern:}"
		pattern="${pattern# }"
		;;
	--*) ;;
	"")
		run_query
		;;
	*)
		if [ -z "$query" ]; then
			query="$line"
		else
			query="$query"$'\n'"$line"
		fi
		;;
	esac
done <"$samples"
run_query

printf '\n%d queries, %d failed\n' "$total" "$failed"
[ "$failed" -eq 0 ]
