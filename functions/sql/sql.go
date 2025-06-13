package sql

import (
	"regexp"
	"strings"

	"github.com/jsavajols/goframework/functions/logs"
)

var SuspiciousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`1=1`), // SQL comment
	regexp.MustCompile(`#`),   // SQL comment
	regexp.MustCompile(`--`),  // SQL comment
	regexp.MustCompile(`/\*`), // SQL multi-line comment
	regexp.MustCompile(`;`),   // Statement terminator
}

func DateDiff(unit, date1, date2, dialect string) string {
	if dialect == "mysql" {
		return "DATEDIFF(" + date1 + "," + date2 + ")"
	} else if dialect == "postgres" {
		return "DATE_PART('" + unit + "'," + date1 + "::timestamp - " + date2 + "::timestamp)"
	} else if dialect == "sqlite3" {
		return "JULIANDAY(" + date1 + ") - JULIANDAY(" + date2 + ")"
	} else {
		return ""
	}
}

// CheckForSQLInjection inspects a string for potential SQL injection patterns
func CheckForSQLInjection(input string) bool {
	input = strings.TrimSpace(input)
	for _, pattern := range SuspiciousPatterns {
		if pattern.MatchString(input) {
			logs.Logs("Suspicious pattern found: ", pattern.String())
			return true
		}
	}
	return false
}
