package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// datasetGlob finds the pattern datasets. Each file holds the valid
// queries of one pattern and feeds both TestFormat and TestTokenize.
// MANIFEST.md in the same folder lists the patterns.
const datasetGlob = "testdata/patterns/*.json"

// dataset is one pattern file.
type dataset struct {
	file    string
	Pattern string         `json:"pattern"`
	Classes []datasetClass `json:"classes"`
	Cases   []datasetCase  `json:"cases"`
}

// datasetClass is one equivalence class of a pattern.
type datasetClass struct {
	ID   string `json:"id"`
	Desc string `json:"desc"`
}

// datasetCase is one input query with its expected output.
// FormatWant is the formatted query, one entry per line.
// TokensWant holds [type, value] pairs. Each is optional, but a case
// needs at least one of them.
type datasetCase struct {
	Name       string     `json:"name"`
	Class      string     `json:"class"`
	SQL        *string    `json:"sql"`
	FormatWant []string   `json:"format_want"`
	TokensWant [][]string `json:"tokens_want"`
}

// testName is the subtest name: pattern/class/name.
func (d dataset) testName(c datasetCase) string {
	return d.Pattern + "/" + c.Class + "/" + c.Name
}

// loadDatasets reads every dataset file. It fails the test when no
// file is found or a file does not decode.
func loadDatasets(t *testing.T) []dataset {
	t.Helper()
	files, err := filepath.Glob(datasetGlob)
	if err != nil {
		t.Fatalf("glob %s: %v", datasetGlob, err)
	}
	if len(files) == 0 {
		t.Fatalf("no dataset files match %s", datasetGlob)
	}
	var sets []dataset
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.DisallowUnknownFields()
		var d dataset
		if err := dec.Decode(&d); err != nil {
			t.Fatalf("decode %s: %v", f, err)
		}
		if err := dec.Decode(new(json.RawMessage)); err != io.EOF {
			t.Fatalf("decode %s: trailing data after the dataset object", f)
		}
		d.file = f
		sets = append(sets, d)
	}
	return sets
}

// tokenTypeByName maps a token type name, as used in tokens_want,
// to its TokenType.
func tokenTypeByName() map[string]TokenType {
	m := make(map[string]TokenType, len(tokenTypeNames))
	for i, name := range tokenTypeNames {
		m[name] = TokenType(i)
	}
	return m
}

// parseTokens turns tokens_want pairs into Tokens.
func parseTokens(pairs [][]string) ([]Token, error) {
	types := tokenTypeByName()
	tokens := make([]Token, 0, len(pairs))
	for i, p := range pairs {
		if len(p) != 2 {
			return nil, fmt.Errorf("tokens_want[%d]: want [type, value], got %q", i, p)
		}
		typ, ok := types[p[0]]
		if !ok {
			return nil, fmt.Errorf("tokens_want[%d]: unknown token type %q", i, p[0])
		}
		tokens = append(tokens, Token{typ, p[1]})
	}
	return tokens, nil
}

// TestDatasets checks the shape of every dataset file.
func TestDatasets(t *testing.T) {
	for _, d := range loadDatasets(t) {
		t.Run(filepath.Base(d.file), func(t *testing.T) {
			want := strings.TrimSuffix(filepath.Base(d.file), ".json")
			if d.Pattern != want {
				t.Errorf("pattern = %q, want %q (the file name)", d.Pattern, want)
			}
			if len(d.Classes) == 0 {
				t.Errorf("no classes")
			}
			if len(d.Cases) == 0 {
				t.Errorf("no cases")
			}

			classes := map[string]bool{}
			for _, c := range d.Classes {
				if c.ID == "" || c.Desc == "" {
					t.Errorf("class %q: id and desc must not be empty", c.ID)
				}
				if classes[c.ID] {
					t.Errorf("class %q: duplicate id", c.ID)
				}
				classes[c.ID] = true
			}

			used := map[string]bool{}
			names := map[string]bool{}
			for i, c := range d.Cases {
				if c.Name == "" {
					t.Errorf("case %d: empty name", i)
				}
				if names[c.Name] {
					t.Errorf("case %q: duplicate name", c.Name)
				}
				names[c.Name] = true
				if c.SQL == nil {
					t.Errorf("case %q: missing sql", c.Name)
				}
				if c.FormatWant == nil && c.TokensWant == nil {
					t.Errorf("case %q: needs format_want or tokens_want", c.Name)
				}
				if c.FormatWant != nil && len(c.FormatWant) == 0 {
					t.Errorf("case %q: format_want is empty", c.Name)
				}
				if !classes[c.Class] {
					t.Errorf("case %q: unknown class %q", c.Name, c.Class)
				}
				used[c.Class] = true
				if _, err := parseTokens(c.TokensWant); err != nil {
					t.Errorf("case %q: %v", c.Name, err)
				}
			}
			for _, c := range d.Classes {
				if !used[c.ID] {
					t.Errorf("class %q: no case uses it", c.ID)
				}
			}
		})
	}
}
