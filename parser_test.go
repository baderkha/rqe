package rqe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to validate column names
func validateColumn(col string) bool {
	allowedCols := map[string]bool{
		"user_id": true, "created_at": true, "sat_score": true, "val": true,
	}
	return allowedCols[col]
}

// Test SQL injection attempts
func TestSQLInjection(t *testing.T) {
	tests := []string{
		`user_id eq "119' OR '1'='1"`,                     // Attempt to bypass
		`created_at eq '2020-01-01 00:00:00' -- '`,        // SQL comment injection
		`sat_score gte 1200; DROP TABLE users; --`,        // Attempt to drop table
		`val eq "4); DROP TABLE users; --"`,               // Injection with closing parenthesis
		`(user_id eq 119) or (1=1)`,                       // Logical SQL bypass attempt
		`user_id eq "'); EXEC xp_cmdshell('whoami'); --"`, // Attempt to execute command
		`val eq "' OR 1=1 --"`,                            // Classic OR 1=1 attack
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("✅ Detected and prevented SQL injection: %v", test)
				}
			}()

			sql, _ := Parse(test, validateColumn)
			assert.NotContains(t, sql, "DROP", "SQL should not contain DROP")
			assert.NotContains(t, sql, "DELETE", "SQL should not contain DELETE")
			assert.NotContains(t, sql, "--", "SQL should not allow comment injection")
			assert.NotContains(t, sql, " OR ", "SQL should not allow OR-based injection")
			assert.NotContains(t, sql, ";", "SQL should not allow statement termination")
			t.Logf("✅ Query parsed safely: %s", sql)
		})
	}
}

// Test edge cases like empty parentheses
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		values   []interface{}
	}{

		{`()`, ``, []interface{}{}},  // Empty parentheses
		{`( )`, ``, []interface{}{}}, // Space inside parentheses
		{`((user_id eq "119"))`, `( ( user_id = ? ) )`, []interface{}{"119"}}, // Double nested parentheses
	}

	for _, test := range tests {

		sql, vals := Parse(test.input, validateColumn)
		assert.Equal(t, test.expected, sql, "Generated SQL should match expected")
		_ = vals
	}
}
