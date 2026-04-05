package config

import (
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

// stringToCleanSliceHookFunc returns a DecodeHookFunc that converts a
// delimiter-separated string to []string (or any named slice-of-string type).
//
// Differences from the standard mapstructure.StringToSliceHookFunc:
//   - Empty input → empty slice (not []string{""}).
//   - Each segment is trimmed of leading/trailing whitespace.
//   - Empty segments after trimming are dropped ("a,,b" → ["a","b"]).
//   - Works with named string slice types (e.g. type Tags []string).
func stringToCleanSliceHookFunc(sep string) mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data any) (any, error) {
		if f.Kind() != reflect.String ||
			t.Kind() != reflect.Slice || t.Elem().Kind() != reflect.String {
			return data, nil
		}

		s, ok := data.(string)
		if !ok || s == "" {
			return reflect.MakeSlice(t, 0, 0).Interface(), nil
		}

		parts := strings.Split(s, sep)
		result := reflect.MakeSlice(t, 0, len(parts))
		for _, p := range parts {
			if tp := strings.TrimSpace(p); tp != "" {
				result = reflect.Append(result, reflect.ValueOf(tp).Convert(t.Elem()))
			}
		}
		return result.Interface(), nil
	}
}
