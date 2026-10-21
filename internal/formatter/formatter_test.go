package formatter

import (
	"strings"
	"testing"
)

// TestFormat runs the cases of the pattern datasets in
// testdata/patterns that have format_want. The layout rules are in
// STYLE.md at the repo root.
func TestFormat(t *testing.T) {
	for _, d := range loadDatasets(t) {
		for _, c := range d.Cases {
			if c.FormatWant == nil || c.SQL == nil {
				continue // a case without sql is reported by TestDatasets
			}
			t.Run(d.testName(c), func(t *testing.T) {
				want := strings.Join(c.FormatWant, "\n")
				got, err := Format(*c.SQL)
				if err != nil {
					t.Fatalf("Format(%q) returned error: %v", *c.SQL, err)
				}
				if got != want {
					t.Errorf("Format(%q)\n got:\n%s\nwant:\n%s", *c.SQL, got, want)
				}
			})
		}
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
