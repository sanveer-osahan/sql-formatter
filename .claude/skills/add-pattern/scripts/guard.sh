#!/usr/bin/env bash
# Diff guard for /add-pattern (spec D23).
#
#   guard.sh freeze <worktree> <statedir>  Record frozen-set hashes (first call only)
#                                          and take a fresh snapshot of every file.
#   guard.sh check  <worktree> <statedir>  After an A4 iteration: every change since the
#                                          last snapshot that is NOT a non-test .go file
#                                          is a violation. Reverts violations.
#                                          Exit 0 = clean, exit 3 = violations reverted.
#   guard.sh verify <worktree> <statedir>  Check that the frozen set still matches the
#                                          hashes from the first freeze. Exit 0 / 3.
#
# Frozen set: internal/formatter/testdata/patterns/*.json, *_test.go, STYLE.md.
# Files ignored by git (bin/, etc.) are not tracked by the guard. Only files that exist at
# freeze time are recorded; a recorded file that is later missing is a violation (check
# restores it from the snapshot, verify reports it as changed).
set -euo pipefail

cmd=${1:?cmd}; wt=${2:?worktree}; sd=${3:?statedir}
mkdir -p "$sd"

# Only files that exist: a tracked file deleted in the worktree is not listed. If it
# reappears later it counts as a new file (check removes it; verify reports a frozen one).
list_all() {
  git -C "$wt" ls-files -co --exclude-standard | LC_ALL=C sort -u |
    while IFS= read -r f; do if [[ -f $wt/$f ]]; then printf '%s\n' "$f"; fi; done
}
is_code()   { [[ $1 == *.go && $1 != *_test.go ]]; }
is_frozen() { [[ $1 == internal/formatter/testdata/patterns/*.json || $1 == *_test.go || $1 == STYLE.md ]]; }
hash_of()   { (cd "$wt" && shasum -a 256 -- "$1" | cut -d' ' -f1); }

case $cmd in
freeze)
  if [[ ! -f $sd/frozen.sha ]]; then
    list_all | while IFS= read -r f; do
      if is_frozen "$f"; then echo "$(hash_of "$f")  $f"; fi
    done > "$sd/frozen.sha"
  fi
  : > "$sd/snapshot.sha"
  list_all | while IFS= read -r f; do
    if ! is_code "$f"; then echo "$(hash_of "$f")  $f"; fi
  done > "$sd/snapshot.sha"
  (cd "$wt" && cut -d' ' -f3- "$sd/snapshot.sha" | tar -cf "$sd/snapshot.tar" -T -)
  echo "FROZEN: $(wc -l < "$sd/frozen.sha" | tr -d ' ') files; SNAPSHOT: $(wc -l < "$sd/snapshot.sha" | tr -d ' ') files"
  ;;
check)
  [[ -f $sd/snapshot.sha ]] || { echo "no snapshot; run freeze first" >&2; exit 2; }
  violations=()
  # Changed or deleted files that were in the snapshot.
  while IFS= read -r line; do
    h=${line%%  *}; f=${line#*  }
    if [[ ! -e $wt/$f ]] || [[ $(hash_of "$f") != "$h" ]]; then
      (cd "$wt" && tar -xf "$sd/snapshot.tar" -- "$f")
      violations+=("reverted $f")
    fi
  done < "$sd/snapshot.sha"
  # New files that are not non-test .go files.
  while IFS= read -r f; do
    is_code "$f" && continue
    if ! cut -d' ' -f3- "$sd/snapshot.sha" | grep -qxF -- "$f"; then
      rm -f -- "$wt/$f"
      violations+=("removed new file $f")
    fi
  done < <(list_all)
  if ((${#violations[@]})); then
    echo "VIOLATIONS: ${#violations[@]}"; printf '  %s\n' "${violations[@]}"; exit 3
  fi
  echo "VIOLATIONS: 0"
  ;;
verify)
  [[ -f $sd/frozen.sha ]] || { echo "no frozen.sha" >&2; exit 2; }
  bad=0
  while IFS= read -r line; do
    h=${line%%  *}; f=${line#*  }
    if [[ ! -e $wt/$f ]] || [[ $(hash_of "$f") != "$h" ]]; then echo "  changed $f"; bad=1; fi
  done < "$sd/frozen.sha"
  while IFS= read -r f; do
    if is_frozen "$f" && ! cut -d' ' -f3- "$sd/frozen.sha" | grep -qxF -- "$f"; then echo "  new $f"; bad=1; fi
  done < <(list_all)
  if ((bad)); then echo "FROZEN: CHANGED"; exit 3; fi
  echo "FROZEN: OK"
  ;;
*) echo "unknown command $cmd" >&2; exit 2 ;;
esac
