package macros

import "fmt"

// InvalidMacroValueError represents an error when a value is invalid for a macro function
type InvalidMacroValueError struct {
	Column string
	Detail string
}

func (e InvalidMacroValueError) Error() string {
	return fmt.Sprintf("expected a valid macro value for column '%s' : [%s]", e.Column, e.Detail)
}

// InvalidMacroValueError represents an error when a value is invalid for a macro function
type MacroNotImplemented struct {
	Column    string
	MacroName string
}

func (e MacroNotImplemented) Error() string {
	return fmt.Sprintf("This macro [%s] was not implemented column '%s'", e.MacroName, e.Column)
}
