# A9 ship - commit, push, open the PR

You are a background agent for the `/add-pattern` skill. Repo: Go SQL formatter `sqlfmt`.

**Rules for you**
- Context limit: stay well below 120K tokens. Do not read source files.
- Work only in `{{WORKTREE}}` on branch `pattern/{{SLUG}}`. Do not switch branches. Do not
  change any file content.
- Your final answer is ONLY the `report` block below.

## Input
- Draft PR: {{DRAFT}} (yes = tests still red)
- PR title: `{{PR_TITLE}}`
- PR body file: `{{STATE}}/pr-body.md` (contains a line `{{CLI_OUTPUTS}}` to replace)
- CLI outputs file: `{{STATE}}/cli-outputs.md` (may not exist in a draft run)

## Steps
1. Build the final body: replace the line `{{CLI_OUTPUTS}}` in `{{STATE}}/pr-body.md` with the
   contents of `{{STATE}}/cli-outputs.md` (or with "Not run (tests red)." if missing), e.g.
   with `awk` writing to `{{STATE}}/pr-body.final.md`.
2. Stage: `git -C {{WORKTREE}} add -A` then make sure nothing under `bin/` is staged:
   `git -C {{WORKTREE}} reset -q -- bin/ 2>/dev/null || true`;
   `git -C {{WORKTREE}} diff --cached --name-only` must not list `bin/` (never commit bin/).
3. Commit with a message file (`git commit -F`):
   ```
   {{COMMIT_SUBJECT}}

   <2-4 lines: classes added, STYLE.md rules, draft reason if any>

   Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>
   ```
4. `git -C {{WORKTREE}} push -u origin pattern/{{SLUG}}`
5. From inside `{{WORKTREE}}`:
   `gh pr create --base main --head pattern/{{SLUG}} --title "{{PR_TITLE}}" --body-file {{STATE}}/pr-body.final.md`
   and add `--draft` if Draft PR is yes. The body already ends with the attribution line.

## Report (required, exact keys)
```report
COMMIT: <short sha> | fail (<error>)
FILES: <count committed>
PUSHED: yes | no (<error>)
PR: <url> | fail (<error>)
DRAFT: yes | no
```
