# sql-formatter

A CLI that prints the SQL text it receives.

## Usage

```sh
make build
./bin/sqlfmt --sql "SELECT * FROM users"
```

Or run it without a build:

```sh
go run ./cmd/sqlfmt --sql "SELECT * FROM users"
```
