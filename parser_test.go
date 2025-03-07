package rqe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to validate column names
func validateColumn(col string) bool {
	return true
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
	sqlInjectionTests := []string{
		"username eq ' OR '1'='1'",
		"password ne ' OR '1'='1'",
		"age lt ' OR '1'='1'",
		"salary lte ' OR '1'='1'",
		"score gt ' OR '1'='1'",
		"rank gte ' OR '1'='1'",
		"status in ['admin', ' OR '1'='1']",
		"date between ' OR '1'='1' AND '2025-12-31'",

		"username eq '; DROP TABLE users; --'",
		"password ne '; DROP TABLE users; --'",
		"age lt '; DROP TABLE users; --'",
		"salary lte '; DROP TABLE users; --'",
		"score gt '; DROP TABLE users; --'",
		"rank gte '; DROP TABLE users; --'",
		"status in ['admin', '; DROP TABLE users; --']",
		"date between '; DROP TABLE users; --' AND '2025-12-31'",

		"(username eq 'admin' and password eq ' OR '1'='1')",
		"(email eq ' OR EXISTS(SELECT * FROM users WHERE username = 'admin') and status ne 'banned')",

		"username eq ' OR SLEEP(10)--'",
		"password ne ' OR SLEEP(10)--'",
		"age lt ' OR SLEEP(10)--'",
		"salary lte ' OR SLEEP(10)--'",
		"score gt ' OR SLEEP(10)--'",
		"rank gte ' OR SLEEP(10)--'",
		"status in ['admin', ' OR SLEEP(10)--']",
		"date between ' OR SLEEP(10)--' AND '2025-12-31'",

		"(username eq 'admin' and (password eq ' OR SLEEP(10)--' or rank gt 5))",
		"(email eq ' OR (1=1) and status ne 'inactive')",

		"username eq ' UNION SELECT NULL, NULL, NULL--'",
		"password ne ' UNION SELECT NULL, NULL, NULL--'",
		"age lt ' UNION SELECT NULL, NULL, NULL--'",
		"salary lte ' UNION SELECT NULL, NULL, NULL--'",
		"score gt ' UNION SELECT NULL, NULL, NULL--'",
		"rank gte ' UNION SELECT NULL, NULL, NULL--'",
		"status in ['admin', ' UNION SELECT NULL, NULL, NULL--']",
		"date between ' UNION SELECT NULL, NULL, NULL--' AND '2025-12-31'",

		"(username eq 'admin' and (password eq ' UNION SELECT NULL, NULL, NULL--' or rank gt 5))",
		"(email eq ' UNION SELECT version(), user()--' and status ne 'inactive')",
	}

	advCases := []string{
		// Basic injections with nested conditions
		"(username eq 'admin' and password eq ' OR '1'='1')",
		"(email eq ' OR EXISTS(SELECT * FROM users WHERE username = 'admin') and status ne 'banned')",

		// **Unicode Homoglyph Attack**
		"username eq 'ＯＲ '1'＝'1'",    // Full-width characters
		"email eq '𝓞𝓡 1=1 --'",       // Mathematical script letters
		"password eq 'ＯＲ 1＝1#'",      // Full-width OR and =
		"username eq 'or%ef%bc%9d1'", // URL-encoded full-width '='

		// **Comment Exploits**
		"username eq 'admin'--'",
		"password eq 'admin' #'",
		"status in ['admin', 'pending' --']",
		"age gt 20; -- DROP TABLE users;",

		// **Hex Encoding & ASCII**
		"username eq 0x61646d696e",              // `admin` in hex
		"password eq unhex('70617373776f7264')", // `password` in hex
		"email eq 'admin' /*!50000 UNION */ SELECT null, null, null --",

		// **Whitespace & Case Bypass**
		"username eq 'a'/**/OR/**/'1'='1'",
		"password eq 'a' --%0aDROP TABLE users;",
		"email eq ' ' OR 1=1 /*'--*/",
		"score gte 10; /**/DROP TABLE users/**/--",

		// **Subquery Injection**
		"(username eq 'admin' or (SELECT COUNT(*) FROM users) > 0)",
		"(email eq ' OR (SELECT database()) IS NOT NULL)",
		"(status in ['active', (SELECT column_name FROM information_schema.columns WHERE table_name='users')])",

		// **Time-Based Blind SQL Injection**
		"(username eq 'admin' and password eq ' OR SLEEP(10)')",
		"(email eq ' OR pg_sleep(5)')",
		"(status in ['active', ' OR BENCHMARK(5000000,MD5(1))'])",
		"username eq ' OR CASE WHEN (1=1) THEN SLEEP(5) ELSE 1 END --",

		// **Nested Logical Operations**
		"((username eq 'admin') and ((password eq ' OR '1'='1') or (role eq 'superadmin')))",
		"((email eq ' OR (SELECT COUNT(*) FROM users) > 0) or (status eq 'banned'))",
		"(username eq 'admin' and (password eq ' OR (SELECT version())' or status in ['active', ' OR (SELECT user())']))",

		// **Bypassing Length Restrictions**
		"(username eq 'admin' and password eq 'a' or 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'='aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa')",
		"(email eq ' OR 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'='aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa')",
		"(status in ['active', 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'='aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'])",

		// **Boolean Logic Exploits**
		"(username eq 'admin' and password eq ' OR NOT 0')",
		"(email eq ' OR NOT NULL IS NULL')",
		"(status in ['active', ' OR NULL IS NOT NULL'])",

		// **Error-Based SQL Injection**
		"(username eq ' OR 1=1 UNION SELECT 1,2,@@version)",
		"(password eq ' OR 1=1 UNION SELECT NULL, NULL, database())",
		"(email eq ' OR 1=1 UNION SELECT table_name FROM information_schema.tables)",
		"(status in ['active', ' OR (SELECT COUNT(*) FROM users)])",
	}

	tests = append(tests, sqlInjectionTests...)
	tests = append(tests, advCases...)

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			sql, _ := Parse(test, validateColumn)
			assert.NotContains(t, sql.SQL, "DROP", "SQL should not contain DROP")
			assert.NotContains(t, sql.SQL, "DELETE", "SQL should not contain DELETE")
			assert.NotContains(t, sql.SQL, "--", "SQL should not allow comment injection")
			assert.NotContains(t, sql.SQL, " OR ", "SQL should not allow OR-based injection")
			assert.NotContains(t, sql.SQL, ";", "SQL should not allow statement termination")
			assert.NotContains(t, sql.SQL, "UNION", "no union")

			if sql.SQL == "" {
				// t.Logf("✅ Query parsed safely: [NO QUERY]")
			} else {
				t.Logf("✅ Query parsed safely: [%s]", sql.SQL)
			}

		})
	}
}
