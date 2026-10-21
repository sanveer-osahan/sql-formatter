# sql-formatter

A CLI that formats a SQL query: it upper-cases keywords and lays the
query out one clause per line, with aligned columns, conditions, joins
and subqueries.

## Usage

```sh
make build
./bin/sqlfmt --sql "SELECT * FROM users"
```

Or run it without a build:

```sh
go run ./cmd/sqlfmt --sql "SELECT * FROM users"
```

## Supported SQL

`SELECT` queries with `WHERE`, `GROUP BY`, `HAVING`, `ORDER BY`, `LIMIT`,
`OFFSET`, joins, subqueries and CTEs.

`INSERT INTO ... VALUES` (one or many rows), `INSERT ... SELECT` and
`WITH ... INSERT`.

`UPDATE` queries with multi-column `SET`, `FROM`, `WHERE`, subqueries and CTEs.

- [internal/formatter/testdata/patterns/MANIFEST.md](internal/formatter/testdata/patterns/MANIFEST.md)
  lists each supported pattern, its test dataset and what is not covered yet.
- [STYLE.md](STYLE.md) holds the layout rules.

## Tests

```sh
go test ./...
```

Test cases live in `internal/formatter/testdata/patterns/*.json`.

## Sample queries

`samples/queries.sql` holds 2 or 3 sample queries for each pattern. To
build the CLI and format every sample:

```sh
scripts/run-samples.sh
```

The script prints each query and its formatted output, then a count of
queries and failures. It exits non-zero if any query fails. To use a
different file, pass its path: `scripts/run-samples.sh path/to/file.sql`.

In the samples file, a `-- pattern: <name>` line starts a group and a
blank line ends a query, so one query can span several lines.

## Adding a pattern

Use the `/add-pattern` skill in Claude Code to add a new query pattern
(for example `CASE`, `UNION` or window functions):

```
/add-pattern <pattern name> [-- example SQL | @path/to/examples.sql]
```

The skill runs end to end in a separate git worktree: coverage check,
test dataset, STYLE.md rules, implementation, CLI check, review and a
PR. It asks you to confirm at two decision points. After the PR merges,
add 2 or 3 queries for the new pattern to `samples/queries.sql`.
