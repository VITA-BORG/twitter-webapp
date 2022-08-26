package validation

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var EmailEXP = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

//Validator is a struct that contains a map of validation errors.
type Validator struct {
	FieldErrors map[string]string
}

//Valid returns true if there are no validation errors in the FieldErrors
func (v *Validator) Valid() bool {
	return len(v.FieldErrors) == 0
}

//AddFieldError adds a error message to the FieldErrors map.
func (v *Validator) AddFieldError(field, message string) {
	//creates an instance of the map if it is nil
	if v.FieldErrors == nil {
		v.FieldErrors = make(map[string]string)
	}

	if _, exists := v.FieldErrors[field]; !exists {
		v.FieldErrors[field] = message
	}
}

//CheckField adds and error message to the FieldErrors map if a validation check fails
func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFieldError(key, message)
	}
}

//NotEmpty checks if a string is not empty
func NotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}

//MaxCharacters checks if a string is less than or equal to a given number of characters
func MaxCharacters(s string, n int) bool {
	return len(s) <= n
}

//ValidInt checks if a string is a valid integer
func ValidInt(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

//PermittedDate checks if a string is a valid date
func PermittedDate(s string) bool {
	_, err := time.Parse("2006-01-02", s)
	return err == nil
}

//MinChars returns true if a value contains at least n number of characters
func MinChars(s string, n int) bool {
	return utf8.RuneCountInString(s) >= n
}

//Matches returns true if a value matches a regular expression
func Matches(s string, re *regexp.Regexp) bool {
	return re.MatchString(s)
}
