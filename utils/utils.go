// Email validation utility package
package utils

import "regexp"

// ValidateEmail returns true if s looks like a valid email address.
func ValidateEmail(s string) bool {
	// RFC-5322-ish regex
	const pattern = `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`

	// regexp.MustCompile compiles the regular expression pattern.
	// It will panic if the pattern is invalid, which is acceptable here since the pattern is a constant.
	re := regexp.MustCompile(pattern)

	// re.MatchString checks if the input string `s` matches the compiled regular expression.
	return re.MatchString(s)
}
