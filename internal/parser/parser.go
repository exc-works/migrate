package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

type Direction bool

const (
	DirectionUp   Direction = true
	DirectionDown Direction = false
)

const sqlCmdPrefix = "-- +migrate "

func SplitSQLStatements(r io.Reader, direction Direction) ([]string, error) {
	var buf bytes.Buffer
	scanner := bufio.NewScanner(r)

	upSections := 0
	downSections := 0
	statementEnded := false
	ignoreSemicolons := false
	directionIsActive := false
	statements := make([]string, 0)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, sqlCmdPrefix) {
			cmd := strings.TrimSpace(line[len(sqlCmdPrefix):])
			switch cmd {
			case "Up":
				directionIsActive = direction == DirectionUp
				upSections++
			case "Down":
				directionIsActive = direction == DirectionDown
				downSections++
			case "StatementBegin":
				if directionIsActive {
					ignoreSemicolons = true
				}
			case "StatementEnd":
				if directionIsActive {
					statementEnded = ignoreSemicolons
					ignoreSemicolons = false
				}
			default:
				return nil, fmt.Errorf("unknown migrate command: %q", cmd)
			}
		}

		if !directionIsActive {
			continue
		}

		if _, err := buf.WriteString(line + "\n"); err != nil {
			return nil, err
		}

		if (!ignoreSemicolons && endsWithSemicolon(line)) || statementEnded {
			statementEnded = false
			stmt := strings.TrimSpace(buf.String())
			if stmt != "" {
				statements = append(statements, stmt)
			}
			buf.Reset()
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan migration: %w", err)
	}

	if upSections == 0 && downSections == 0 {
		return nil, fmt.Errorf("no Up/Down annotations found")
	}
	if ignoreSemicolons {
		return nil, fmt.Errorf("found StatementBegin with no matching StatementEnd")
	}
	if remaining := strings.TrimSpace(buf.String()); remaining != "" {
		return nil, fmt.Errorf("unfinished SQL query in active section, missing semicolon")
	}

	return statements, nil
}

func endsWithSemicolon(line string) bool {
	prev := ""
	scanner := bufio.NewScanner(strings.NewReader(line))
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := scanner.Text()
		if strings.HasPrefix(word, "--") {
			break
		}
		prev = word
	}
	return strings.HasSuffix(prev, ";")
}
