# Pattern coverage manifest

Each row is one supported query pattern. The dataset file is the source
of truth for its test cases: `TestFormat` runs every case, `TestTokenize`
runs the cases with `tokens_want`, and `TestDatasets` checks the file
shape. Layout rules are in [STYLE.md](../../../../STYLE.md).

Error cases (invalid input) are not in the datasets; they stay inline in
`TestFormatErrors` and `TestTokenizeErrors`.

| Pattern | Dataset file | Classes | STYLE.md section |
|---|---|---|---|
| select_basics | `select_basics.json` | star_and_where, columns_projection, agg_functions, tokens_basic, tokens_keywords, tokens_identifiers, tokens_functions | General; SELECT list |
| where | `where.json` | where_and, where_complex_and_or, where_not_group, where_between, where_in, where_literals, where_functions, where_arithmetic, tokens_literals, tokens_operators | WHERE and HAVING |
| group_having_order | `group_having_order.json` | group_by, having, order_by_limit | GROUP BY, ORDER BY and LIMIT; WHERE and HAVING |
| subquery | `subquery.json` | subquery, tokens_subqueries | Subqueries |
| cte | `cte.json` | cte, tokens_cte | CTEs |
| joins | `joins.json` | join_types, join_conditions, join_multiple, join_subquery, tokens_joins | JOINs |

## Known uncovered

- CASE expressions
- UNION / INTERSECT / EXCEPT
- Window functions (`OVER`)
- DML / DDL (`INSERT`, `UPDATE`, `DELETE`, `CREATE`)
- Comments (`--`, `/* */`): rejected with an error
- Quoted identifiers (`"name"`): rejected with an error
