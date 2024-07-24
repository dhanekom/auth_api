package main

import (
	"strconv"
)

type EnvReader struct {
	reader func(string) string
}

func NewEnvReader(r func(string) string) *EnvReader {
	return &EnvReader{
		reader: r,
	}
}

func (r EnvReader) GetString(key string, defaultValue ...string) string {
	value := r.reader(key)
	if len(defaultValue) > 0 && value == "" {
		return defaultValue[0]
	}
	return value
}

func (r EnvReader) GetInt(key string, defaultValue ...int) int {
	value, err := strconv.Atoi(r.reader(key))
	if err != nil {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}
	return value
}

// func (r EnvReader) GetBool(key string, defaultValue ...bool) bool {
// 	value, err := strconv.ParseBool(r.reader(key))
// 	if err != nil {
// 		if len(defaultValue) > 0 {
// 			return defaultValue[0]
// 		}
// 		return false
// 	}
// 	return value
// }
