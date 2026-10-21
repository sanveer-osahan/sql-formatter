package formatter

import (
	"strings"
	"testing"
)

// formatCase is one input query and its expected formatted output.
type formatCase struct {
	category string
	name     string
	sql      string
	want     string
}

// formatCases is the dataset of SELECT query structures.
//
// Layout conventions:
//   - clause keywords are upper case and padded to 7 characters ("SELECT "),
//     longer ones such as GROUP BY are followed by one space
//   - SELECT columns go one per line after a comma
//   - top-level AND / OR in WHERE and HAVING go on a new line
//   - text inside parentheses stays on one line, except a subquery
//   - GROUP BY and ORDER BY lists stay on one line
//   - a subquery starts after its "(" and its lines line up with
//     the column after the "("; the ")" closes on its last line
//   - a CTE body is a subquery; each next CTE starts on a new line
//     under the first CTE name
//   - each JOIN starts a new line, indented 7 under FROM
//   - ON and USING start a new line, indented 9 under FROM
//   - AND / OR in ON start a new line, lined up with the text after "ON "
//   - FROM tables after a comma go one per line, like SELECT columns,
//     also after a JOIN
var formatCases = []formatCase{
	// ---- star and WHERE ----
	{
		category: "star_and_where",
		name:     "star without where",
		sql:      "select * from table_name",
		want: `SELECT *
FROM   table_name`,
	},
	{
		category: "star_and_where",
		name:     "star with where",
		sql:      "SELECT * FROM table_name WHERE id = 1",
		want: `SELECT *
FROM   table_name
WHERE  id = 1`,
	},
	{
		category: "star_and_where",
		name:     "fully qualified table",
		sql:      "select * from db.sales.orders where status = 'open'",
		want: `SELECT *
FROM   db.sales.orders
WHERE  status = 'open'`,
	},

	{
		category: "star_and_where",
		name:     "qualified star alias.*",
		sql:      "select u.*, o.id from users u join orders o on o.user_id = u.id",
		want: `SELECT u.*,
       o.id
FROM   users u
       JOIN orders o
         ON o.user_id = u.id`,
	},

	// ---- columns projection ----
	{
		category: "columns_projection",
		name:     "single column",
		sql:      "select id from users",
		want: `SELECT id
FROM   users`,
	},
	{
		category: "columns_projection",
		name:     "two columns",
		sql:      "SELECT col1, col2 FROM table_name WHERE id = 1",
		want: `SELECT col1,
       col2
FROM   table_name
WHERE  id = 1`,
	},
	{
		category: "columns_projection",
		name:     "aliases with and without AS",
		sql:      "select id, first_name as fname, last_name lname from users",
		want: `SELECT id,
       first_name AS fname,
       last_name lname
FROM   users`,
	},
	{
		category: "columns_projection",
		name:     "qualified columns and table alias",
		sql:      "select u.id,u.email from users u",
		want: `SELECT u.id,
       u.email
FROM   users u`,
	},
	{
		category: "columns_projection",
		name:     "distinct",
		sql:      "select distinct country from users",
		want: `SELECT DISTINCT country
FROM   users`,
	},
	{
		category: "columns_projection",
		name:     "arithmetic expression",
		sql:      "select price*quantity as total from order_items",
		want: `SELECT price * quantity AS total
FROM   order_items`,
	},

	// ---- aggregate functions ----
	{
		category: "agg_functions",
		name:     "count star",
		sql:      "select count(*) from users",
		want: `SELECT count(*)
FROM   users`,
	},
	{
		category: "agg_functions",
		name:     "several aggregates with aliases",
		sql:      "select COUNT(*) as total, SUM(amount) as revenue, avg(amount), min(created_at), max(created_at) from orders",
		want: `SELECT COUNT(*) AS total,
       SUM(amount) AS revenue,
       avg(amount),
       min(created_at),
       max(created_at)
FROM   orders`,
	},
	{
		category: "agg_functions",
		name:     "count distinct",
		sql:      "select count( distinct user_id ) from orders",
		want: `SELECT count(DISTINCT user_id)
FROM   orders`,
	},
	{
		category: "agg_functions",
		name:     "nested function with where",
		sql:      "select coalesce(sum(amount),0) as total from orders where status = 'paid'",
		want: `SELECT coalesce(sum(amount), 0) AS total
FROM   orders
WHERE  status = 'paid'`,
	},

	// ---- WHERE with AND / OR ----
	{
		category: "where_and",
		name:     "two conditions with AND, column name is not a keyword",
		sql:      "SELECT * FROM table_name WHERE id = 1 and name = 'ABC'",
		want: `SELECT *
FROM   table_name
WHERE  id = 1
       AND name = 'ABC'`,
	},
	{
		category: "where_and",
		name:     "three conditions with AND",
		sql:      "select * from orders where status='paid' and amount>100 and created_at >= '2026-01-01'",
		want: `SELECT *
FROM   orders
WHERE  status = 'paid'
       AND amount > 100
       AND created_at >= '2026-01-01'`,
	},
	{
		category: "where_and",
		name:     "OR",
		sql:      "select * from t where a = 1 or b = 2",
		want: `SELECT *
FROM   t
WHERE  a = 1
       OR b = 2`,
	},
	{
		category: "where_and",
		name:     "NOT with AND",
		sql:      "select * from users where not is_deleted and active = 1",
		want: `SELECT *
FROM   users
WHERE  NOT is_deleted
       AND active = 1`,
	},

	// ---- WHERE with complex AND / OR ----
	{
		category: "where_complex_and_or",
		name:     "group then AND",
		sql:      "select * from t where (a = 1 and b = 2 or c = 3) and d = 4",
		want: `SELECT *
FROM   t
WHERE  (a = 1 AND b = 2 OR c = 3)
       AND d = 4`,
	},
	{
		category: "where_complex_and_or",
		name:     "nested groups",
		sql:      "select * from t where ((a = 1 and b = 2) or c = 3) and d = 4",
		want: `SELECT *
FROM   t
WHERE  ((a = 1 AND b = 2) OR c = 3)
       AND d = 4`,
	},
	{
		category: "where_complex_and_or",
		name:     "AND then group",
		sql:      "select * from t where a = 1 and (b = 2 or c = 3)",
		want: `SELECT *
FROM   t
WHERE  a = 1
       AND (b = 2 OR c = 3)`,
	},
	{
		category: "where_complex_and_or",
		name:     "two groups and OR",
		sql:      "select * from t where ( a = 1 or b = 2 ) and ( c = 3 or d = 4 ) or e = 5",
		want: `SELECT *
FROM   t
WHERE  (a = 1 OR b = 2)
       AND (c = 3 OR d = 4)
       OR e = 5`,
	},
	{
		category: "where_complex_and_or",
		name:     "three levels of nesting",
		sql:      "select * from t where (((a = 1 or b = 2) and c = 3) or d = 4) and e = 5",
		want: `SELECT *
FROM   t
WHERE  (((a = 1 OR b = 2) AND c = 3) OR d = 4)
       AND e = 5`,
	},

	// ---- WHERE BETWEEN ----
	{
		category: "where_between",
		name:     "between numbers",
		sql:      "select * from orders where amount between 10 and 20",
		want: `SELECT *
FROM   orders
WHERE  amount BETWEEN 10 AND 20`,
	},
	{
		category: "where_between",
		name:     "between dates then AND",
		sql:      "select * from orders where created_at between '2026-01-01' and '2026-12-31' and status = 'paid'",
		want: `SELECT *
FROM   orders
WHERE  created_at BETWEEN '2026-01-01' AND '2026-12-31'
       AND status = 'paid'`,
	},
	{
		category: "where_between",
		name:     "not between",
		sql:      "select * from orders where amount not between 10 and 20",
		want: `SELECT *
FROM   orders
WHERE  amount NOT BETWEEN 10 AND 20`,
	},

	// ---- WHERE IN ----
	{
		category: "where_in",
		name:     "not in strings then AND",
		sql:      "select * from users where country not in ( 'US', 'CA' ) and active = 1",
		want: `SELECT *
FROM   users
WHERE  country NOT IN ('US', 'CA')
       AND active = 1`,
	},
	{
		category: "where_in",
		name:     "long list stays on one line",
		sql:      "select * from users where id in (101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115)",
		want: `SELECT *
FROM   users
WHERE  id IN (101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115)`,
	},

	// ---- WHERE NOT on a group ----
	{
		category: "where_not_group",
		name:     "not group",
		sql:      "select * from t where not (a = 1 or b = 2)",
		want: `SELECT *
FROM   t
WHERE  NOT (a = 1 OR b = 2)`,
	},
	{
		category: "where_not_group",
		name:     "AND not nested group",
		sql:      "select * from t where a = 1 and not ((b = 2 and c = 3) or d = 4)",
		want: `SELECT *
FROM   t
WHERE  a = 1
       AND NOT ((b = 2 AND c = 3) OR d = 4)`,
	},

	// ---- WHERE functions ----
	{
		category: "where_functions",
		name:     "function on column",
		sql:      "select * from users where lower(email) = 'a@example.com'",
		want: `SELECT *
FROM   users
WHERE  lower(email) = 'a@example.com'`,
	},
	{
		category: "where_functions",
		name:     "nested functions with AND",
		sql:      "select * from users where length( trim(last_name) ) > 0 and coalesce(age,0) >= 18",
		want: `SELECT *
FROM   users
WHERE  length(trim(last_name)) > 0
       AND coalesce(age, 0) >= 18`,
	},

	// ---- WHERE arithmetic ----
	{
		category: "where_arithmetic",
		name:     "star after a value is multiply",
		sql:      "select * from order_items where price*quantity > 100",
		want: `SELECT *
FROM   order_items
WHERE  price * quantity > 100`,
	},
	{
		category: "where_arithmetic",
		name:     "divide, add and modulo",
		sql:      "select * from orders where amount/2+10 >= total and id%2 = 0",
		want: `SELECT *
FROM   orders
WHERE  amount / 2 + 10 >= total
       AND id % 2 = 0`,
	},

	{
		category: "where_arithmetic",
		name:     "unary sign after multiply",
		sql:      "select a * -1 from t where b*+2 > 0",
		want: `SELECT a * -1
FROM   t
WHERE  b * +2 > 0`,
	},

	// ---- WHERE literals ----
	{
		category: "where_literals",
		name:     "unary minus and subtraction",
		sql:      "select * from accounts where balance > -5 and balance - 5 < 100",
		want: `SELECT *
FROM   accounts
WHERE  balance > -5
       AND balance - 5 < 100`,
	},

	// ---- GROUP BY ----
	{
		category: "group_by",
		name:     "single column",
		sql:      "select country, count(*) from users group by country",
		want: `SELECT country,
       count(*)
FROM   users
GROUP BY country`,
	},
	{
		category: "group_by",
		name:     "multiple columns with where",
		sql:      "select country, city, count(*) as n from users where active = 1 group by country, city",
		want: `SELECT country,
       city,
       count(*) AS n
FROM   users
WHERE  active = 1
GROUP BY country, city`,
	},
	{
		category: "group_by",
		name:     "positional columns",
		sql:      "select country, count(*) from users group   by 1",
		want: `SELECT country,
       count(*)
FROM   users
GROUP BY 1`,
	},

	// ---- HAVING ----
	{
		category: "having",
		name:     "single condition",
		sql:      "select country, count(*) as n from users group by country having count(*) > 10",
		want: `SELECT country,
       count(*) AS n
FROM   users
GROUP BY country
HAVING count(*) > 10`,
	},
	{
		category: "having",
		name:     "HAVING with AND",
		sql:      "select user_id, sum(amount) from orders group by user_id having count(*) > 10 and sum(amount) < 1000",
		want: `SELECT user_id,
       sum(amount)
FROM   orders
GROUP BY user_id
HAVING count(*) > 10
       AND sum(amount) < 1000`,
	},

	// ---- ORDER BY and LIMIT ----
	{
		category: "order_by_limit",
		name:     "order by desc",
		sql:      "select * from orders order by created_at desc",
		want: `SELECT *
FROM   orders
ORDER BY created_at DESC`,
	},
	{
		category: "order_by_limit",
		name:     "order by several columns with limit and offset",
		sql:      "select id, amount from orders order by amount desc, id asc limit 10 offset 20",
		want: `SELECT id,
       amount
FROM   orders
ORDER BY amount DESC, id ASC
LIMIT  10
OFFSET 20`,
	},
	{
		category: "order_by_limit",
		name:     "all clauses",
		sql:      "select country, count(*) as n from users where active = 1 and age >= 18 group by country having count(*) > 5 order by n desc limit 3",
		want: `SELECT country,
       count(*) AS n
FROM   users
WHERE  active = 1
       AND age >= 18
GROUP BY country
HAVING count(*) > 5
ORDER BY n DESC
LIMIT  3`,
	},

	// ---- subqueries ----
	{
		category: "subquery",
		name:     "IN subquery",
		sql:      "select * from users where id in (select user_id from orders where amount > 100)",
		want: `SELECT *
FROM   users
WHERE  id IN (SELECT user_id
              FROM   orders
              WHERE  amount > 100)`,
	},
	{
		category: "subquery",
		name:     "NOT IN subquery with AND inside",
		sql:      "select * from users where id not in (select user_id from orders where status = 'refunded' and amount > 100)",
		want: `SELECT *
FROM   users
WHERE  id NOT IN (SELECT user_id
                  FROM   orders
                  WHERE  status = 'refunded'
                         AND amount > 100)`,
	},
	{
		category: "subquery",
		name:     "scalar subquery in comparison",
		sql:      "select id, amount from orders where amount > (select avg(amount) from orders)",
		want: `SELECT id,
       amount
FROM   orders
WHERE  amount > (SELECT avg(amount)
                 FROM   orders)`,
	},
	{
		category: "subquery",
		name:     "correlated EXISTS",
		sql:      "select u.id from users u where exists (select 1 from orders o where o.user_id = u.id)",
		want: `SELECT u.id
FROM   users u
WHERE  EXISTS (SELECT 1
               FROM   orders o
               WHERE  o.user_id = u.id)`,
	},
	{
		category: "subquery",
		name:     "AND NOT EXISTS",
		sql:      "select u.id from users u where u.active = 1 and not exists (select 1 from orders o where o.user_id = u.id)",
		want: `SELECT u.id
FROM   users u
WHERE  u.active = 1
       AND NOT EXISTS (SELECT 1
                       FROM   orders o
                       WHERE  o.user_id = u.id)`,
	},
	{
		category: "subquery",
		name:     "derived table in FROM",
		sql:      "select t.user_id, t.total from (select user_id, sum(amount) as total from orders group by user_id) t where t.total > 1000",
		want: `SELECT t.user_id,
       t.total
FROM   (SELECT user_id,
               sum(amount) AS total
        FROM   orders
        GROUP BY user_id) t
WHERE  t.total > 1000`,
	},
	{
		category: "subquery",
		name:     "derived table with AS alias",
		sql:      "select * from (select id from users where active = 1) as active_users",
		want: `SELECT *
FROM   (SELECT id
        FROM   users
        WHERE  active = 1) AS active_users`,
	},
	{
		category: "subquery",
		name:     "scalar subquery in SELECT",
		sql:      "select u.id, (select count(*) from orders o where o.user_id = u.id) as order_count from users u",
		want: `SELECT u.id,
       (SELECT count(*)
        FROM   orders o
        WHERE  o.user_id = u.id) AS order_count
FROM   users u`,
	},
	{
		category: "subquery",
		name:     "subquery in HAVING",
		sql:      "select user_id, count(*) as n from orders group by user_id having count(*) > (select avg(cnt) from order_counts)",
		want: `SELECT user_id,
       count(*) AS n
FROM   orders
GROUP BY user_id
HAVING count(*) > (SELECT avg(cnt)
                   FROM   order_counts)`,
	},
	{
		category: "subquery",
		name:     "nested subqueries",
		sql:      "select * from users where id in (select user_id from orders where product_id in (select id from products where price > 100))",
		want: `SELECT *
FROM   users
WHERE  id IN (SELECT user_id
              FROM   orders
              WHERE  product_id IN (SELECT id
                                    FROM   products
                                    WHERE  price > 100))`,
	},

	{
		category: "subquery",
		name:     "subquery as function argument, inside parentheses",
		sql:      "select coalesce((select max(x) from t), 0) as m from u",
		want: `SELECT coalesce((SELECT max(x)
                 FROM   t), 0) AS m
FROM   u`,
	},

	// ---- CTEs ----
	{
		category: "cte",
		name:     "single CTE",
		sql:      "with paid_orders as (select user_id, amount from orders where status = 'paid') select user_id, sum(amount) as total from paid_orders group by user_id",
		want: `WITH paid_orders AS (SELECT user_id,
                            amount
                     FROM   orders
                     WHERE  status = 'paid')
SELECT user_id,
       sum(amount) AS total
FROM   paid_orders
GROUP BY user_id`,
	},
	{
		category: "cte",
		name:     "two CTEs, second uses first",
		sql:      "with daily as (select created_at, sum(amount) as total from orders group by created_at), ranked as (select created_at, total from daily where total > 100) select * from ranked order by total desc limit 10",
		want: `WITH daily AS (SELECT created_at,
                      sum(amount) AS total
               FROM   orders
               GROUP BY created_at),
     ranked AS (SELECT created_at,
                       total
                FROM   daily
                WHERE  total > 100)
SELECT *
FROM   ranked
ORDER BY total DESC
LIMIT  10`,
	},
	{
		category: "cte",
		name:     "CTE with column list",
		sql:      "with totals (user_id, total) as (select user_id, sum(amount) from orders group by user_id) select * from totals",
		want: `WITH totals (user_id, total) AS (SELECT user_id,
                                        sum(amount)
                                 FROM   orders
                                 GROUP BY user_id)
SELECT *
FROM   totals`,
	},
	{
		category: "cte",
		name:     "CTEs with subquery in main query",
		sql:      "with active_users as (select id from users where active = 1), user_totals as (select user_id, sum(amount) as total from orders group by user_id) select id from active_users where id in (select user_id from user_totals where total > 1000)",
		want: `WITH active_users AS (SELECT id
                      FROM   users
                      WHERE  active = 1),
     user_totals AS (SELECT user_id,
                            sum(amount) AS total
                     FROM   orders
                     GROUP BY user_id)
SELECT id
FROM   active_users
WHERE  id IN (SELECT user_id
              FROM   user_totals
              WHERE  total > 1000)`,
	},

	// ---- JOIN types ----
	{
		category: "join_types",
		name:     "any JOIN type with ON, multi-word LEFT OUTER JOIN",
		sql:      "select * from users u left outer join orders o on o.user_id = u.id",
		want: `SELECT *
FROM   users u
       LEFT OUTER JOIN orders o
         ON o.user_id = u.id`,
	},
	{
		category: "join_types",
		name:     "JOIN without ON or USING",
		sql:      "select * from sizes s cross join colors c",
		want: `SELECT *
FROM   sizes s
       CROSS JOIN colors c`,
	},
	{
		category: "join_types",
		name:     "USING instead of ON",
		sql:      "select * from users u join orders o using (user_id)",
		want: `SELECT *
FROM   users u
       JOIN orders o
         USING (user_id)`,
	},
	{
		category: "join_types",
		name:     "comma-joined tables one per line",
		sql:      "select * from users u, orders o where o.user_id = u.id",
		want: `SELECT *
FROM   users u,
       orders o
WHERE  o.user_id = u.id`,
	},
	{
		category: "join_types",
		name:     "comma-joined table after JOIN goes under FROM",
		sql:      "select * from t join b on a.id=b.id, c",
		want: `SELECT *
FROM   t
       JOIN b
         ON a.id = b.id,
       c`,
	},
	{
		category: "join_types",
		name:     "only listed JOIN phrases merge, an alias named left stays",
		sql:      "select * from t left left join b on t.id = b.id",
		want: `SELECT *
FROM   t left
       LEFT JOIN b
         ON t.id = b.id`,
	},

	// ---- JOIN conditions ----
	{
		category: "join_conditions",
		name:     "ON with AND",
		sql:      "select * from users u join orders o on o.user_id = u.id and o.status = 'paid'",
		want: `SELECT *
FROM   users u
       JOIN orders o
         ON o.user_id = u.id
            AND o.status = 'paid'`,
	},
	{
		category: "join_conditions",
		name:     "ON with AND and OR group",
		sql:      "select * from users u join orders o on o.user_id = u.id and (o.status = 'paid' or o.status = 'shipped')",
		want: `SELECT *
FROM   users u
       JOIN orders o
         ON o.user_id = u.id
            AND (o.status = 'paid' OR o.status = 'shipped')`,
	},
	{
		category: "join_conditions",
		name:     "ON with BETWEEN",
		sql:      "select * from events e join periods p on e.ts between p.start_ts and p.end_ts",
		want: `SELECT *
FROM   events e
       JOIN periods p
         ON e.ts BETWEEN p.start_ts AND p.end_ts`,
	},
	{
		category: "join_conditions",
		name:     "ON with AND then WHERE with AND",
		sql:      "select * from users u join orders o on o.user_id = u.id and o.amount > 0 where u.active = 1 and o.status = 'paid'",
		want: `SELECT *
FROM   users u
       JOIN orders o
         ON o.user_id = u.id
            AND o.amount > 0
WHERE  u.active = 1
       AND o.status = 'paid'`,
	},

	// ---- multiple JOINs ----
	{
		category: "join_multiple",
		name:     "three tables",
		sql:      "select u.id, p.amount from users u join orders o on o.user_id = u.id left join payments p on p.order_id = o.id",
		want: `SELECT u.id,
       p.amount
FROM   users u
       JOIN orders o
         ON o.user_id = u.id
       LEFT JOIN payments p
         ON p.order_id = o.id`,
	},
	{
		category: "join_multiple",
		name:     "self join",
		sql:      "select e.first_name, m.first_name as manager from employees e left join employees m on m.id = e.manager_id",
		want: `SELECT e.first_name,
       m.first_name AS manager
FROM   employees e
       LEFT JOIN employees m
         ON m.id = e.manager_id`,
	},
	{
		category: "join_multiple",
		name:     "all clauses",
		sql:      "select u.country, count(*) as n from users u join orders o on o.user_id = u.id where o.amount > 10 group by u.country having count(*) > 5 order by n desc limit 3",
		want: `SELECT u.country,
       count(*) AS n
FROM   users u
       JOIN orders o
         ON o.user_id = u.id
WHERE  o.amount > 10
GROUP BY u.country
HAVING count(*) > 5
ORDER BY n DESC
LIMIT  3`,
	},
	{
		category: "join_multiple",
		name:     "input whitespace and line breaks are ignored",
		sql: `SELECT u.id, o.amount
FROM users u
LEFT JOIN orders o ON o.user_id = u.id
WHERE u.active = 1`,
		want: `SELECT u.id,
       o.amount
FROM   users u
       LEFT JOIN orders o
         ON o.user_id = u.id
WHERE  u.active = 1`,
	},

	// ---- JOINs with subqueries and CTEs ----
	{
		category: "join_subquery",
		name:     "JOIN a derived table",
		sql:      "select u.id, t.total from users u join (select user_id, sum(amount) as total from orders group by user_id) t on t.user_id = u.id",
		want: `SELECT u.id,
       t.total
FROM   users u
       JOIN (SELECT user_id,
                    sum(amount) AS total
             FROM   orders
             GROUP BY user_id) t
         ON t.user_id = u.id`,
	},
	{
		category: "join_subquery",
		name:     "JOIN inside a subquery",
		sql:      "select * from users where id in (select o.user_id from orders o join payments p on p.order_id = o.id)",
		want: `SELECT *
FROM   users
WHERE  id IN (SELECT o.user_id
              FROM   orders o
                     JOIN payments p
                       ON p.order_id = o.id)`,
	},
	{
		category: "join_subquery",
		name:     "JOIN a CTE",
		sql:      "with paid as (select order_id from payments where status = 'ok') select o.id from orders o join paid p on p.order_id = o.id",
		want: `WITH paid AS (SELECT order_id
              FROM   payments
              WHERE  status = 'ok')
SELECT o.id
FROM   orders o
       JOIN paid p
         ON p.order_id = o.id`,
	},
}

func TestFormat(t *testing.T) {
	for _, tc := range formatCases {
		t.Run(tc.category+"/"+tc.name, func(t *testing.T) {
			got, err := Format(tc.sql)
			if err != nil {
				t.Fatalf("Format(%q) returned error: %v", tc.sql, err)
			}
			if got != tc.want {
				t.Errorf("Format(%q)\n got:\n%s\nwant:\n%s", tc.sql, got, tc.want)
			}
		})
	}
}

func TestFormatErrors(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		wantErr string
	}{
		{"unclosed ( outside a subquery", "select count(* from t where a = 1 and b = 2", `missing ")"`},
		{"stray )", "select a) from t", `unexpected ")"`},
		{"unclosed subquery", "select * from t where id in (select id from u", `missing ")"`},
		{"line comment", "select a from t -- c", "comments are not supported at position 16"},
		{"block comment", "select a /* c */ from t", "comments are not supported at position 9"},
		{"malformed number", "select 1.2.3 from t", "malformed number at position 7"},
		{"quoted identifier", `select "a b" from t`, `unexpected character '"' at position 7`},
		{"tokenizer error", "select 'abc", "unterminated string at position 7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.sql)
			if err == nil {
				t.Fatalf("Format(%q) = %q, want error %q", tt.sql, got, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Format(%q) error = %q, want %q", tt.sql, err, tt.wantErr)
			}
		})
	}
}
