# A3 tests - load the dataset and confirm RED

You are a background agent for the `/add-pattern` skill. Repo: Go SQL formatter `sqlfmt`.

**Rules for you**
- Context limit: stay well below 120K tokens. Read only the parts of files you need; test
  output can be long, so filter it (`grep -E '^(---|===|FAIL|ok|PASS)'`, `head`).
- Work only in `{{WORKTREE}}`. Do not commit, push, or switch branches.
- Your final answer is ONLY the `report` block below. No file dumps.

## Input
- Dataset: `internal/formatter/testdata/patterns/{{SLUG}}.json` (mode {{MODE}})

## Allowed file changes
- The JSON loader code in `internal/formatter/*_test.go`, ONLY if the loader itself is broken
  (does not find or parse a valid dataset). Never change any dataset, STYLE.md, or
  non-test code. If the dataset is invalid, report it; do not fix it.

## Steps
1. `go vet ./...` and `go test ./internal/formatter -run TestDatasets -v 2>&1 | tail -30`.
   TestDatasets validates dataset shape (class links, names, each case has `format_want`
   or `tokens_want`). TestFormat runs cases with `format_want`; TestTokenize runs cases
   with `tokens_want`, so a case can appear in one or both.
2. `go test ./internal/formatter -run '^(TestFormat|TestTokenize)$/^{{SLUG}}$' -v 2>&1 | grep -E '^\s*--- (FAIL|PASS)' | grep -F -e 'TestFormat/{{SLUG}}/' -e 'TestTokenize/{{SLUG}}/'`
   and collect the new cases that fail. Also run `go test ./... 2>&1 | tail -20` to check that
   no OLD case fails.
3. Confirm the new cases are loaded (they appear in the -v output). If not, the loader is
   broken: fix it in the test file, rerun.

## Report (required, exact keys)
```report
LOADER_OK: yes | fixed (<what>) | no (<why>)
DATASETS_TEST: pass | fail (<first error line>)
NEW_CASES_SEEN: <count>
NEW_FAILING: <count>
RED_CASES: <comma list of failing test names, max 20>
OLD_FAILING: <count, names if any>
NOTES: <one line, optional>
```
