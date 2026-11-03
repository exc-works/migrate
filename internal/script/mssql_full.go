//go:build full

package script

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

type mssqlToken struct {
	word  string
	depth int
}

// prepareMSSQL rejects batches that need no semicolon in T-SQL. This is a
// conservative direct-statement gate, not a general T-SQL parser. Quoted values,
// delimited identifiers and nested comments never become control keywords.
func prepareMSSQL(query string) (string, error) {
	tokens, err := mssqlTokens(query)
	if err != nil {
		return "", err
	}
	root := ""
	alterTable := len(tokens) > 1 && tokens[0].word == "ALTER" && tokens[1].word == "TABLE"
	updatePending, insertSource, setSelect, referentialSet := false, false, false, false
	previous := ""
	for index, token := range tokens {
		word := token.word
		switch word {
		case "COMMIT", "ROLLBACK", "ABORT", "BEGIN", "SAVE", "TRAN", "TRANSACTION", "EXEC", "EXECUTE", "CALL", "DECLARE", "USE", "DBCC", "WAITFOR", "KILL", "BACKUP", "RESTORE", "GRANT", "REVOKE", "DENY", "RECONFIGURE", "SHUTDOWN", "PROC", "PROCEDURE", "FUNCTION", "TRIGGER":
			return "", errors.New("MSSQL manual transaction, session control or dynamic SQL is unsupported; use SQL migrations for batches")
		}
		if token.depth != 0 {
			continue
		}
		if root == "" {
			root = word
			updatePending = word == "UPDATE"
			previous = word
			continue
		}
		switch word {
		case "SET":
			if referentialSet {
				referentialSet = false
				break
			}
			if !updatePending {
				return "", errors.New("MSSQL session control is unsupported; use SQL migrations for batches")
			}
			updatePending = false
		case "STATISTICS":
			updatePending = false
		case "UNION", "INTERSECT", "EXCEPT":
			setSelect = true
		case "VALUES":
			if root == "INSERT" {
				insertSource = true
			}
		case "SELECT", "INSERT", "UPDATE", "DELETE", "MERGE", "CREATE", "ALTER", "DROP", "TRUNCATE", "REPLACE":
			allowed := false
			next, afterNext := mssqlTopLevelWord(tokens, index+1), mssqlTopLevelWord(tokens, index+2)
			referential := alterTable && previous == "ON" && (word == "DELETE" || word == "UPDATE") &&
				(next == "CASCADE" || next == "NO" && afterNext == "ACTION" || next == "SET" && (afterNext == "NULL" || afterNext == "DEFAULT"))
			switch {
			case root == "WITH":
				root = word
				allowed = true
			case alterTable && (word == "ALTER" && next == "COLUMN" || word == "DROP" && (next == "COLUMN" || next == "CONSTRAINT")):
				// These clauses cannot start a standalone T-SQL statement. Keep
				// ALTER TABLE and DROP TABLE rejected as additional statements.
				allowed = true
			case referential:
				allowed = true
				referentialSet = next == "SET"
			case root == "CREATE" && previous == "OR" && word == "ALTER" && next == "VIEW":
				allowed = true
			case word == "SELECT" && setSelect:
				setSelect = false
				allowed = true
			case word == "SELECT" && root == "INSERT" && !insertSource:
				insertSource = true
				allowed = true
			case word == "SELECT" && (root == "CREATE" || root == "ALTER") && previous == "AS":
				allowed = true
			case root == "MERGE" && previous == "THEN" && (word == "UPDATE" || word == "INSERT" || word == "DELETE"):
				allowed = true
			}
			if !allowed {
				return "", errors.New("MSSQL scripts require one direct statement; use SQL migrations for batches")
			}
			if word == "UPDATE" && !referential {
				updatePending = true
			}
		}
		previous = word
	}
	// Unlike most T-SQL statements, MERGE requires a terminator. A newline also
	// keeps it outside any trailing line comment stripped by neither SQL nor JS.
	if root == "MERGE" {
		return query + "\n;", nil
	}
	return query, nil
}

func mssqlTopLevelWord(tokens []mssqlToken, index int) string {
	if index < len(tokens) && tokens[index].depth == 0 {
		return tokens[index].word
	}
	return ""
}

func mssqlTokens(query string) ([]mssqlToken, error) {
	var tokens []mssqlToken
	depth := 0
	for pos := 0; pos < len(query); {
		switch {
		case strings.HasPrefix(query[pos:], "--"):
			pos += 2
			for pos < len(query) && query[pos] != '\r' && query[pos] != '\n' {
				pos++
			}
		case strings.HasPrefix(query[pos:], "/*"):
			pos += 2
			nesting := 1
			for pos < len(query) && nesting > 0 {
				if strings.HasPrefix(query[pos:], "/*") {
					nesting++
					pos += 2
				} else if strings.HasPrefix(query[pos:], "*/") {
					nesting--
					pos += 2
				} else {
					pos++
				}
			}
			if nesting != 0 {
				return nil, errors.New("unterminated MSSQL comment")
			}
		case query[pos] == '\'' || query[pos] == '"' || query[pos] == '[':
			end := query[pos]
			if end == '[' {
				end = ']'
			}
			pos++
			closed := false
			for pos < len(query) {
				if query[pos] != end {
					pos++
					continue
				}
				pos++
				if pos < len(query) && query[pos] == end {
					pos++
					continue
				}
				closed = true
				break
			}
			if !closed {
				return nil, errors.New("unterminated MSSQL quoted value or identifier")
			}
		case query[pos] == '(':
			depth++
			pos++
		case query[pos] == ')':
			depth--
			pos++
			if depth < 0 {
				return nil, errors.New("unbalanced MSSQL statement")
			}
		case mssqlWordSize(query[pos:]) > 0:
			start := pos
			for pos < len(query) {
				size := mssqlWordSize(query[pos:])
				if size == 0 {
					break
				}
				pos += size
			}
			tokens = append(tokens, mssqlToken{
				word:  strings.ToUpper(query[start:pos]),
				depth: depth,
			})
		default:
			_, size := utf8.DecodeRuneInString(query[pos:])
			pos += size
		}
	}
	if depth != 0 {
		return nil, errors.New("unbalanced MSSQL statement")
	}
	return tokens, nil
}

func mssqlWordSize(text string) int {
	value, size := utf8.DecodeRuneInString(text)
	if value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' || value == '_' || value == '@' || value == '#' || value == '$' || value >= utf8.RuneSelf && !unicode.IsSpace(value) {
		return size
	}
	return 0
}
