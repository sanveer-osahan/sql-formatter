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
	Star                        // *
	Number                      // 1, 3.14
	String                      // 'text'
	Operator                    // =, <>, !=, <, >, <=, >=, +, -, /, %
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

// keywords is the set of reserved words. GROUP and ORDER are joined
// with a following BY into a single "GROUP BY" or "ORDER BY" keyword.
var keywords = map[string]bool{
	"SELECT": true, "DISTINCT": true, "FROM": true, "WHERE": true,
	"GROUP": true, "ORDER": true, "BY": true, "HAVING": true, "LIMIT": true,
	"OFFSET": true, "AS": true, "AND": true, "OR": true, "NOT": true,
	"IN": true, "IS": true, "NULL": true, "LIKE": true, "BETWEEN": true,
	"ASC": true, "DESC": true, "NAME": true,
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
			start := i
			i = scanIdent(src, i)
			for i+1 < len(src) && src[i] == '.' && isIdentStart(src[i+1]) {
				i = scanIdent(src, i+1)
			}
			word := string(src[start:i])
			upper := strings.ToUpper(word)
			if !keywords[upper] {
				tokens = append(tokens, Token{Identifier, word})
				break
			}
			if upper == "BY" && len(tokens) > 0 {
				last := &tokens[len(tokens)-1]
				if last.Type == Keyword && (last.Value == "GROUP" || last.Value == "ORDER") {
					last.Value += " BY"
					break
				}
			}
			tokens = append(tokens, Token{Keyword, upper})

		case unicode.IsDigit(c):
			start := i
			for i < len(src) && (unicode.IsDigit(src[i]) || src[i] == '.') {
				i++
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
			tokens = append(tokens, Token{Star, "*"})
			i++

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
