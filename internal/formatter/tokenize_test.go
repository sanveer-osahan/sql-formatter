package formatter

import (
	"slices"
	"strings"
	"testing"
)

// TestTokenize runs the cases of the pattern datasets in
// testdata/patterns that have tokens_want.
func TestTokenize(t *testing.T) {
	for _, d := range loadDatasets(t) {
		for _, c := range d.Cases {
			if c.TokensWant == nil || c.SQL == nil {
				continue // a case without sql is reported by TestDatasets
			}
			t.Run(d.testName(c), func(t *testing.T) {
				want, err := parseTokens(c.TokensWant)
				if err != nil {
					t.Fatalf("bad tokens_want: %v", err)
				}
				got, err := Tokenize(*c.SQL)
				if err != nil {
					t.Fatalf("Tokenize(%q) returned error: %v", *c.SQL, err)
				}
				if !slices.Equal(got, want) {
					t.Errorf("Tokenize(%q)\n got: %v\nwant: %v", *c.SQL, got, want)
				}
			})
		}
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
