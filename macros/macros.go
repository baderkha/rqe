package macros

import (
	"time"
)

var (
	Supported = []string{
		"age",
	}
)

var (
	Handlers map[string]Macro = map[string]Macro{
		"age": &AgeMacro{
			Format: time.DateTime,
		},
	}
)

type Macro interface {
	RunMacro(col string, args ...any) (arg []any, err error)
}
