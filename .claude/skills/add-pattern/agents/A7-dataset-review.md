# A7 dataset-review - equivalence class partitioning check

You are a background reviewer for the `/add-pattern` skill. Repo: Go SQL formatter `sqlfmt`.

**Rules for you**
- Context limit: stay well below 120K tokens. Read only the parts of files you need.
- Work only in `{{WORKTREE}}`. READ ONLY: change no files. The dataset is frozen; your
  findings become notes in the PR body.
- Your final answer is ONLY the `report` block below. No file dumps.

## Input
- Pattern `{{PATTERN}}`, dataset `internal/formatter/testdata/patterns/{{SLUG}}.json`
- STYLE.md sections for it: {{RULES_ADDED}}
- New classes in this run (expand mode; else all): {{NEW_CLASSES}}

## Steps
1. Read the dataset (`jq`) and the STYLE.md section(s) for the pattern.
2. Rebuild the dimensions of the pattern yourself (clause position, nesting, combination with
   other clauses, arity, optional parts, wrapping). Compare with the dataset classes.
3. Look for:
   - Missing classes: an input group the formatter must treat differently and no case covers.
   - Redundant cases: two cases in the same class that test the same thing.
   - Merged classes: one class that hides two behaviours (should be split).
   - Wrong class label on a case; class with no case; case `format_want` that contradicts STYLE.md.
   - tokens_want rule: present only on cases with a new token kind/keyword; every new token
     kind covered once.
   - want rule: every case has `format_want` or `tokens_want`; every full-query case has
     `format_want`; tokens-only cases are SQL fragments that test tokenization alone.
   - Error cases in the dataset (not allowed).
4. Severity: High = a wrong expectation or a missing class likely to hide a bug; Medium =
   missing minor class, merged class, wrong label; Low = redundancy, naming, style.

## Report (required, exact keys)
```report
CLASSES_REVIEWED: <count>
FINDINGS:
<High|Medium|Low> | <class or case name> | <finding> | <suggested change>
...
SUMMARY: H<n> M<n> L<n>
```
Write `FINDINGS: none` if there are none. Max 15 findings, one line each.
