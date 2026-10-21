---
name: add-pattern
description: Add support for a new SQL query pattern (for example CASE, UNION, window functions) to the sqlfmt formatter, end to end - coverage check, equivalence-class dataset, STYLE.md rules, implementation, CLI check, review, and a PR. Use when the user asks to add, cover, or support a SQL query pattern or construct in the formatter.
argument-hint: <pattern name> [-- example SQL | @path/to/examples.sql]
---

# /add-pattern

You are the **orchestrator**. Background agents do all the work. You keep a small state block,
run the diff guard, and talk to the user. Briefs for the agents are in `agents/` next to this
file; `scripts/guard.sh` is the diff guard; `templates/pr-body.md` is the PR body.

## Hard rules

1. Keep this context small (auto-compaction starts above 120K tokens). Do NOT read source
   files, test files, datasets, STYLE.md, README, or MANIFEST.md yourself. Read only the brief
   files in `agents/` and `templates/`, and agent reports.
2. Do all work through the Agent tool: `subagent_type: general-purpose` and the `model` from
   the phase table. Agents started with the Agent tool run in the background; wait for the
   completion notice and do not poll. To run independent agents in parallel, send their
   Agent calls in ONE message.
3. To start an agent: Read its brief file, replace every `{{PLACEHOLDER}}` with values from the
   state block, and pass the result as the `prompt`. Every brief already says: 120K limit,
   read only the needed parts of files, return only the short `report` block.
4. Keep only the agent's `report` block in state. Never ask an agent for file contents.
5. The only user interaction is AskUserQuestion at the decision points below (D6, D20).
   Everything else runs without stopping.
6. Never commit, push, or edit files in the main checkout. All agents work in `{{WORKTREE}}`.
7. The dataset is FROZEN after P3: `internal/formatter/testdata/patterns/*.json`,
   `*_test.go`, `STYLE.md`. Code changes go only to non-test `.go` files.

## State block

Keep this block up to date in your replies (one compact copy, replace the old one):

```
PATTERN=   SLUG=   MODE=new|expand   EXAMPLES=<inline text or @path or none>
REPO=   WT=   STATE=   GUARD=   OLD_N=   BRANCH=pattern/<slug>   (literal absolute paths)
P1=covered|partial|not_covered  GAPS=  SIMILAR=
CLASSES=  CASES=  RULES_ADDED=  RED_CASES=
P4_ITER=0/5  P7_ITER=0/5  VIOLATIONS=0  TESTS=green|red  LAST_ERROR=
CLI_MISMATCH=  DOCS=  A7_FINDINGS=H0 M0 L0  A8_FINDINGS=H0 M0 L0  PR=
```

## Placeholders

Fill these in briefs (use `none` for an empty value):
`{{WORKTREE}}` = `<WT>`, `{{STATE}}` = `<STATE>`, `{{PATTERN}}`, `{{SLUG}}`, `{{MODE}}`, `{{EXAMPLES}}`,
`{{GAPS}}`, `{{ITERATION}}`, `{{FAILURES}}`, `{{FINDINGS}}`, `{{RULES_ADDED}}` (from A2),
`{{NEW_CLASSES}}` (class ids from A2; all classes in new mode), `{{DRAFT}}`,
`{{PR_TITLE}}` = `Add <pattern> formatting` (prefix `WIP: ` when draft),
`{{COMMIT_SUBJECT}}` = `Add <pattern> pattern to formatter`.
In A9 leave `{{CLI_OUTPUTS}}` literally as is.
The class table rows for the PR body come from a one-line command you run:
`jq -r '.classes[].id as $c | "| \($c) | \([.cases[]|select(.class==$c)|.name]|join(", ")) |"' "<WT>/internal/formatter/testdata/patterns/<slug>.json"`
(same command in new and expand mode).

## Input (D3)

Parse `$ARGUMENTS`:
- Text before ` -- ` (or all of it) is the pattern name. Text after ` -- ` is inline example SQL.
- An argument that starts with `@` or is an existing file path is an examples file. Do not read
  it; pass the absolute path to A1 and A2 as `{{EXAMPLES}}`.
- No pattern name: ask the user once with AskUserQuestion, then continue.
- Slug: lower-case the name, turn every run of non-alphanumeric characters into `_`, trim `_`.
  "window functions" -> `window_functions`.

## Phases and models (D17, D18)

| Phase | Agent brief | model | Runs |
|---|---|---|---|
| P1 ground | `agents/A1-scan.md` | sonnet | alone |
| P2 design | `agents/A2-design.md` | opus | alone |
| P3 tests | `agents/A3-tests.md` | sonnet | alone |
| P4 implement | `agents/A4-implement.md` | opus | loop, max 5 |
| P5 verify | `agents/A5-cli-verify.md` + `agents/A6-docs.md` | sonnet + sonnet | parallel |
| P6 review | `agents/A7-dataset-review.md` + `agents/A8-code-review.md` | opus + opus | parallel |
| P7 fix | A4, then P5 again | as above | only if A8 has High/Medium |
| P8 ship | `agents/A9-ship.md` | haiku | alone |

### Paths in commands

Shell variables do NOT persist between Bash tool calls. Commands in this file use the
placeholders `<REPO>`, `<WT>`, `<STATE>`, `<GUARD>`, `<slug>`, `<OLD_N>`. Before running a
command, replace every placeholder with the literal value from the state block. Never run a
command that still contains `<...>` or relies on a variable set in an earlier call.

### P0 setup (you, Bash)

1. Run `git rev-parse --show-toplevel`. Record the output as `<REPO>` in the state block.
   Record `<WT>` = `<parent dir of REPO>/sql-formatter-wt-<slug>` and
   `<GUARD>` = `<REPO>/.claude/skills/add-pattern/scripts/guard.sh`.
2. Create the worktree (one call, literal paths):
   ```bash
   git -C "<REPO>" fetch origin && git -C "<REPO>" worktree add "<WT>" -b pattern/<slug> origin/main
   ```
   If the branch or path already exists, stop and tell the user (do not delete their work).
3. Precondition: run
   `test -f "<WT>/internal/formatter/testdata/patterns/MANIFEST.md" && test -f "<WT>/STYLE.md" && echo PREP_OK`.
   If it does not print `PREP_OK`, run Cleanup for this slug and stop with
   "Prep PR not merged into main; merge it first."
4. Run `git -C "<WT>" rev-parse --absolute-git-dir`. Record `<STATE>` = `<output>/add-pattern`
   and run `mkdir -p "<STATE>"`.
`<STATE>` lives in the worktree's private git dir: never committed, removed with the worktree.

### P1 ground (A1)

Start A1. Its report gives `STATUS`, `SLUG` (the existing dataset slug if covered), `GAPS`,
`SIMILAR` (3 closest uncovered patterns), `MANIFEST_DRIFT_FIXED`.
- `not_covered` -> MODE=new, go to P2.
- `covered` or `partial` -> tell the user in 2-4 lines what is covered and the gap list, then
  run **Decision D6**.

**Decision D6** (AskUserQuestion, one question, three options):
1. "Expand" - add the missing classes to the existing dataset. Set MODE=expand and
   SLUG=<existing slug>. If the slug changed, redo P0 with the new slug (remove the old
   worktree and its local branch first, see Cleanup). Go to P2.
2. "Do nothing" - run Cleanup, report, stop.
3. "Cover a similar pattern" - run **Decision D20**.

**Decision D20**: AskUserQuestion with the 3 `SIMILAR` patterns as options (the user can type
their own with "Other"). Then run Cleanup for the current slug and restart at P0 with the new
pattern name (new slug, new worktree), with no examples.

### P2 design (A2)

Expand mode only, BEFORE starting A2, record the existing dataset (existing cases are frozen
from the start):

```bash
jq -c '.cases' "<WT>/internal/formatter/testdata/patterns/<slug>.json" > "<STATE>/old-cases.json" && jq '.cases|length' "<WT>/internal/formatter/testdata/patterns/<slug>.json"
```
Record the printed number as `<OLD_N>`.

Start A2 (pass MODE, GAPS, EXAMPLES). Report: `RULES_ADDED`, `CLASSES`, `CASES`, `SEEDS_DROPPED`.
Expand mode only, after A2: `jq -c ".cases[:<OLD_N>]" "<WT>/internal/formatter/testdata/patterns/<slug>.json" | cmp - "<STATE>/old-cases.json"`.
If it differs, restore with `git -C "<WT>" checkout -- internal/formatter/testdata/patterns/<slug>.json` and start A2 again once
with "existing cases were changed; only append". A second failure: stop and report.

### P3 tests (A3)

Start A3. Report: `LOADER_OK`, `DATASETS_TEST`, `NEW_FAILING`, `RED_CASES`.
- `NEW_FAILING >= 1` (RED) -> freeze, then P4:
  `"<GUARD>" freeze "<WT>" "<STATE>"`
- `NEW_FAILING = 0` (all new cases already pass, Defaults RED rule): the pattern is already
  covered. Start A6 with `{{MODE}}=manifest-only` to record it in MANIFEST.md, then run
  **Decision D6** with GAPS="none found". "Expand" goes back to P2 with MODE=expand.
- `DATASETS_TEST=fail` (invalid dataset) -> start A2 once more with the A3 error as
  `{{FAILURES}}`, then rerun P3. A second failure: stop and report.

### P4 implement (A4 loop, cap 5)

Each iteration:
1. Increase `P4_ITER`. Start A4 with `{{ITERATION}}`, `{{FAILURES}}` (failing case names + last
   error from the previous iteration or from A3), `{{FINDINGS}}` (empty in P4).
2. Run `"<GUARD>" check "<WT>" "<STATE>"`. Exit 3 = A4 touched a frozen or non-code file: the guard
   already reverted it. Add 1 to VIOLATIONS and put the reverted paths into the next
   `{{FAILURES}}` as "you changed frozen files; do not". The iteration still counts.
3. If the guard was clean and A4 reports `TESTS: pass`, go to P5. Otherwise loop.

After 5 iterations still red (D8): set TESTS=red, skip P5-P7, go to P8 (draft PR).

### P5 verify (A5 || A6, parallel)

Start A5 and A6 in one message. A5 writes CLI outputs to `<STATE>/cli-outputs.md` and reports
`MISMATCHES`. A6 reports `MANIFEST`, `README`.
A5 mismatch (CLI output differs from `format_want` while tests pass): treat each as a High code
finding for P7.

### P6 review (A7 || A8, parallel)

Before starting, run `"<GUARD>" verify "<WT>" "<STATE>"` (must print `FROZEN: OK`; else revert with
`git -C "<WT>" checkout -- <path>` and note it). Start A7 and A8 in one message. Both report
findings as `<SEV> | <where> | <finding> | <fix>` lines with SEV in High/Medium/Low.
- A7 findings (any severity) never change files; they all go into the PR body as notes (D7).
- A8 Low findings go into the PR body.

### P7 fix (only if A8 or A5 gave at least one High or Medium)

Refresh the snapshot so A6's doc edits are not flagged: `"<GUARD>" freeze "<WT>" "<STATE>"`
(frozen hashes are kept from P3). Run the P4 loop again with `{{FINDINGS}}` = the High and
Medium findings, counted in `P7_ITER` (new cap of 5). Then run P5 again. No second review.
Still red after 5 -> TESTS=red, go to P8 (draft).

### P8 ship (A9)

1. Run `"<GUARD>" verify "<WT>" "<STATE>"` one last time.
2. Read `templates/pr-body.md`, fill it from the state block and reports, and Write it to
   `<STATE>/pr-body.md`. Insert `<STATE>/cli-outputs.md` content by reference: write the line
   `{{CLI_OUTPUTS}}` as is; A9 replaces it with the file contents.
3. Start A9 with `{{DRAFT}}` = `yes` if TESTS=red, else `no`. Report: `COMMIT`, `PUSHED`, `PR`.

### Cleanup (D24)

After A9 reports a PR URL (or on "Do nothing" / restart):

```bash
git -C "<REPO>" worktree remove --force "<WT>"
git -C "<REPO>" branch -D pattern/<slug>    # only when NO PR was created (branch was never pushed)
```

After a PR, keep the remote branch; deleting the local branch is fine since it is pushed.

## Final report to the user

```
PR: <url> (ready | draft)
Pattern: <name> (<slug>), mode new|expand
Dataset: <N> classes, <M> cases
STYLE.md rules added: <list of rule titles>
Review: dataset H<a> M<b> L<c>; code H<d> M<e> L<f> (High/Medium code fixed: yes|no)
Iterations: P4 <x>/5, P7 <y>/5, guard violations <v>
Failing cases (draft only): <names>; last error: <one line>
```

## Failure rules

- An agent returns no `report` block or crashes: start it once more with the same brief. A
  second failure: stop, keep the worktree, and tell the user the worktree path and the phase.
- `git worktree add`, `git push`, or `gh pr create` fails: stop and report the error line.
