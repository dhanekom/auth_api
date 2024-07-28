package validator

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	EmailRegexStr string = "^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
)

var (
	EmailRegex = regexp.MustCompile(EmailRegexStr)
)

type Validator struct {
	keys      []string
	ErrorList map[string]string
}

// Error generates a error string from all the added errors
func (v Validator) Error() string {
	sb := strings.Builder{}
	delim := ""

	for _, key := range v.keys {
		if value, ok := v.ErrorList[key]; ok {
			sb.WriteString(fmt.Sprintf("%s%s: %s", delim, key, value))
			delim = ", "
		}
	}

	return sb.String()
}

// Valid returns true if there are no errors
func (v *Validator) Valid() bool {
	return len(v.ErrorList) == 0
}

// AddError adds an error to ErrorList if no errors have been added yet for argument "key"
func (v *Validator) AddError(key, message string) {
	if v.ErrorList == nil {
		v.ErrorList = make(map[string]string)
	}

	if _, exists := v.ErrorList[key]; !exists {
		v.keys = append(v.keys, key)
		v.ErrorList[key] = message
	}
}

// CheckValue can be used to check an expression (argument "ok") and adds an error if "ok" == false
func (v *Validator) CheckValue(ok bool, key string, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// CheckRequired checks that a required key is present
func (v *Validator) CheckRequired(value string, key string) {
	if !NotBlank(value) {
		v.AddError(key, "required")
	}
}

// NonBlank checks that the trimmed value is not an emptry string
func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

// // MaxLength checks that the number of runes (characters) in value is less or equal to n
// func MaxLength(value string, n int) bool {
// 	return utf8.RuneCountInString(value) <= n
// }

// // MinLength checks that the number of runes (characters) in value is more or equal to n
// func MinLength(value string, n int) bool {
// 	return utf8.RuneCountInString(value) <= n
// }

// // BetweenLength checks that the length of value is between min and mx
// func BetweenLength(value string, min, max int) bool {
// 	return utf8.RuneCountInString(value) >= min && utf8.RuneCountInString(value) <= max
// }

// func PermittedValue[T comparable](value T, permittedValues ...T) bool {
// 	return slices.Contains(permittedValues, value)
// }

func IsEmail(value string) bool {
	return EmailRegex.MatchString(value)
}
