// Package formatter formats SQL text.
package formatter

import (
	"fmt"
	"strings"
)

// indent is the width of "SELECT ", so continuation lines line up
// with the text after the clause keyword.
const indent = "       "

// clauseKeywords start a new line.
var clauseKeywords = map[string]bool{
	"SELECT": true, "FROM": true, "WHERE": true, "GROUP BY": true,
	"ORDER BY": true, "HAVING": true, "LIMIT": true, "OFFSET": true,
}

// Format tokenizes the SQL and lays it out one clause per line:
//   - each clause keyword starts a line, padded to the indent width
//   - each SELECT column after a comma goes on its own line
//   - each AND or OR in WHERE and HAVING goes on its own line
func Format(sql string) (string, error) {
	tokens, err := Tokenize(sql)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	var prev Token
	clause := ""
	depth := 0        // parenthesis depth
	lineStart := true // nothing written on the current line after the indent

	for i, t := range tokens {
		switch {
		case t.Type == Keyword && clauseKeywords[t.Value] && depth == 0:
			if i > 0 {
				b.WriteByte('\n')
			}
			fmt.Fprintf(&b, "%-*s", len(indent), t.Value+" ")
			clause = t.Value
			lineStart = true

		case t.Type == Keyword && (t.Value == "AND" || t.Value == "OR") &&
			(clause == "WHERE" || clause == "HAVING") && depth == 0:
			b.WriteString("\n" + indent + t.Value)
			lineStart = false

		default:
			if !lineStart && needsSpace(prev, t) {
				b.WriteByte(' ')
			}
			b.WriteString(t.Value)
			lineStart = false

			switch t.Type {
			case LParen:
				depth++
			case RParen:
				depth--
			case Comma:
				if clause == "SELECT" && depth == 0 {
					b.WriteString("\n" + indent)
					lineStart = true
				}
			}
		}
		prev = t
	}

	return b.String(), nil
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
