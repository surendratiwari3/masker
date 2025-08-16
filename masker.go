package masker

import (
	"reflect"
	"strings"
)

// MaskFunc is a strategy function type
type MaskFunc func(string) string

// Global registry of masking strategies
var maskRegistry = map[string]MaskFunc{}

// RegisterMaskFunc lets users register custom strategies
func RegisterMaskFunc(name string, fn MaskFunc) {
	maskRegistry[name] = fn
}

// init loads default strategies and groups
func init() {
	// ---------------- Basic reusable strategies ----------------
	RegisterMaskFunc("partial", func(s string) string {
		if len(s) > 4 {
			return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
		}
		return strings.Repeat("*", len(s))
	})
	RegisterMaskFunc("full", func(s string) string {
		return strings.Repeat("*", len(s))
	})
	RegisterMaskFunc("email", func(s string) string {
		parts := strings.Split(s, "@")
		if len(parts) == 2 {
			return parts[0][:1] + strings.Repeat("*", len(parts[0])-1) + "@" + parts[1]
		}
		return strings.Repeat("*", len(s))
	})
	RegisterMaskFunc("phone", func(s string) string {
		if len(s) > 4 {
			return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
		}
		return strings.Repeat("*", len(s))
	})
	RegisterMaskFunc("creditcard", func(s string) string {
		if len(s) > 4 {
			return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
		}
		return strings.Repeat("*", len(s))
	})
	RegisterMaskFunc("dob", func(s string) string {
		if len(s) == 10 {
			return "****-**-" + s[len(s)-2:]
		}
		return strings.Repeat("*", len(s))
	})
	RegisterMaskFunc("password", func(s string) string { return strings.Repeat("*", len(s)) })
	RegisterMaskFunc("token", func(s string) string { return strings.Repeat("*", len(s)) })

	// ---------------- Group strategies ----------------
	RegisterMaskFunc("PII", func(s string) string { return maskRegistry["partial"](s) })
	RegisterMaskFunc("PHI", func(s string) string { return maskRegistry["dob"](s) })
	RegisterMaskFunc("PCI", func(s string) string { return maskRegistry["creditcard"](s) })
	RegisterMaskFunc("CREDENTIALS", func(s string) string { return maskRegistry["full"](s) })
	RegisterMaskFunc("FINANCIAL", func(s string) string { return maskRegistry["partial"](s) })
	RegisterMaskFunc("GDPR", func(s string) string { return maskRegistry["full"](s) })
	RegisterMaskFunc("none", func(s string) string { return s }) // explicit no-mask
}

// Mask applies masking based on struct tags (field/group/strategy)
func Mask(v interface{}) {
	maskValue(reflect.ValueOf(v))
}

// internal recursive function to handle nested structs, slices, and maps
func maskValue(rv reflect.Value) {
	if !rv.IsValid() {
		return
	}

	if rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		rt := rv.Type()
		for i := 0; i < rv.NumField(); i++ {
			field := rt.Field(i)
			value := rv.Field(i)
			if !value.CanSet() {
				continue
			}

			tag := field.Tag.Get("mask")
			if tag != "" && value.Kind() == reflect.String {
				parts := strings.Split(tag, ";")
				for _, part := range parts {
					if strings.HasPrefix(part, "strategy:") {
						strategy := strings.TrimPrefix(part, "strategy:")
						if fn, ok := maskRegistry[strategy]; ok {
							value.SetString(fn(value.String()))
						}
					}
				}
			} else {
				// recurse nested struct
				maskValue(value)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			maskValue(rv.Index(i))
		}
	case reflect.Map:
		for _, key := range rv.MapKeys() {
			val := rv.MapIndex(key)
			maskValue(val)
		}
	}
}

