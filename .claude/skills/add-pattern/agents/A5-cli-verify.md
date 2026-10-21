# A5 cli-verify - build and run the CLI, one case per class

You are a background agent for the `/add-pattern` skill. Repo: Go SQL formatter `sqlfmt`.

**Rules for you**
- Context limit: stay well below 120K tokens. Read only the parts of files you need.
- Work only in `{{WORKTREE}}`. Do not commit, push, or switch branches.
- Your final answer is ONLY the `report` block below. No file dumps.

## Input
- Dataset `internal/formatter/testdata/patterns/{{SLUG}}.json`
- Output file for the PR body: `{{STATE}}/cli-outputs.md`

## Allowed file changes
Only `{{STATE}}/cli-outputs.md` and the build output `bin/sqlfmt` (git-ignored). No source,
test, or data changes.

## Steps
1. `go build -o bin/sqlfmt ./cmd/sqlfmt`.
2. For each class id, take its first case that has `format_want` (tokens-only fragment
   cases cannot be checked with the CLI; skip a class that has only those and list it):
   `jq -c '. as $d | $d.classes[].id as $c | [$d.cases[] | select(.class==$c and .format_want)][0] // {class:$c, skip:true} | {class, name, sql, skip, want: ((.format_want // [])|join("\n"))}' <dataset>`
   (In expand mode, do only the classes from this run: {{NEW_CLASSES}}.)
3. Run `./bin/sqlfmt --sql "<sql>"` (quote safely, e.g. pass the SQL through a variable) and
   compare stdout (strip one trailing newline) with `want`. Record the exit code.
4. Write `{{STATE}}/cli-outputs.md`: for each class a `#### <class>` heading, the command, and
   the output in a ```sql block. Mark mismatches with "MISMATCH" and a short diff.

## Report (required, exact keys)
```report
BUILD: ok | fail (<error line>)
CLASSES_RUN: <count>
CLASSES_SKIPPED: <tokens-only class ids, or none>
MISMATCHES: <count>: <class/case: one-line diff; ...> or 0
OUTPUT_FILE: {{STATE}}/cli-outputs.md
```
