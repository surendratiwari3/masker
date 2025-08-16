package masker

import (
	"reflect"
	"strings"
)

// MaskFunc is a strategy function type
type MaskFunc func(string) string

// Global registry of masking strategies (write once at init)
var maskRegistry = map[string]MaskFunc{}

// RegisterMaskFunc lets users register custom strategies (call during init only)
func RegisterMaskFunc(name string, fn MaskFunc) {
	maskRegistry[name] = fn
}

// MaskOverrides allows runtime override of masking rules
type MaskOverrides map[string]string
// key = struct field name
// value = strategy name ("none" = skip masking, or any registered strategy)

// ---------------------- Initialize default strategies ----------------------
func init() {
	// Basic reusable strategies
	maskRegistry["partial"] = func(s string) string {
		if len(s) > 4 {
			return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
		}
		return strings.Repeat("*", len(s))
	}
	maskRegistry["full"] = func(s string) string { return strings.Repeat("*", len(s)) }
	maskRegistry["email"] = func(s string) string {
		parts := strings.Split(s, "@")
		if len(parts) == 2 {
			return parts[0][:1] + strings.Repeat("*", len(parts[0])-1) + "@" + parts[1]
		}
		return strings.Repeat("*", len(s))
	}
	maskRegistry["phone"] = func(s string) string {
		if len(s) > 4 {
			return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
		}
		return strings.Repeat("*", len(s))
	}
	maskRegistry["creditcard"] = func(s string) string {
		if len(s) > 4 {
			return strings.Repeat("*", len(s)-4) + s[len(s)-4:]
		}
		return strings.Repeat("*", len(s))
	}
	maskRegistry["dob"] = func(s string) string {
		if len(s) == 10 {
			return "****-**-" + s[len(s)-2:]
		}
		return strings.Repeat("*", len(s))
	}
	maskRegistry["password"] = func(s string) string { return strings.Repeat("*", len(s)) }
	maskRegistry["token"] = func(s string) string { return strings.Repeat("*", len(s)) }

	// Group strategies
	maskRegistry["PII"] = maskRegistry["partial"]
	maskRegistry["PHI"] = maskRegistry["dob"]
	maskRegistry["PCI"] = maskRegistry["creditcard"]
	maskRegistry["CREDENTIALS"] = maskRegistry["full"]
	maskRegistry["FINANCIAL"] = maskRegistry["partial"]
	maskRegistry["GDPR"] = maskRegistry["full"]
	maskRegistry["none"] = func(s string) string { return s }
}

// ---------------------- Public Functions ----------------------

// Mask applies masking based on struct tags (default behavior)
func Mask(v interface{}) {
	maskValue(reflect.ValueOf(v), nil)
}

// MaskWithOverrides applies masking with runtime overrides
func MaskWithOverrides(v interface{}, overrides MaskOverrides) {
	maskValue(reflect.ValueOf(v), overrides)
}

// MaskCopy returns a masked copy of the struct, keeping original safe
func MaskCopy(v interface{}, overrides MaskOverrides) interface{} {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return v
	}
	copyVal := reflect.New(rv.Elem().Type())
	copyVal.Elem().Set(rv.Elem())
	maskValue(copyVal, overrides)
	return copyVal.Interface()
}

// ---------------------- Internal recursive functions ----------------------
func maskValue(rv reflect.Value, overrides MaskOverrides) {
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
			fieldName := field.Name

			// ---------------- Runtime override ----------------
			if overrides != nil {
				if strategy, ok := overrides[fieldName]; ok {
					if strategy == "none" {
						continue
					}
					if fn, exists := maskRegistry[strategy]; exists && value.Kind() == reflect.String {
						value.SetString(fn(value.String()))
						continue
					}
				}
			}

			// ---------------- Tag-based default ----------------
			tag := field.Tag.Get("mask")
			if tag != "" && value.Kind() == reflect.String {
				parts := strings.Split(tag, ";")
				for _, part := range parts {
					if strings.HasPrefix(part, "strategy:") {
						strat := strings.TrimPrefix(part, "strategy:")
						if fn, ok := maskRegistry[strat]; ok {
							value.SetString(fn(value.String()))
						}
					}
				}
			} else {
				maskValue(value, overrides)
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			maskValue(rv.Index(i), overrides)
		}
	case reflect.Map:
		for _, key := range rv.MapKeys() {
			val := rv.MapIndex(key)
			maskValue(val, overrides)
		}
	}
}

