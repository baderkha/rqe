package rqe

import (
	"fmt"
)

// Custom error types
type ParseError interface {
	Error() string
	Position() (int, int) // Returns line and column position
}

// InvalidColumnError represents an error when an invalid column is used
type InvalidColumnError struct {
	Column string
	Line   int
	Pos    int
}

func (e InvalidColumnError) Error() string {
	return fmt.Sprintf("invalid column '%s' at line %d, offset %d", e.Column, e.Line, e.Pos)
}

func (e InvalidColumnError) Position() (int, int) {
	return e.Line, e.Pos
}

// UnexpectedTokenError represents an error when an unexpected token appears
type UnexpectedTokenError struct {
	Token string
	Line  int
	Pos   int
}

func (e UnexpectedTokenError) Error() string {
	return fmt.Sprintf("unexpected token '%s' at line %d, offset %d", e.Token, e.Line, e.Pos)
}

// UnexpectedTokenError represents an error when an unexpected token appears
type LogicalTokenError struct {
	Reason string
	Line   int
	Pos    int
}

func (e LogicalTokenError) Error() string {
	return fmt.Sprintf("unexpected logical operation due to ['%s'] at line %d, offset %d", e.Reason, e.Line, e.Pos)
}

func (e LogicalTokenError) Position() (int, int) {
	return e.Line, e.Pos
}

// MissingValueError represents an error when a value is missing after an operation
type MissingValueError struct {
	Column string
	Line   int
	Pos    int
}

func (e MissingValueError) Error() string {
	return fmt.Sprintf("expected a valid value for column '%s' at line %d, offset %d", e.Column, e.Line, e.Pos)
}

func (e MissingValueError) Position() (int, int) {
	return e.Line, e.Pos
}

// InvalidOperationError represents an error when an invalid operation is used
type InvalidOperationError struct {
	Operation string
	Column    string
	Line      int
	Pos       int
}

func (e InvalidOperationError) Error() string {
	return fmt.Sprintf("invalid equality operation '%s' for column '%s' at line %d, offset %d", e.Operation, e.Column, e.Line, e.Pos)
}

func (e InvalidOperationError) Position() (int, int) {
	return e.Line, e.Pos
}

// UnmatchedParenthesisError represents an error for unmatched parentheses
type UnmatchedParenthesisError struct {
	Type string // "opening" or "closing"
	Line int
	Pos  int
}

func (e UnmatchedParenthesisError) Error() string {
	return fmt.Sprintf("unmatched %s parenthesis at line %d, offset %d", e.Type, e.Line, e.Pos)
}

func (e UnmatchedParenthesisError) Position() (int, int) {
	return e.Line, e.Pos
}
