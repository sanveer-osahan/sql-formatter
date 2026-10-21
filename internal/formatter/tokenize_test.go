package formatter

import (
	"slices"
	"testing"
)

func TestTokenize(t *testing.T) {
	sql := "SELECT * FROM table_name WHERE id = 1"
	want := []Token{
		{Keyword, "SELECT"},
		{Star, "*"},
		{Keyword, "FROM"},
		{Identifier, "table_name"},
		{Keyword, "WHERE"},
		{Identifier, "id"},
		{Operator, "="},
		{Number, "1"},
	}

	got, err := Tokenize(sql)
	if err != nil {
		t.Fatalf("Tokenize(%q) returned error: %v", sql, err)
	}
	if !slices.Equal(got, want) {
		t.Errorf("Tokenize(%q)\n got: %v\nwant: %v", sql, got, want)
	}
}
