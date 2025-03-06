package rqe

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bzick/tokenizer"
	"github.com/davecgh/go-spew/spew"
)

const (
	TEquality = iota + 1
	TMath
	TDoubleQuoted
	TLogicalOperation
	TParenOpen
	TParenClose
	TArray
	TMacro
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

// Parse takes a human-readable query string and converts it into a structured SQL statement
// with placeholders and corresponding argument values. It allows logical operators (`AND`, `OR`),
// comparison operators (`=`, `!=`, `>`, `<`, `>=`, `<=`), and multi-value expressions (`IN`, `BETWEEN`).
//
// The function ensures that only valid column names are used, supports nested expressions with
// parentheses, and generates a properly formatted SQL string.
//
// Errors include detailed context such as line and column numbers.
//
// Parameters:
//   - filter (string): The query string to parse. Example: `name eq "John" and age gte 25 or (city eq "New York")`.
//   - validateCol (func(string) bool): A function that validates column names. It should return `true`
//     if the column is valid, otherwise `false`.
//
// Returns:
//   - ParsedQuery: A struct containing:
//   - SQL (string): The formatted SQL query with placeholders (`?`).
//   - Args ([]interface{}): A slice containing the argument values for the SQL query.
//   - error: An error if parsing fails, including detailed error messages with line and column numbers.
//
// Example Usage:
//
//	validateCol := func(col string) bool {
//		validColumns := map[string]bool{"name": true, "age": true, "city": true, "status": true}
//		return validColumns[col]
//	}
//
//	query, err := Parse(`name eq "John" and age gte 25 or (city eq "New York" and status in ["active", "pending"])`, validateCol)
//	if err != nil {
//		fmt.Println("Error:", err)
//		return
//	}
//
//	fmt.Println("SQL:", query.SQL)
//	fmt.Println("Args:", query.Args)
//
// Example Output:
//
//	SQL:
//	name = ?
//	AND age >= ?
//	OR ( city = ? AND status IN ( ?, ? ) )
//
//	Args:
//	["John", 25, "New York", "active", "pending"]
//
// Possible Errors:
//   - InvalidColumnError: When a column name is not allowed.
//   - UnexpectedTokenError: When the query contains an unexpected token.
//   - MissingValueError: When an operation is missing a required value.
//   - InvalidOperationError: When an operation is not valid for a given context.
//   - UnmatchedParenthesisError: When there are unmatched opening or closing parentheses.
//
// Notes:
//   - Multi-value expressions (`IN`, `BETWEEN`) must have the correct number of values.
//   - Strings should be enclosed in double (`"`) or single (`'`) quotes.
//   - Arrays should be enclosed in square brackets (`[ ]`).
func Parse(filter string, validateCol func(col string) bool) (ParsedQuery, error) {
	var sb strings.Builder
	vals := make([]interface{}, 0)

	// Configure tokenizer
	parser := tokenizer.New()
	parser.DefineTokens(TEquality, []string{"lt", "lte", "eq", "gte", "gt", "ne", "in", "between"})
	parser.DefineTokens(TLogicalOperation, []string{"and", "or"})
	parser.DefineTokens(TParenOpen, []string{"("})
	parser.DefineTokens(TParenClose, []string{")"})
	parser.DefineStringToken(TDoubleQuoted, `"`, `"`).SetEscapeSymbol(tokenizer.BackSlash)
	parser.DefineStringToken(TDoubleQuoted, `'`, `'`).SetEscapeSymbol(tokenizer.BackSlash)
	parser.DefineStringToken(TArray, `[`, `]`).SetEscapeSymbol(tokenizer.BackSlash)
	parser.DefineTokens(TMacro, SupportedMacros)

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
			macroType := ""
			quotesNeeded := 1
			currentVals := []any{}

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

			if !stream.GoNextIfNextIs(tokenizer.TokenFloat, tokenizer.TokenInteger, tokenizer.TokenString, TMacro) {
				return ParsedQuery{}, MissingValueError{Column: col, Line: line, Pos: column + len(col) + len(opValue)}
			}

			// parse macro + precheck
			if stream.CurrentToken().Is(TMacro) {
				macroType = stream.CurrentToken().ValueString()
				spew.Dump(stream.NextToken().ValueString())
				if !stream.GoNextIfNextIs(TParenOpen) {
					return ParsedQuery{}, UnexpectedTokenError{Token: "Macro expressions must have opening parenthesis and closing ones", Line: line, Pos: column}
				}
				stream.GoNext()
				if !stream.NextToken().Is(TParenClose) {
					return ParsedQuery{}, UnexpectedTokenError{Token: "Macro expressions must have opening parenthesis and closing ones", Line: line, Pos: column}
				}
			}

			// value parsing logic remains the same
			switch {
			case stream.CurrentToken().IsFloat():
				currentVals = append(currentVals, stream.CurrentToken().ValueFloat64())
			case stream.CurrentToken().IsInteger():
				currentVals = append(currentVals, stream.CurrentToken().ValueInt64())
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
					currentVals = append(currentVals, value...)
				} else {
					strVal := stream.CurrentToken().ValueString()
					currentVals = append(currentVals, strVal[1:len(strVal)-1]) // Strip quotes
				}
			}

			// run macro transformation after we have a value
			if macroType != "" {
				h, ok := MacroHandlers[macroType]
				if !ok {
					return ParsedQuery{}, MacroNotImplemented{Column: col, MacroName: macroType}
				}
				transformedArgs, err := h.RunMacro(col, currentVals...)
				if err != nil {
					return ParsedQuery{}, err
				}
				currentVals = transformedArgs
				stream.GoNext().GoNext() // we did a check before so we good
			}

			writeWithSpaces(fmt.Sprintf("%s %s", col, op.Value(quotesNeeded)))
			vals = append(vals, currentVals...)
		case stream.CurrentToken().Is(TLogicalOperation):
			if stream.PrevToken().Is(TLogicalOperation) || stream.NextToken().Is(TLogicalOperation) {
				return ParsedQuery{}, &LogicalTokenError{Reason: "before or after a logical operation, you must have an expression or nested expression"}
			} else if stream.CurrentToken().Offset() == 0 {
				return ParsedQuery{}, &LogicalTokenError{Reason: "cannot start with a logical operation"}
			}
			if !stream.GoNext().IsValid() {
				return ParsedQuery{}, &LogicalTokenError{Reason: "cannot end with a logical operation"}
			}
			writeWithSpaces(tokenValue)
			continue

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
