package formatter

import "testing"

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want string
	}{
		{
			name: "select star with where",
			sql:  "SELECT * FROM table_name WHERE id = 1",
			want: `SELECT *
FROM   table_name
WHERE  id = 1`,
		},
		{
			name: "multiple columns",
			sql:  "SELECT col1, col2 FROM table_name WHERE id = 1",
			want: `SELECT col1,
       col2
FROM   table_name
WHERE  id = 1`,
		},
		{
			name: "where with and",
			sql:  "SELECT * FROM table_name WHERE id = 1 and name = 'ABC'",
			want: `SELECT *
FROM   table_name
WHERE  id = 1
       AND NAME = 'ABC'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format(tt.sql)
			if err != nil {
				t.Fatalf("Format(%q) returned error: %v", tt.sql, err)
			}
			if got != tt.want {
				t.Errorf("Format(%q)\n got:\n%s\nwant:\n%s", tt.sql, got, tt.want)
			}
		})
	}
}
