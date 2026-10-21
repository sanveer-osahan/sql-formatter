# A2 design - STYLE.md rules + dataset (equivalence class partitioning)

You are a background agent for the `/add-pattern` skill. Repo: Go SQL formatter `sqlfmt`.

**Rules for you**
- Context limit: stay well below 120K tokens. Read only the parts of files you need (use
  `grep`, `jq`, `sed -n`); never read whole large files.
- Work only in `{{WORKTREE}}`. Do not commit, push, or switch branches.
- Your final answer is ONLY the `report` block below. No file dumps.

## Input
- Pattern: `{{PATTERN}}`, slug `{{SLUG}}`, mode `{{MODE}}` (new | expand)
- Gaps from the scan (expand mode): {{GAPS}}
- Seed examples (inline SQL or file path, may be "none"): {{EXAMPLES}}
- Previous error to fix (may be empty): {{FAILURES}}

## Allowed file changes
- `STYLE.md` (repo root): add rules only; do not rewrite existing rules.
- `internal/formatter/testdata/patterns/{{SLUG}}.json`: create (new) or APPEND (expand).
  In expand mode existing classes and cases are frozen: do not change, reorder, or delete them.
- Nothing else. No Go code, no tests, no other datasets.

## Steps
1. Learn the conventions: `grep -n '^#' STYLE.md`, then read only the sections near the new
   construct. Look at one small existing dataset (`ls -S internal/formatter/testdata/patterns/`,
   pick the smallest) for the JSON shape.
2. Equivalence class partitioning. List the dimensions of the pattern, for example:
   clause position (SELECT list, WHERE, JOIN ON, ORDER BY...), nesting (inside a subquery,
   function, or itself), combination with other clauses, arity (1, 2, many branches/items),
   optional parts present/absent, line-length or wrapping behaviour, keyword case of input.
   Derive the classes: one class = one set of inputs the formatter must treat the same way.
3. Pick the MINIMUM cases: each class is covered by exactly one case unless a second is
   needed; one case may combine several non-conflicting dimensions, but it names exactly one
   class (its main one). Every class has at least one case. Map each seed example to a class;
   drop seeds that duplicate a class (list them in the report).
4. Write the STYLE.md rules FIRST: a new section for the construct (or extend the closest
   section) with clear, testable layout rules and one short example.
5. Then write the dataset, following STYLE.md exactly. Shape:
   `{"pattern": "{{SLUG}}", "classes": [{"id","desc"}], "cases": [{"name","class","sql","format_want":[lines],"tokens_want":[["Kind","TEXT"]]}]}`
   - Every case needs `format_want` OR `tokens_want` (at least one). TestFormat runs cases
     with `format_want`; TestTokenize runs cases with `tokens_want`.
   - Every full-query case MUST have `format_want`: an array of output lines (joined with "\n").
   - Tokens-only cases (no `format_want`) are allowed only for SQL fragments that test
     tokenization alone.
   - `tokens_want`: add it only to cases that bring a new token kind or new keyword, so each
     new token class is covered at least once. Use the existing kind names
     (`grep -n 'Kind\|const' internal/formatter/tokenize.go`).
   - Valid input only. No error cases (errors stay in inline Go tests).
   - Unique case names, lower case, short.
6. Check: `jq empty <file>`; every case class exists; every class has a case; every case has
   `format_want` or `tokens_want`.

## Report (required, exact keys)
```report
RULES_ADDED: <STYLE.md section/rule titles, semicolon list>
CLASSES: <count>: <comma list of class ids>
CASES: <count new> (total <count in file>)
SEEDS_DROPPED: <list or none>
TOKENS_WANT_CASES: <case names or none>
NOTES: <one line, optional>
```
