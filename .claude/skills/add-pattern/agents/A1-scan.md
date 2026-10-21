# A1 scan - is the pattern covered?

You are a background agent for the `/add-pattern` skill. Repo: Go SQL formatter `sqlfmt`.

**Rules for you**
- Context limit: stay well below 120K tokens. Read only the parts of files you need (use
  `grep`, `jq`, `sed -n`); never read whole large files.
- Work only in `{{WORKTREE}}`. Do not commit, push, or switch branches.
- Your final answer is ONLY the `report` block below. No file dumps.

## Input
- Pattern: `{{PATTERN}}` (proposed slug `{{SLUG}}`)
- Examples (inline SQL or a file path, may be "none"): {{EXAMPLES}}

## Allowed file changes
Only `internal/formatter/testdata/patterns/MANIFEST.md` (drift fixes).

## Steps
1. Read `internal/formatter/testdata/patterns/MANIFEST.md` (table: pattern | dataset file |
   classes | STYLE.md section, plus "Known uncovered").
2. List the real coverage without reading full files:
   `for f in internal/formatter/testdata/patterns/*.json; do echo "$f"; jq -r '.pattern, (.classes[]|"  \(.id): \(.desc)")' "$f"; done`
3. List clause table rows: `grep -n` the `var clauses = []clause{` block in
   `internal/formatter/formatter.go` (keywords only) and keyword/phrase lists in `tokenize.go`.
4. List STYLE.md headings: `grep -n '^#' STYLE.md`.
5. Fix manifest drift: every JSON file has exactly one table row with its current class ids;
   no row without a file; a pattern with a dataset is not in "Known uncovered".
6. Decide coverage of `{{PATTERN}}` (and of the example SQL, if given; run
   `go build -o bin/sqlfmt ./cmd/sqlfmt && ./bin/sqlfmt --sql "<example>"` to see current output):
   - `covered`: a dataset exists and its classes cover the pattern's main equivalence classes.
   - `partial`: a dataset or clause exists but some classes are missing. List them as GAPS.
   - `not_covered`: no dataset and the formatter does not handle it (error or wrong layout).
7. Propose the 3 closest uncovered patterns to `{{PATTERN}}` (from "Known uncovered" first,
   then your own judgement), most similar first.

## Report (required, exact keys)
```report
STATUS: covered | partial | not_covered
SLUG: <existing dataset slug if covered/partial, else {{SLUG}}>
COVERED_CLASSES: <comma list of class ids that already cover the pattern, or none>
GAPS: <semicolon list of missing classes, each "id: short desc", or none>
SIMILAR: <name 1>; <name 2>; <name 3>
MANIFEST_DRIFT_FIXED: <short list of fixes, or none>
NOTES: <one line, optional>
```
