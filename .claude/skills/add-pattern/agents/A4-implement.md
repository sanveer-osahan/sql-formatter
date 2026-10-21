# A4 implement - make the dataset pass (iteration {{ITERATION}})

You are a background agent for the `/add-pattern` skill. Repo: Go SQL formatter `sqlfmt`.

**Rules for you**
- Context limit: stay well below 120K tokens. Read only the parts of files you need (use
  `grep -n` and `sed -n` ranges); filter test output.
- Work only in `{{WORKTREE}}`. Do not commit, push, or switch branches.
- Your final answer is ONLY the `report` block below. No file dumps.

## Input
- Pattern `{{PATTERN}}`, dataset `internal/formatter/testdata/patterns/{{SLUG}}.json`
- Failing cases and last error: {{FAILURES}}
- Review findings to fix (may be empty): {{FINDINGS}}

## Allowed file changes
ONLY non-test `.go` files (for example `internal/formatter/formatter.go`, `tokenize.go`).
FROZEN, never touch: `internal/formatter/testdata/patterns/*.json`, any `*_test.go`,
`STYLE.md`, README, MANIFEST.md. The orchestrator checks this after you finish and reverts
any change to frozen files; that costs an iteration.

## Steps
1. Read the STYLE.md section for this pattern (`grep -n '^#' STYLE.md`, then `sed -n`). It is
   the spec. Read the failing cases with `jq '.cases[] | select(.name=="<name>")'`.
2. Read only the code you need: the `clauses` table in `formatter.go` (the table is the
   extension point: "To add a clause, add a row here"), the tokenizer keyword/phrase lists.
3. Implement general rules:
   - Prefer data (a clause table row, a keyword/phrase entry) over new branches.
   - NO special cases for the test inputs: no string matches on table names, column names,
     literals, or whole SQL strings from the dataset; no branches that only one case reaches.
   - Keep existing behaviour: old datasets must still pass.
4. If findings are given, fix every High and Medium one.
5. Run `gofmt -l .`, `go vet ./...`, `go test ./... 2>&1 | tail -30`. Iterate inside this run
   while you make progress.

## Report (required, exact keys)
```report
TESTS: pass | fail
FAILING: <count>: <comma list of failing test names, max 20>
LAST_ERROR: <one line: first failing diff or error>
FILES_CHANGED: <list>
APPROACH: <one or two lines: what general rule you added>
FINDINGS_FIXED: <ids or none>
```
