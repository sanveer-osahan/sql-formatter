package formatter

import (
	"slices"
	"strings"
	"testing"
)

// tokenizeCase is one input and its expected tokens.
type tokenizeCase struct {
	category string
	name     string
	sql      string
	want     []Token
}

// tokenizeCases is the dataset of token patterns.
var tokenizeCases = []tokenizeCase{
	// ---- basic query ----
	{
		category: "basic",
		name:     "select star with where",
		sql:      "SELECT * FROM table_name WHERE id = 1",
		want: []Token{
			{Keyword, "SELECT"}, {Star, "*"}, {Keyword, "FROM"}, {Identifier, "table_name"},
			{Keyword, "WHERE"}, {Identifier, "id"}, {Operator, "="}, {Number, "1"},
		},
	},
	{
		category: "basic",
		name:     "empty input",
		sql:      "",
		want:     nil,
	},
	{
		category: "basic",
		name:     "whitespace only",
		sql:      " \n\t ",
		want:     nil,
	},
	{
		category: "basic",
		name:     "extra whitespace and newlines",
		sql:      "  select\n\t*\n  from   users  ",
		want: []Token{
			{Keyword, "SELECT"}, {Star, "*"}, {Keyword, "FROM"}, {Identifier, "users"},
		},
	},

	// ---- keywords ----
	{
		category: "keywords",
		name:     "lower and mixed case become upper case",
		sql:      "select DISTINCT a From t wHeRe b is not null",
		want: []Token{
			{Keyword, "SELECT"}, {Keyword, "DISTINCT"}, {Identifier, "a"}, {Keyword, "FROM"},
			{Identifier, "t"}, {Keyword, "WHERE"}, {Identifier, "b"}, {Keyword, "IS"},
			{Keyword, "NOT"}, {Keyword, "NULL"},
		},
	},
	{
		category: "keywords",
		name:     "GROUP BY and ORDER BY are one token each",
		sql:      "group by a order by b desc",
		want: []Token{
			{Keyword, "GROUP BY"}, {Identifier, "a"}, {Keyword, "ORDER BY"}, {Identifier, "b"},
			{Keyword, "DESC"},
		},
	},
	{
		category: "keywords",
		name:     "GROUP BY across whitespace and a newline",
		sql:      "group \n  by a",
		want: []Token{
			{Keyword, "GROUP BY"}, {Identifier, "a"},
		},
	},
	{
		category: "keywords",
		name:     "condition keywords",
		sql:      "and or not in is null like between exists",
		want: []Token{
			{Keyword, "AND"}, {Keyword, "OR"}, {Keyword, "NOT"}, {Keyword, "IN"},
			{Keyword, "IS"}, {Keyword, "NULL"}, {Keyword, "LIKE"}, {Keyword, "BETWEEN"},
			{Keyword, "EXISTS"},
		},
	},
	{
		category: "keywords",
		name:     "clause keywords",
		sql:      "with select from where having limit offset as asc desc",
		want: []Token{
			{Keyword, "WITH"}, {Keyword, "SELECT"}, {Keyword, "FROM"}, {Keyword, "WHERE"},
			{Keyword, "HAVING"}, {Keyword, "LIMIT"}, {Keyword, "OFFSET"}, {Keyword, "AS"},
			{Keyword, "ASC"}, {Keyword, "DESC"},
		},
	},

	// ---- identifiers ----
	{
		category: "identifiers",
		name:     "keeps original case",
		sql:      "select UserId, first_name from Users",
		want: []Token{
			{Keyword, "SELECT"}, {Identifier, "UserId"}, {Comma, ","}, {Identifier, "first_name"},
			{Keyword, "FROM"}, {Identifier, "Users"},
		},
	},
	{
		category: "identifiers",
		name:     "underscore and digits",
		sql:      "_tmp col2 a_1_b",
		want: []Token{
			{Identifier, "_tmp"}, {Identifier, "col2"}, {Identifier, "a_1_b"},
		},
	},
	{
		category: "identifiers",
		name:     "qualified star alias.* is one token",
		sql:      "u.* db.t.*",
		want: []Token{
			{Identifier, "u.*"}, {Identifier, "db.t.*"},
		},
	},
	{
		category: "identifiers",
		name:     "keyword-like column names stay identifiers",
		sql:      "name left outer",
		want: []Token{
			{Identifier, "name"}, {Identifier, "left"}, {Identifier, "outer"},
		},
	},
	{
		category: "identifiers",
		name:     "qualified names are one token",
		sql:      "select u.id from db.sales.orders",
		want: []Token{
			{Keyword, "SELECT"}, {Identifier, "u.id"}, {Keyword, "FROM"},
			{Identifier, "db.sales.orders"},
		},
	},
	{
		category: "identifiers",
		name:     "table alias with and without AS",
		sql:      "from users u join orders as o",
		want: []Token{
			{Keyword, "FROM"}, {Identifier, "users"}, {Identifier, "u"}, {Keyword, "JOIN"},
			{Identifier, "orders"}, {Keyword, "AS"}, {Identifier, "o"},
		},
	},

	// ---- literals ----
	{
		category: "literals",
		name:     "integer and decimal",
		sql:      "1 42 3.14",
		want: []Token{
			{Number, "1"}, {Number, "42"}, {Number, "3.14"},
		},
	},
	{
		category: "literals",
		name:     "negative number is an operator and a number",
		sql:      "> -5",
		want: []Token{
			{Operator, ">"}, {Operator, "-"}, {Number, "5"},
		},
	},
	{
		category: "literals",
		name:     "string",
		sql:      "'paid'",
		want: []Token{
			{String, "'paid'"},
		},
	},
	{
		category: "literals",
		name:     "string keeps case, spaces and keywords",
		sql:      "'Select FROM  x'",
		want: []Token{
			{String, "'Select FROM  x'"},
		},
	},
	{
		category: "literals",
		name:     "string with escaped quote",
		sql:      "'it''s here'",
		want: []Token{
			{String, "'it''s here'"},
		},
	},
	{
		category: "literals",
		name:     "empty string",
		sql:      "''",
		want: []Token{
			{String, "''"},
		},
	},

	// ---- operators ----
	{
		category: "operators",
		name:     "comparison operators",
		sql:      "= <> != < > <= >=",
		want: []Token{
			{Operator, "="}, {Operator, "<>"}, {Operator, "!="}, {Operator, "<"},
			{Operator, ">"}, {Operator, "<="}, {Operator, ">="},
		},
	},
	{
		category: "operators",
		name:     "arithmetic operators",
		sql:      "a + b - c / d % e",
		want: []Token{
			{Identifier, "a"}, {Operator, "+"}, {Identifier, "b"}, {Operator, "-"},
			{Identifier, "c"}, {Operator, "/"}, {Identifier, "d"}, {Operator, "%"},
			{Identifier, "e"},
		},
	},
	{
		category: "operators",
		name:     "no spaces around operators",
		sql:      "id>=10",
		want: []Token{
			{Identifier, "id"}, {Operator, ">="}, {Number, "10"},
		},
	},
	{
		category: "operators",
		name:     "star is a Star as a wildcard and an Operator after a value",
		sql:      "select *, count(*), price*qty, (a)*2, 3*-1",
		want: []Token{
			{Keyword, "SELECT"}, {Star, "*"}, {Comma, ","}, {Identifier, "count"},
			{LParen, "("}, {Star, "*"}, {RParen, ")"}, {Comma, ","}, {Identifier, "price"},
			{Operator, "*"}, {Identifier, "qty"}, {Comma, ","}, {LParen, "("},
			{Identifier, "a"}, {RParen, ")"}, {Operator, "*"}, {Number, "2"}, {Comma, ","},
			{Number, "3"}, {Operator, "*"}, {Operator, "-"}, {Number, "1"},
		},
	},

	// ---- punctuation and functions ----
	{
		category: "functions",
		name:     "function call",
		sql:      "sum(amount)",
		want: []Token{
			{Identifier, "sum"}, {LParen, "("}, {Identifier, "amount"}, {RParen, ")"},
		},
	},
	{
		category: "functions",
		name:     "nested functions with arguments",
		sql:      "coalesce(sum(amount),0)",
		want: []Token{
			{Identifier, "coalesce"}, {LParen, "("}, {Identifier, "sum"}, {LParen, "("},
			{Identifier, "amount"}, {RParen, ")"}, {Comma, ","}, {Number, "0"}, {RParen, ")"},
		},
	},
	{
		category: "functions",
		name:     "count distinct",
		sql:      "count(distinct user_id)",
		want: []Token{
			{Identifier, "count"}, {LParen, "("}, {Keyword, "DISTINCT"},
			{Identifier, "user_id"}, {RParen, ")"},
		},
	},
	{
		category: "functions",
		name:     "IN list",
		sql:      "id in (1,2,3)",
		want: []Token{
			{Identifier, "id"}, {Keyword, "IN"}, {LParen, "("}, {Number, "1"}, {Comma, ","},
			{Number, "2"}, {Comma, ","}, {Number, "3"}, {RParen, ")"},
		},
	},

	// ---- JOINs ----
	{
		category: "joins",
		name:     "JOIN alone",
		sql:      "join t on a = b",
		want: []Token{
			{Keyword, "JOIN"}, {Identifier, "t"}, {Keyword, "ON"}, {Identifier, "a"},
			{Operator, "="}, {Identifier, "b"},
		},
	},
	{
		category: "joins",
		name:     "JOIN types are one token each",
		sql:      "inner join a left join b right join c full join d cross join e",
		want: []Token{
			{Keyword, "INNER JOIN"}, {Identifier, "a"}, {Keyword, "LEFT JOIN"}, {Identifier, "b"},
			{Keyword, "RIGHT JOIN"}, {Identifier, "c"}, {Keyword, "FULL JOIN"}, {Identifier, "d"},
			{Keyword, "CROSS JOIN"}, {Identifier, "e"},
		},
	},
	{
		category: "joins",
		name:     "OUTER JOIN types are one token each",
		sql:      "Left Outer Join a full outer join b",
		want: []Token{
			{Keyword, "LEFT OUTER JOIN"}, {Identifier, "a"}, {Keyword, "FULL OUTER JOIN"},
			{Identifier, "b"},
		},
	},
	{
		category: "joins",
		name:     "JOIN type across a newline",
		sql:      "left\n  join a",
		want: []Token{
			{Keyword, "LEFT JOIN"}, {Identifier, "a"},
		},
	},
	{
		category: "joins",
		name:     "only listed JOIN phrases merge",
		sql:      "left outer inner join a left left join b",
		want: []Token{
			{Identifier, "left"}, {Identifier, "outer"}, {Keyword, "INNER JOIN"}, {Identifier, "a"},
			{Identifier, "left"}, {Keyword, "LEFT JOIN"}, {Identifier, "b"},
		},
	},
	{
		category: "joins",
		name:     "USING",
		sql:      "join t using (id)",
		want: []Token{
			{Keyword, "JOIN"}, {Identifier, "t"}, {Keyword, "USING"}, {LParen, "("},
			{Identifier, "id"}, {RParen, ")"},
		},
	},
	{
		category: "joins",
		name:     "LEFT and RIGHT functions stay identifiers",
		sql:      "left(email, 3) right(email, 2)",
		want: []Token{
			{Identifier, "left"}, {LParen, "("}, {Identifier, "email"}, {Comma, ","},
			{Number, "3"}, {RParen, ")"}, {Identifier, "right"}, {LParen, "("},
			{Identifier, "email"}, {Comma, ","}, {Number, "2"}, {RParen, ")"},
		},
	},

	// ---- subqueries and CTEs ----
	{
		category: "subqueries",
		name:     "subquery in parentheses",
		sql:      "id in (select user_id from orders)",
		want: []Token{
			{Identifier, "id"}, {Keyword, "IN"}, {LParen, "("}, {Keyword, "SELECT"},
			{Identifier, "user_id"}, {Keyword, "FROM"}, {Identifier, "orders"}, {RParen, ")"},
		},
	},
	{
		category: "subqueries",
		name:     "CTE with column list",
		sql:      "with t (a, b) as (select 1, 2)",
		want: []Token{
			{Keyword, "WITH"}, {Identifier, "t"}, {LParen, "("}, {Identifier, "a"}, {Comma, ","},
			{Identifier, "b"}, {RParen, ")"}, {Keyword, "AS"}, {LParen, "("}, {Keyword, "SELECT"},
			{Number, "1"}, {Comma, ","}, {Number, "2"}, {RParen, ")"},
		},
	},
}

func TestTokenize(t *testing.T) {
	for _, tc := range tokenizeCases {
		t.Run(tc.category+"/"+tc.name, func(t *testing.T) {
			got, err := Tokenize(tc.sql)
			if err != nil {
				t.Fatalf("Tokenize(%q) returned error: %v", tc.sql, err)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("Tokenize(%q)\n got: %v\nwant: %v", tc.sql, got, tc.want)
			}
		})
	}
}

func TestTokenizeErrors(t *testing.T) {
	tests := []struct {
		name    string
		sql     string
		wantErr string
	}{
		{"unterminated string", "select 'abc", "unterminated string at position 7"},
		{"string ends with escaped quote", "select 'it''", "unterminated string at position 7"},
		{"lone exclamation mark", "a ! b", "unexpected character '!' at position 2"},
		{"unknown character", "select # from t", "unexpected character '#' at position 7"},
		{"line comment", "a -- b", "comments are not supported at position 2"},
		{"block comment", "a /* b */", "comments are not supported at position 2"},
		{"number with two dots", "1.2.3", "malformed number at position 0"},
		{"number ends with a dot", "1. ", "malformed number at position 0"},
		{"number followed by letters", "12ab", "malformed number at position 0"},
		{"quoted identifier", `"a"`, `unexpected character '"' at position 0`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Tokenize(tt.sql)
			if err == nil {
				t.Fatalf("Tokenize(%q) = %v, want error %q", tt.sql, got, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Tokenize(%q) error = %q, want %q", tt.sql, err, tt.wantErr)
			}
		})
	}
}
