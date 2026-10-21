// Package formatter formats SQL text.
package formatter

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// clauseWidth is the width of "SELECT ". Clause keywords are padded
// to it, so their lists line up with the text after SELECT.
const clauseWidth = len("SELECT ")

// clause is the layout of one clause keyword.
type clause struct {
	keyword     string // upper case words, such as "GROUP BY"
	indent      int    // column of the keyword from the start of the query
	width       int    // the keyword and a space are padded to this width
	commas      bool   // a top-level comma starts a new line under the first item
	conditions  bool   // a top-level AND or OR starts a new line under the first condition
	nested      bool   // part of the clause above it, such as JOIN in FROM; commas follow that clause
	join        bool   // ON and USING can follow it
	afterJoin   bool   // a clause only after a join, otherwise a plain keyword
	startsQuery bool   // a "(" before it opens a subquery
	columnLists bool   // a "(" after a name opens a column list, not a function call
}

// clauses is the layout of each clause. A multi-word keyword is one
// token, see phrases. To add a clause, add a row here.
var clauses = []clause{
	{keyword: "WITH", commas: true, startsQuery: true, columnLists: true},
	{keyword: "SELECT", width: clauseWidth, commas: true, startsQuery: true},
	{keyword: "FROM", width: clauseWidth, commas: true},
	{keyword: "WHERE", width: clauseWidth, conditions: true},
	{keyword: "GROUP BY", width: clauseWidth},
	{keyword: "HAVING", width: clauseWidth, conditions: true},
	{keyword: "ORDER BY", width: clauseWidth},
	{keyword: "LIMIT", width: clauseWidth},
	{keyword: "OFFSET", width: clauseWidth},
	{keyword: "JOIN", indent: clauseWidth, nested: true, join: true},
	{keyword: "INNER JOIN", indent: clauseWidth, nested: true, join: true},
	{keyword: "LEFT JOIN", indent: clauseWidth, nested: true, join: true},
	{keyword: "LEFT OUTER JOIN", indent: clauseWidth, nested: true, join: true},
	{keyword: "RIGHT JOIN", indent: clauseWidth, nested: true, join: true},
	{keyword: "RIGHT OUTER JOIN", indent: clauseWidth, nested: true, join: true},
	{keyword: "FULL JOIN", indent: clauseWidth, nested: true, join: true},
	{keyword: "FULL OUTER JOIN", indent: clauseWidth, nested: true, join: true},
	{keyword: "CROSS JOIN", indent: clauseWidth, nested: true, join: true},
	{keyword: "ON", indent: clauseWidth + 2, nested: true, afterJoin: true, conditions: true},
	{keyword: "USING", indent: clauseWidth + 2, nested: true, afterJoin: true},
}

// clauseByKeyword finds a row of clauses by its keyword.
var clauseByKeyword = map[string]*clause{}

func init() {
	for i := range clauses {
		clauseByKeyword[clauses[i].keyword] = &clauses[i]
	}
}

// clauseOf returns the clause that t starts, or nil.
func clauseOf(t Token) *clause {
	if t.Type != Keyword {
		return nil
	}
	return clauseByKeyword[t.Value]
}

// Format tokenizes the SQL and lays it out one clause per line, as
// the clauses table says. See STYLE.md at the repo root for the
// layout rules. It returns an error when the SQL does not tokenize
// (see Tokenize) or its parentheses are unbalanced.
func Format(sql string) (string, error) {
	tokens, err := Tokenize(sql)
	if err != nil {
		return "", err
	}
	if err := checkParens(tokens); err != nil {
		return "", err
	}

	var w writer
	if err := formatQuery(&w, tokens, 0); err != nil {
		return "", err
	}
	return w.b.String(), nil
}

// checkParens returns an error if a "(" is not closed or a ")" is
// not opened.
func checkParens(tokens []Token) error {
	depth := 0
	for _, t := range tokens {
		switch t.Type {
		case LParen:
			depth++
		case RParen:
			depth--
			if depth < 0 {
				return fmt.Errorf("unbalanced parentheses: unexpected \")\"")
			}
		}
	}
	if depth > 0 {
		return fmt.Errorf("unbalanced parentheses: missing \")\"")
	}
	return nil
}

// writer builds the output and tracks the current column.
type writer struct {
	b   strings.Builder
	col int // column of the next character on the current line
}

func (w *writer) write(s string) {
	w.b.WriteString(s)
	if k := strings.LastIndexByte(s, '\n'); k >= 0 {
		w.col = utf8.RuneCountInString(s[k+1:])
	} else {
		w.col += utf8.RuneCountInString(s)
	}
}

// newline starts a new line and indents it to col.
func (w *writer) newline(col int) {
	w.write("\n" + strings.Repeat(" ", col))
}

// formatQuery writes one query. base is the column where the query
// starts; each new line of the query is indented to base.
func formatQuery(w *writer, tokens []Token, base int) error {
	var prev Token
	var cur *clause    // the clause of the current line
	commaCol := -1     // column of the line after a top-level comma, -1 if commas do not break
	conditionCol := -1 // column of a top-level AND or OR line, -1 if they do not break
	depth := 0         // parenthesis depth, not counting subqueries
	lineStart := true  // nothing written on the current line after the indent
	inBetween := false // the next AND belongs to BETWEEN x AND y
	prevUnary := false // the previous token was a sign, such as the - in -5

	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		c := clauseOf(t)
		switch {
		case c != nil && depth == 0 && (!c.afterJoin || cur != nil && cur.join):
			if i > 0 {
				w.newline(base + c.indent)
			}
			w.write(fmt.Sprintf("%-*s", c.width, c.keyword+" "))
			if !c.nested {
				commaCol = -1
				if c.commas {
					commaCol = w.col
				}
			}
			conditionCol = -1
			if c.conditions {
				conditionCol = w.col
			}
			cur = c
			lineStart = true

		case t.Type == Keyword && (t.Value == "AND" || t.Value == "OR") &&
			conditionCol >= 0 && depth == 0 && !inBetween:
			w.newline(conditionCol)
			w.write(t.Value)
			lineStart = false

		case t.Type == LParen && startsSubquery(tokens, i):
			end := matchingParen(tokens, i)
			if end < 0 {
				return fmt.Errorf("unbalanced parentheses")
			}
			if !lineStart && !prevUnary && needsSpace(prev, t) {
				w.write(" ")
			}
			w.write("(")
			if err := formatQuery(w, tokens[i+1:end], w.col); err != nil {
				return err
			}
			w.write(")")
			i = end
			t = tokens[end]
			lineStart = false
			prevUnary = false

		default:
			unary := isUnarySign(prev, t, lineStart)
			space := needsSpace(prev, t) ||
				cur != nil && cur.columnLists && depth == 0 && t.Type == LParen // CTE column list: totals (a, b)
			if !lineStart && !prevUnary && space {
				w.write(" ")
			}
			w.write(t.Value)
			lineStart = false
			prevUnary = unary

			if t.Type == Keyword {
				switch t.Value {
				case "BETWEEN":
					inBetween = true
				case "AND":
					inBetween = false
				}
			}

			switch t.Type {
			case LParen:
				depth++
			case RParen:
				depth--
			case Comma:
				if depth == 0 && commaCol >= 0 {
					w.newline(commaCol)
					lineStart = true
				}
			}
		}
		prev = t
	}

	return nil
}

// startsSubquery reports whether the "(" at tokens[i] opens a subquery.
func startsSubquery(tokens []Token, i int) bool {
	if i+1 >= len(tokens) {
		return false
	}
	c := clauseOf(tokens[i+1])
	return c != nil && c.startsQuery
}

// matchingParen returns the index of the ")" that closes the "(" at
// tokens[i], or -1 if there is none.
func matchingParen(tokens []Token, i int) int {
	depth := 0
	for j := i; j < len(tokens); j++ {
		switch tokens[j].Type {
		case LParen:
			depth++
		case RParen:
			depth--
			if depth == 0 {
				return j
			}
		}
	}
	return -1
}

// isUnarySign reports whether t is a + or - sign before a value,
// such as the - in "> -5", and not an arithmetic operator.
func isUnarySign(prev, t Token, lineStart bool) bool {
	if t.Type != Operator || (t.Value != "-" && t.Value != "+") {
		return false
	}
	if lineStart {
		return true
	}
	switch prev.Type {
	case Operator, Keyword, LParen, Comma:
		return true
	}
	return false
}

// needsSpace reports whether a space goes between prev and t.
func needsSpace(prev, t Token) bool {
	switch {
	case t.Type == Comma, t.Type == RParen:
		return false
	case prev.Type == LParen:
		return false
	case t.Type == LParen && prev.Type == Identifier: // function call: count(
		return false
	}
	return true
}
