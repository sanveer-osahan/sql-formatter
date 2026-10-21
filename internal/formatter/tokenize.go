package formatter

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType is the kind of a token.
type TokenType int

const (
	Keyword    TokenType = iota // SELECT, FROM, WHERE, GROUP BY, ...
	Identifier                  // column or table name, can be qualified: db.schema.table
	Star                        // * as in SELECT * and count(*)
	Number                      // 1, 3.14
	String                      // 'text'
	Operator                    // =, <>, !=, <, >, <=, >=, +, -, *, /, %
	Comma                       // ,
	LParen                      // (
	RParen                      // )
)

var tokenTypeNames = [...]string{
	Keyword:    "Keyword",
	Identifier: "Identifier",
	Star:       "Star",
	Number:     "Number",
	String:     "String",
	Operator:   "Operator",
	Comma:      "Comma",
	LParen:     "LParen",
	RParen:     "RParen",
}

func (t TokenType) String() string {
	if int(t) < len(tokenTypeNames) {
		return tokenTypeNames[t]
	}
	return fmt.Sprintf("TokenType(%d)", int(t))
}

// Token is one unit of SQL text.
type Token struct {
	Type  TokenType
	Value string
}

// keywords is the set of reserved words. Each one-word clause keyword
// in the clauses table is added to it.
var keywords = map[string]bool{
	"DISTINCT": true, "GROUP": true, "ORDER": true, "BY": true, "AS": true,
	"AND": true, "OR": true, "NOT": true, "IN": true, "IS": true, "NULL": true,
	"LIKE": true, "BETWEEN": true, "ASC": true, "DESC": true, "EXISTS": true,
}

// plainPhrases are multi-word keywords that are not clauses, so they
// stay on the current line, such as DEFAULT VALUES after INSERT INTO t.
var plainPhrases = []string{"DEFAULT VALUES"}

// phrases are the multi-word keywords from the clauses table and
// plainPhrases, such as GROUP BY and LEFT OUTER JOIN. Their words become a single keyword
// only in this order; a word alone is a keyword only if it is in keywords.
var phrases [][]string

func init() {
	for _, c := range clauses {
		words := strings.Fields(c.keyword)
		if len(words) == 1 {
			keywords[c.keyword] = true
		} else {
			phrases = append(phrases, words)
		}
	}
	for _, p := range plainPhrases {
		phrases = append(phrases, strings.Fields(p))
	}
}

// Tokenize splits a SELECT query into tokens.
// Keyword values are upper case. Other values keep the original text.
func Tokenize(sql string) ([]Token, error) {
	src := []rune(sql)
	var tokens []Token

	for i := 0; i < len(src); {
		c := src[i]

		switch {
		case unicode.IsSpace(c):
			i++

		case isIdentStart(c):
			if phrase, end := matchPhrase(src, i); end > 0 {
				tokens = append(tokens, Token{Keyword, phrase})
				i = end
				break
			}
			start := i
			i = scanIdent(src, i)
			for i+1 < len(src) && src[i] == '.' && isIdentStart(src[i+1]) {
				i = scanIdent(src, i+1)
			}
			if i+1 < len(src) && src[i] == '.' && src[i+1] == '*' { // alias.*
				i += 2
			}
			word := string(src[start:i])
			upper := strings.ToUpper(word)
			if keywords[upper] {
				tokens = append(tokens, Token{Keyword, upper})
			} else {
				tokens = append(tokens, Token{Identifier, word})
			}

		case unicode.IsDigit(c):
			start := i
			for i < len(src) && unicode.IsDigit(src[i]) {
				i++
			}
			if i+1 < len(src) && src[i] == '.' && unicode.IsDigit(src[i+1]) {
				i = i + 1
				for i < len(src) && unicode.IsDigit(src[i]) {
					i++
				}
			}
			if i < len(src) && (src[i] == '.' || isIdentStart(src[i])) {
				return nil, fmt.Errorf("malformed number at position %d", start)
			}
			tokens = append(tokens, Token{Number, string(src[start:i])})

		case c == '\'':
			start := i
			i++
			for {
				if i >= len(src) {
					return nil, fmt.Errorf("unterminated string at position %d", start)
				}
				if src[i] == '\'' {
					if i+1 < len(src) && src[i+1] == '\'' { // '' is an escaped quote
						i += 2
						continue
					}
					i++
					break
				}
				i++
			}
			tokens = append(tokens, Token{String, string(src[start:i])})

		case c == '*':
			if isValue(tokens) {
				tokens = append(tokens, Token{Operator, "*"}) // multiply: price * qty
			} else {
				tokens = append(tokens, Token{Star, "*"})
			}
			i++

		case c == '-' && i+1 < len(src) && src[i+1] == '-',
			c == '/' && i+1 < len(src) && src[i+1] == '*':
			return nil, fmt.Errorf("comments are not supported at position %d", i)

		case c == ',':
			tokens = append(tokens, Token{Comma, ","})
			i++

		case c == '(':
			tokens = append(tokens, Token{LParen, "("})
			i++

		case c == ')':
			tokens = append(tokens, Token{RParen, ")"})
			i++

		case strings.ContainsRune("=<>!+-/%", c):
			start := i
			i++
			if i < len(src) {
				two := string(src[start : i+1])
				if two == "<>" || two == "!=" || two == "<=" || two == ">=" {
					i++
				}
			}
			op := string(src[start:i])
			if op == "!" {
				return nil, fmt.Errorf("unexpected character %q at position %d", c, start)
			}
			tokens = append(tokens, Token{Operator, op})

		default:
			return nil, fmt.Errorf("unexpected character %q at position %d", c, i)
		}
	}

	return tokens, nil
}

func isIdentStart(c rune) bool {
	return unicode.IsLetter(c) || c == '_'
}

func scanIdent(src []rune, i int) int {
	for i < len(src) && (unicode.IsLetter(src[i]) || unicode.IsDigit(src[i]) || src[i] == '_') {
		i++
	}
	return i
}

// matchPhrase reports the longest multi-word keyword that starts at
// src[i], and the position after it. end is 0 if there is none.
func matchPhrase(src []rune, i int) (phrase string, end int) {
	for _, words := range phrases {
		j := i
		matched := true
		for k, word := range words {
			if k > 0 {
				for j < len(src) && unicode.IsSpace(src[j]) {
					j++
				}
			}
			start := j
			j = scanIdent(src, j)
			if j == start || strings.ToUpper(string(src[start:j])) != word ||
				j < len(src) && src[j] == '.' {
				matched = false
				break
			}
		}
		if matched && j > end {
			phrase, end = strings.Join(words, " "), j
		}
	}
	return phrase, end
}

// isValue reports whether the last token ends a value, so a "*"
// after it is a multiply and not a Star.
func isValue(tokens []Token) bool {
	if len(tokens) == 0 {
		return false
	}
	switch tokens[len(tokens)-1].Type {
	case Identifier, Number, String, RParen:
		return true
	}
	return false
}
