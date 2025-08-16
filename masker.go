package masker

import (
	"reflect"
	"strings"
)

// Mask applies masking based on struct tags like `mask:"group:PII;strategy:partial"`
func Mask(v interface{}) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		value := rv.Field(i)

		tag := field.Tag.Get("mask")
		if tag == "" || !value.CanSet() {
			continue
		}

		// Example: simple strategy
		if strings.Contains(tag, "strategy:partial") && value.Kind() == reflect.String {
			str := value.String()
			if len(str) > 4 {
				masked := str[:2] + strings.Repeat("*", len(str)-4) + str[len(str)-2:]
				value.SetString(masked)
			}
		}

		if strings.Contains(tag, "strategy:full") && value.Kind() == reflect.String {
			masked := strings.Repeat("*", len(value.String()))
			value.SetString(masked)
		}
	}
}
