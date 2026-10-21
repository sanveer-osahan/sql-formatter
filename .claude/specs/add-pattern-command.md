# Spec: Claude command that adds a query pattern to the SQL formatter

Status: CONFIRMED (2026-10-21). Implemented on branch chore/add-pattern-prep.

## Original requirements (from user)

The command does the following:

1. Grounding: check the codebase and collect all patterns that are already covered.
2. Check if the requested pattern is already covered.
3. If already covered: tell the user and ask what to do next. Options: expand the cases in the pattern, do nothing, cover a similar pattern.
4. If not covered: create a dataset of all cases to cover for the pattern.
5. The dataset follows equivalence class partitioning: minimum cases, maximum pattern coverage.
6. Write tests in formatter and tokenizer for the cases.
7. Implement the changes so all scenarios pass.
8. Run tests and iterate until all pass, without changes to the test datasets.
9. Build and run some cases with CLI commands and check the output.
10. Review dataset and code: equivalence class partitioning (minimal dataset, maximal scenarios); code does not hard-code scenarios to fit tests. Categorize findings as high / medium / low.
11. Implement review changes (repeat steps 7-9). If no major suggestions, skip this step.
12. Commit changes to a branch and raise a PR.

Constraints:
- The command starts background agents so every step stays under a 120K token context window.
- Agents run in parallel where possible.

## Repo facts (from grounding agent, 2026-10-21)

- Go module `github.com/sanveer-osahan/sql-formatter`, go 1.23.
- CLI: `cmd/sqlfmt/main.go`. Input only through `--sql "<text>"`. Empty -> usage, exit 2. Format error -> exit 1.
- One package `internal/formatter/`: `tokenize.go`, `tokenize_test.go`, `formatter.go`, `formatter_test.go`.
- Tests are table-driven, inline Go slices. No testdata/, no golden files.
  - `formatter_test.go`: `var formatCases []formatCase` {category, name, sql, want (raw backtick string)}, run by `TestFormat`, subtest `category/name`. Groups separated by `// ---- <group> ----` comments. Layout conventions in doc comment above `formatCases` (lines ~16-34). Errors: `TestFormatErrors` {name, sql, wantErr} with strings.Contains.
  - `tokenize_test.go`: `var tokenizeCases []tokenizeCase` {category, name, sql, want []Token}, run by `TestTokenize`; entries like `{Keyword, "SELECT"}`. Errors: `TestTokenizeErrors`.
- Clause table: `var clauses = []clause{...}` in `formatter.go` (~lines 30-51), "To add a clause, add a row here." Fields: keyword, width, indent, commas, conditions, startsQuery, nested, join, afterJoin. Multi-word keywords merged through `phrases`.
- Covered (formatter categories): star_and_where, columns_projection, agg_functions, where_and, where_complex_and_or, where_not_group, where_between, where_in, where_literals, where_functions, where_arithmetic, group_by, having, order_by_limit, subquery, cte, join_types, join_conditions, join_multiple, join_subquery.
- Tokenizer categories: basic, keywords, identifiers, literals, operators, functions, joins, subqueries.
- Not covered: CASE, UNION/INTERSECT/EXCEPT, window functions (OVER), DML/DDL, comments (error), quoted identifiers (error).
- Makefile: `build` (-> bin/sqlfmt), `run` (ARGS). No test target: use `go test ./...`. `bin/sqlfmt` is committed.
- No `.claude/` content before this spec. No repo CLAUDE.md. No CI.
- Remote origin `https://github.com/sanveer-osahan/sql-formatter.git`. `gh` authenticated (repo, workflow scopes).
- README is minimal; formatting rules live only in code comments.

## Decisions (grilling log)

### Round 1
- D1 Form: project skill at `.claude/skills/add-pattern/SKILL.md` (invoked as `/add-pattern`). Helper files (agent briefs, templates) may live beside it.
- D2 Orchestration: main context uses the Agent tool to start background agents step by step (NOT a Workflow script). Main context keeps only short agent reports.
- D3 Input: pattern name + optional example SQL, given inline in the arguments OR in a file whose path is in the arguments.
- D4 Checkpoints: none (see Round 2 for the partial-coverage branch). Fully autonomous. The only user question is the "already covered" branch (expand / do nothing / cover similar pattern).

### Round 2
- D5 Coverage source: manifest + scan. A coverage manifest (pattern -> test categories -> equivalence classes) lives with the skill. Each run, an agent scans test categories + clause table + tokenizer and fixes drift in the manifest. Each PR updates the manifest.
- D6 Partial coverage: treat as covered. Show what is covered and the missing classes (gap list). Ask: expand with the missing classes / do nothing / cover a similar pattern.
- D7 Review: findings are High / Medium / Low. The fix round runs only if at least one High or Medium finding exists. Apply High + Medium. List Low in the PR body. The dataset is FROZEN for the whole run: review may only change code. Dataset findings (any severity) go into the PR body as notes.
- D8 Loop cap: max 5 implement/test iterations per implement phase. If tests still fail: commit, raise a DRAFT PR that lists the failing cases and the last error, and report to the user.

### Round 3
- D9 Style: a style guide file (STYLE.md) holds all layout rules. For a new construct, the agent writes the new rule in STYLE.md FIRST; the dataset and tests follow it. (Bootstrap and relation to the doc comment: see Round 4.)
- D10 Dataset: a separate dataset file per pattern is the source of truth; tests load it. This changes the current inline-table convention. (Format, location, migration: see Round 4.)
- D11 Git: create a git worktree from latest origin/main on branch `pattern/<slug>`. Main checkout stays untouched. PR base `main`. Ready (non-draft) PR when green; draft PR per D8 when not.
- D12 Extra files per run: coverage manifest (D5), conventions doc comment, README "Supported SQL" section. `bin/sqlfmt` must NOT be committed: remove it from git (timing: see Round 4).

### Round 4
- D13 Dataset format: JSON at `internal/formatter/testdata/patterns/<slug>.json`, stdlib `encoding/json`. One file feeds both TestFormat and TestTokenize. Shape:
  ```json
  {
    "pattern": "case_expression",
    "classes": [{"id": "simple_case", "desc": "CASE x WHEN .."}],
    "cases": [{
      "name": "simple case in select",
      "class": "simple_case",
      "sql": "select case a when 1 then 'x' end from t",
      "format_want": ["SELECT CASE a", "  ..."],
      "tokens_want": [["Keyword", "SELECT"]]
    }]
  }
  ```
  `format_want` is an array of lines (joined with "\n"). Every case names exactly one class; every class has at least one case.
- D14 Migration: a one-time prep moves ALL existing inline cases (formatCases, tokenizeCases) into per-pattern JSON dataset files and builds the first coverage manifest. After prep, there is one convention only; the scan reads testdata only.
- D15 STYLE.md: prep extracts the layout rules from the doc comment above formatCases and the Format() doc comment into `STYLE.md` at repo root. Those doc comments shrink to "See STYLE.md". Each run adds rules ONLY to STYLE.md. So D12 "conventions doc comment" now means STYLE.md.
- D16 Prep timing: all one-time prep (untrack `bin/sqlfmt` + add `bin/` to .gitignore, STYLE.md bootstrap, dataset migration, first manifest, test loaders) is done ONCE in a separate prep PR together with the skill files. Each /add-pattern run assumes prep is in place.

### Round 5
- D17 Pipeline (main context only starts agents, keeps short reports, runs the diff guard):
  ```
  P0 setup (main): worktree pattern/<slug> from origin/main
  P1 ground   : [A1 manifest-scan]  (1 agent)
      -> covered/partial? ask user (D6, D20) / continue
  P2 design   : [A2 STYLE rules + dataset JSON]
  P3 tests    : [A3 test-loader check, run, expect RED]
      main: hash frozen files (freeze)
  P4 implement: [A4 implement] loop <= 5, main runs diff guard after each iteration
  P5 verify   : [A5 go build + CLI one case per class] || [A6 docs: README, manifest]   (parallel)
  P6 review   : [A7 dataset ECP review] || [A8 code anti-hardcode review]              (parallel)
  P7 fix      : if any High/Medium code finding -> P4 + P5 again (loop cap D8 applies again)
  P8 ship     : [A9 commit, push, gh pr create]
  ```
- D18 Models: Opus for A2 (design), A4 (implement), A7 + A8 (reviewers). Sonnet for A1 (scan), A5 (CLI verify), A6 (docs). Haiku for A9 (ship). A3 (tests): Sonnet.
- D19 Error cases: NOT in datasets. Datasets hold valid input only. TestFormatErrors / TestTokenizeErrors stay inline Go tables and are not migrated.
- D20 "Cover a similar pattern": A1 proposes the 3 closest uncovered patterns. The skill asks with AskUserQuestion (user may type their own). Then restart from P1 with the chosen pattern.

### Round 6
- D21 Manifest: `internal/formatter/testdata/patterns/MANIFEST.md`. Table: pattern | dataset file | classes | STYLE.md section. Plus a "Known uncovered" list (seed: CASE, UNION/INTERSECT/EXCEPT, window functions, DML/DDL, comments, quoted identifiers). A1 diffs it against the JSON files and fixes drift.
- D22 tokens_want: optional per case. format_want is required. Put tokens_want only on cases that bring new token kinds/keywords; each new token class is covered at least once. A7 checks this rule.
- D23 Freeze scope after P3: `internal/formatter/testdata/patterns/*.json`, `*_test.go`, `STYLE.md`. A4 and the A8-driven fixes may only change non-test `.go` files. The main context reverts a violation; the violation counts as one iteration.
- D24 Finish: after the PR is created, remove the worktree (branch stays on remote). Report: PR URL, class/case counts, STYLE rules added, review findings by severity, iterations used.

### Defaults (not asked; change if wrong)
- Slug: pattern name -> lower snake_case (e.g. "window functions" -> `window_functions`). Branch `pattern/<slug>`, dataset `<slug>.json`.
- "Expand" branch: append new classes/cases to the existing `<slug>.json`; existing cases are frozen from the start.
- "Do nothing" branch: remove the worktree and stop.
- P3 RED rule: at least one new case must fail before P4. If all pass, the pattern is in fact covered: update the manifest and go to the D6 question.
- Examples from the user (D3) are seeds for A2. Each seed maps to a class; A2 may drop duplicates of a class.
- PR body: summary, class table (class -> cases), STYLE.md rules added, CLI outputs (one per class), review findings (Low code findings + all dataset notes), iterations used. Ends with the Claude Code attribution line.
- Commit messages end with the Co-Authored-By trailer from the session.
- Every agent brief states the 120K context limit, reads only needed file parts, and returns a short report (no file dumps).

## Implementation plan (for a fresh session)
1. Prep PR (branch `chore/add-pattern-prep`): untrack `bin/sqlfmt`, add `bin/` to .gitignore; create STYLE.md from the doc comments (shrink them to "See STYLE.md"); add JSON loaders to TestFormat/TestTokenize; migrate all inline valid cases to `testdata/patterns/<slug>.json` (one file per current category group); write MANIFEST.md; keep error tests inline; `go test ./...` green with identical case count.
2. Same PR: write `.claude/skills/add-pattern/SKILL.md` (pipeline D17) plus agent briefs as helper files (`agents/A1-scan.md` ... `agents/A9-ship.md`).
3. Smoke-test the skill on one uncovered pattern (e.g. CASE) in a later run.

### Round 7 (during implementation, 2026-10-21)
- D25 (changes D13/D22): a case needs `format_want` OR `tokens_want` (at least one). Tokenizer-only cases (SQL fragments) carry only `tokens_want`. TestFormat runs only cases with `format_want`; TestTokenize runs only cases with `tokens_want`. Reason: pinning Format() output for fragments would freeze layout quirks (e.g. trailing spaces).
- Fact correction: `bin/sqlfmt` was never tracked and `bin/` is already in .gitignore. No untrack step is needed.
