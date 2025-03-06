package rqe

import (
	"fmt"
	"strings"
)

func DANGEROUS_DEBUG_COMPILE_SQL(query string, args []interface{}) string {
	var sb strings.Builder
	argIndex := 0

	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			if argIndex >= len(args) {
				sb.WriteString("?") // Leave as `?` if no argument exists
				continue
			}

			// Format argument based on its type
			switch v := args[argIndex].(type) {
			case string:
				sb.WriteString(fmt.Sprintf("'%s'", v)) // Wrap strings in quotes
			case int, int32, int64, float32, float64:
				sb.WriteString(fmt.Sprintf("%v", v)) // Directly append numeric values
			default:
				sb.WriteString(fmt.Sprintf("'%v'", v)) // Default case
			}
			argIndex++
		} else {
			sb.WriteByte(query[i])
		}
	}

	return sb.String()
}
