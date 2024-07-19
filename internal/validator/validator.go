package validator

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	EmailRegexStr string = "^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
)

var (
	EmailRegex = regexp.MustCompile(EmailRegexStr)
)

type Validator struct {
	ErrorList map[string]string
}

func (v Validator) Error() string {
	sb := strings.Builder{}
	delim := ""
	for k, v := range v.ErrorList {
		sb.WriteString(fmt.Sprintf("%s%s: %s", delim, k, v))
		delim = ", "
	}
	return sb.String()
}

func (v *Validator) Valid() bool {
	return len(v.ErrorList) == 0
}

func (v *Validator) AddError(key, message string) {
	if v.ErrorList == nil {
		v.ErrorList = make(map[string]string)
	}

	if _, exists := v.ErrorList[key]; !exists {
		v.ErrorList[key] = message
	}
}

func (v *Validator) CheckValue(ok bool, key string, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func (v *Validator) CheckRequired(value string, key string) {
	if !NotBlank(value) {
		v.AddError(key, "required")
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxLength(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func MinLength(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

func BetweenLength(value string, min, max int) bool {
	return utf8.RuneCountInString(value) >= min && utf8.RuneCountInString(value) <= max
}

func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

func IsEmail(value string) bool {
	return EmailRegex.MatchString(value)
}
