# A6 docs - MANIFEST.md and README

You are a background agent for the `/add-pattern` skill. Repo: Go SQL formatter `sqlfmt`.

**Rules for you**
- Context limit: stay well below 120K tokens. Read only the parts of files you need.
- Work only in `{{WORKTREE}}`. Do not commit, push, or switch branches.
- Your final answer is ONLY the `report` block below. No file dumps.

## Input
- Pattern `{{PATTERN}}`, slug `{{SLUG}}`, mode `{{MODE}}` (new | expand | manifest-only)
- STYLE.md sections added: {{RULES_ADDED}}

## Allowed file changes
- `internal/formatter/testdata/patterns/MANIFEST.md`
- `README.md` (only the "Supported SQL" section), skipped in manifest-only mode.
Nothing else.

## Steps
1. Get the class ids: `jq -r '.classes[].id' internal/formatter/testdata/patterns/{{SLUG}}.json`.
2. MANIFEST.md: add or update the row `pattern | {{SLUG}}.json | <class ids> | <STYLE.md section>`.
   Remove the pattern from "Known uncovered" if listed there. Keep the table sorted as it is.
3. README "Supported SQL": add one short bullet for the pattern (or extend its bullet in
   expand mode), in the same style as the other bullets. Create the section only if missing.
4. Check that the manifest table has one row per JSON file:
   `ls internal/formatter/testdata/patterns/*.json | wc -l` vs the row count.

## Report (required, exact keys)
```report
MANIFEST: updated (<row change>) | unchanged
README: updated | unchanged | skipped
NOTES: <one line, optional>
```
