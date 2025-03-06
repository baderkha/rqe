package rqe

import (
	"fmt"
	"reflect"
	"time"
)

var (
	SupportedMacros = []string{
		"age",
	}
)

var (
	MacroHandlers map[string]Macro = map[string]Macro{
		"age": &AgeMacro{
			Format: time.DateTime,
		},
	}
)

type Macro interface {
	RunMacro(col string, args ...any) (arg []any, err error)
}

var _ Macro = &AgeMacro{}

type AgeMacro struct {
	Format string
}

func (a *AgeMacro) RunMacro(col string, args ...any) (arg []any, err error) {
	arg = make([]any, 0)
	for _, v := range args {
		var newVal int64 = 0
		switch v.(type) {
		case int:
		case int16:
		case int32:
		case float32:
		case float64:
		case int64:
			val, ok := v.(int64)
			if !ok {
				return nil, &InvalidMacroValueError{Column: col, Detail: fmt.Sprintf("%v of type [%v] cannot be casted into an integer", v, reflect.TypeOf(v))}
			}
			newVal = val

		default:

			return nil, &InvalidMacroValueError{Column: col, Detail: fmt.Sprintf("%v of type [%v] cannot be casted into an integer", v, reflect.TypeOf(v))}

		}
		t := time.Now().AddDate(int(-1*newVal), 0, 0).Format(a.Format)
		arg = append(arg, t)
	}
	return arg, nil
}
