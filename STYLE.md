# SQL layout style

These are the layout rules of `sqlfmt`. The pattern datasets in
`internal/formatter/testdata/patterns/*.json` follow them, and
`internal/formatter/testdata/patterns/MANIFEST.md` maps each pattern to
its section here. To support a new construct, add its rule here first,
then the dataset, then the code.

The layout of each clause keyword is one row of the `clauses` table in
`internal/formatter/formatter.go`.

## General

- Keywords are upper case. Identifiers, function names and string
  literals keep their original case.
- Each clause keyword starts a line, padded to its width.
- Clause keywords are padded to 7 characters, the width of `"SELECT "`.
  Longer ones, such as `GROUP BY` and `ORDER BY`, are followed by one space.
- Multi-word keywords (`GROUP BY`, `LEFT OUTER JOIN`, ...) are one token,
  also across whitespace and newlines.
- Whitespace and newlines in the input do not matter. Tokens on one line
  are separated by one space, also around operators (`id>=10` becomes
  `id >= 10`), with these exceptions:
  - no space after `(` or before `)`: `count( * )` becomes `count(*)`,
    `(a+b)` becomes `(a + b)`;
  - no space before `(` after a name (a function call): `left (a,2)`
    becomes `left(a, 2)`. After a keyword there is a space: `x IN (1, 2)`;
  - no space before a comma, and one space after it when it stays on the
    line;
  - no space after a unary sign: `y = - 3` becomes `y = -3`;
  - a qualified name such as `a.b` or `t.*` is one token, so it has no
    spaces around the `.`.
- Text inside parentheses stays on one line, except a subquery.

## SELECT list

- In a clause with commas (`SELECT`, `FROM`, `WITH`), each item after a
  top-level comma goes on its own line, lined up under the first item.
- Function calls and expressions in a column stay on that column's line.

## WHERE and HAVING

- In a clause with conditions (`WHERE`, `HAVING`, `ON`), each top-level
  `AND` / `OR` starts a new line, lined up under the first condition.
- `AND` / `OR` inside parentheses stay on the line.
- The `AND` of `BETWEEN .. AND ..` stays on the line.

## GROUP BY, ORDER BY and LIMIT

- `GROUP BY` and `ORDER BY` lists stay on one line.
- `LIMIT` and `OFFSET` each start a line, padded like other clauses.

## Subqueries

- A subquery starts right after its `(`. Its lines line up with the
  column after the `(`.
- The `)` closes on the subquery's last line.

## CTEs

- A CTE body is a subquery and follows the subquery rules.
- Each next CTE starts on a new line, lined up under the first CTE name.
- A column list after a CTE name, `t (a, b)`, stays on one line.

## JOINs

- Each `JOIN` starts a new line, indented 7 under `FROM`.
- `ON` and `USING` start a new line, indented 9 under `FROM`.
- `AND` / `OR` in `ON` start a new line, lined up with the text after `"ON "`.
- `FROM` tables after a comma go one per line, like `SELECT` columns,
  also after a `JOIN`.
- `LEFT` and `RIGHT` followed by `(` are function calls, not joins.

## INSERT

- `INSERT INTO` is one keyword token, also across whitespace and newlines.
  It starts a line and is followed by one space, like `GROUP BY`.
- The target table and its column list stay on the `INSERT INTO` line.
  The column list has one space before its `(`, like a CTE column list:
  `INSERT INTO t (a, b)`. It stays on one line.
- `VALUES` starts a line, padded to 7 like other clauses.
- Each row tuple after a top-level comma goes on its own line, lined up
  under the first tuple. The `,` stays at the end of the tuple's line.
- Text inside a row tuple stays on one line (General rule).
- `INSERT INTO ... SELECT`: the `SELECT` starts a new line at column 0 and
  the query follows the normal layout rules.
- `WITH ... INSERT INTO ...`: the CTEs follow the CTE rules, and
  `INSERT INTO` starts a new line at column 0.

```sql
INSERT INTO users (id, name)
VALUES (1, 'ann'),
       (2, 'bob')
```
