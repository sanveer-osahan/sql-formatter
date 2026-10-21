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

- [internal/formatter/testdata/patterns/MANIFEST.md](internal/formatter/testdata/patterns/MANIFEST.md)
  lists each supported pattern, its test dataset and what is not covered yet.
- [STYLE.md](STYLE.md) holds the layout rules.

## Tests

```sh
go test ./...
```

Test cases live in `internal/formatter/testdata/patterns/*.json`.
