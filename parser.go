package rqe

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bzick/tokenizer"
)

const (
	TEquality = iota + 1
	TMath
	TDoubleQuoted
	TLogicalOperation
	TParen
	TArray
)

type OperationMeta struct {
	Value           func(quotes int) string
	IsMultiValue    bool
	MultiValueLimit int
}

type ParsedQuery struct {
	SQL  string
	Args []interface{}
}

var operationsMapped = map[string]OperationMeta{
	"lt": {
		Value:        func(_ int) string { return "< ?" },
		IsMultiValue: false,
	},
	"lte": {
		Value:        func(_ int) string { return "<= ?" },
		IsMultiValue: false,
	},
	"eq": {
		Value:        func(_ int) string { return "= ?" },
		IsMultiValue: false,
	},
	"gte": {
		Value:        func(_ int) string { return ">= ?" },
		IsMultiValue: false,
	},
	"gt": {
		Value:        func(_ int) string { return "> ?" },
		IsMultiValue: false,
	},
	"ne": {
		Value:        func(_ int) string { return "<> ?" },
		IsMultiValue: false,
	},
	"in": {
		Value: func(quotes int) string {
			placeholders := make([]string, quotes)
			for i := range placeholders {
				placeholders[i] = "?"
			}
			return fmt.Sprintf("IN (%s)", strings.Join(placeholders, ", "))
		},
		IsMultiValue: true,
	},
	"between": {
		Value:        func(_ int) string { return "BETWEEN ? AND ?" },
		IsMultiValue: true, MultiValueLimit: 2,
	},
}

func Parse(filter string, validateCol func(col string) bool) (ParsedQuery, error) {
	var sb strings.Builder
	vals := make([]interface{}, 0)

	// Configure tokenizer
	parser := tokenizer.New()
	parser.DefineTokens(TEquality, []string{"lt", "lte", "eq", "gte", "gt", "ne", "in", "between"})
	parser.DefineTokens(TLogicalOperation, []string{"and", "or"})
	parser.DefineStringToken(TDoubleQuoted, `"`, `"`).SetEscapeSymbol(tokenizer.BackSlash)
	parser.DefineStringToken(TDoubleQuoted, `'`, `'`).SetEscapeSymbol(tokenizer.BackSlash)
	parser.DefineStringToken(TArray, `[`, `]`).SetEscapeSymbol(tokenizer.BackSlash)

	parser.AllowKeywordSymbols(tokenizer.Underscore, tokenizer.Numbers)

	// Create tokens' stream
	stream := parser.ParseString(filter)
	defer stream.Close()

	// Stack to track nested parentheses
	var parenStack []int

	// Helper function to ensure spaces around content
	writeWithSpaces := func(content string) {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(content)
		sb.WriteString(" ")
	}

	// Iterate over each token
	for stream.IsValid() {
		line, column := stream.CurrentToken().Line(), stream.CurrentToken().Offset()
		tokenValue := stream.CurrentToken().ValueString()

		switch {
		case stream.CurrentToken().Is(tokenizer.TokenKeyword):
			col := tokenValue
			quotesNeeded := 1

			if !validateCol(col) {
				return ParsedQuery{}, InvalidColumnError{Column: col, Line: line, Pos: column}
			}

			if !stream.GoNextIfNextIs(TEquality) {
				return ParsedQuery{}, UnexpectedTokenError{Token: "equality operation", Line: line, Pos: column + len(col)}
			}

			opValue := stream.CurrentToken().ValueString()
			op, foundOp := operationsMapped[opValue]
			if !foundOp {
				return ParsedQuery{}, InvalidOperationError{Operation: opValue, Column: col, Line: line, Pos: column + len(col)}
			}

			if !stream.GoNextIfNextIs(tokenizer.TokenFloat, tokenizer.TokenInteger, tokenizer.TokenString) {
				return ParsedQuery{}, MissingValueError{Column: col, Line: line, Pos: column + len(col) + len(opValue)}
			}

			switch {
			case stream.CurrentToken().IsFloat():
				vals = append(vals, stream.CurrentToken().ValueFloat64())
			case stream.CurrentToken().IsInteger():
				vals = append(vals, stream.CurrentToken().ValueInt64())
			case stream.CurrentToken().IsString():
				if stream.CurrentToken().StringKey() == TArray {
					if !op.IsMultiValue {
						return ParsedQuery{}, InvalidOperationError{Operation: "multi-value array", Column: col, Line: line, Pos: column}
					}

					var value []interface{}
					err := json.Unmarshal([]byte(stream.CurrentToken().ValueString()), &value)
					if err != nil {
						return ParsedQuery{}, UnexpectedTokenError{Token: "invalid array argument", Line: line, Pos: column}
					}
					quotesNeeded = len(value)
					if len(value) == 0 {
						return ParsedQuery{}, InvalidOperationError{Operation: "multi-value array empty arguments", Column: col, Line: line, Pos: column}
					}
					vals = append(vals, value...)
				} else {
					strVal := stream.CurrentToken().ValueString()
					vals = append(vals, strVal[1:len(strVal)-1]) // Strip quotes
				}
			}

			writeWithSpaces(fmt.Sprintf("%s %s", col, op.Value(quotesNeeded)))

		case stream.CurrentToken().Is(TLogicalOperation):
			writeWithSpaces(tokenValue)

		case tokenValue == "(":
			if !stream.NextToken().Is(tokenizer.TokenKeyword) {
				return ParsedQuery{}, UnexpectedTokenError{Token: "expression", Line: line, Pos: column}
			}
			writeWithSpaces("(" + "")
			parenStack = append(parenStack, len(vals)) // Track nested position

		case tokenValue == ")":
			if len(parenStack) == 0 {
				return ParsedQuery{}, UnmatchedParenthesisError{Type: "closing", Line: line, Pos: column}
			}
			parenStack = parenStack[:len(parenStack)-1] // Pop from stack
			writeWithSpaces(") ")

		default:
			return ParsedQuery{}, UnexpectedTokenError{Token: tokenValue, Line: line, Pos: column}
		}

		stream.GoNext()
	}

	// If the stack is not empty, we have unclosed parentheses
	if len(parenStack) > 0 {
		return ParsedQuery{}, UnmatchedParenthesisError{Type: "opening", Line: 0, Pos: 0}
	}

	return ParsedQuery{SQL: strings.TrimSpace(sb.String()), Args: vals}, nil
}
