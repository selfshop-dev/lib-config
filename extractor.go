package config

import (
	"fmt"
	"maps"
	"reflect"
	"strings"
)

// tagOptions holds the result of parsing a `koanf:"name,opt1,opt2"` tag value.
type tagOptions struct {
	name   string
	squash bool
}

// parseTagOptions parses the raw value of a koanf struct tag.
//
// Special values:
//   - "-"       → skip is true; field is excluded from config loading.
//   - ",squash" → flatten the embedded struct into the parent key namespace.
//
// Only the "squash" option is recognised; others are silently ignored.
func parseTagOptions(raw string) (opts tagOptions, skip bool) {
	if raw == "-" {
		return tagOptions{}, true
	}
	parts := strings.SplitN(raw, ",", 2)
	opts.name = parts[0]
	if len(parts) == 2 {
		for opt := range strings.SplitSeq(parts[1], ",") {
			if opt == "squash" {
				opts.squash = true
			}
		}
	}
	return opts, false
}

// extractKnownKeys walks the struct type t via reflection and returns the
// complete set of valid dot-delimited config keys.
//
// Rules:
//   - `koanf:"-"`       → field is skipped.
//   - `koanf:",squash"` → struct fields are registered under the current prefix.
//   - nested struct     → recurse; the struct key itself is not a leaf.
//   - slice/array/map   → collection key registered as prefix; struct elements recursed.
//   - scalar leaf       → full dotted key registered.
//
// Every exported field must carry an explicit koanf tag — a missing tag is a
// programming error returned immediately.
func extractKnownKeys(t reflect.Type, prefix string) (map[string]struct{}, error) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return map[string]struct{}{}, nil
	}

	allowed := make(map[string]struct{})
	for f := range t.Fields() {
		if err := processField(f, prefix, allowed); err != nil {
			return nil, err
		}
	}
	return allowed, nil
}

func processField(f reflect.StructField, prefix string, allowed map[string]struct{}) error {
	ft := derefType(f.Type)

	if f.PkgPath != "" {
		return nil
	}

	opts, skip := parseTagOptions(f.Tag.Get(tag))
	if skip {
		return nil
	}

	if opts.squash {
		return processSquash(f, ft, prefix, allowed)
	}

	if opts.name == "" {
		return fmt.Errorf(
			"%s has no koanf tag; add `koanf:\"<key>\"` or `koanf:\"-\"` to exclude it",
			f.Name,
		)
	}

	return processKeyed(ft, fullKey(prefix, opts.name), allowed)
}

func processSquash(f reflect.StructField, ft reflect.Type, prefix string, allowed map[string]struct{}) error {
	if ft.Kind() != reflect.Struct {
		return fmt.Errorf(
			"%s has koanf:\",squash\" but is not a struct (got %s)",
			f.Name, ft.Kind(),
		)
	}
	nested, err := extractKnownKeys(ft, prefix)
	if err != nil {
		return err
	}
	maps.Copy(allowed, nested)
	return nil
}

func processKeyed(ft reflect.Type, key string, allowed map[string]struct{}) error {
	switch ft.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return processCollection(ft, key, allowed)
	case reflect.Struct:
		return processStruct(ft, key, allowed)
	default:
		allowed[key] = struct{}{}
		return nil
	}
}

func processCollection(ft reflect.Type, key string, allowed map[string]struct{}) error {
	allowed[key] = struct{}{}
	e := derefType(ft.Elem())
	if e.Kind() != reflect.Struct {
		return nil
	}
	nested, err := extractKnownKeys(e, key)
	if err != nil {
		return err
	}
	maps.Copy(allowed, nested)
	return nil
}

func processStruct(ft reflect.Type, key string, allowed map[string]struct{}) error {
	nested, err := extractKnownKeys(ft, key)
	if err != nil {
		return err
	}
	maps.Copy(allowed, nested)
	return nil
}

// derefValue unwraps pointer types and values until a non-pointer is reached
// or a nil pointer is encountered. Returns the unwrapped type, value, and
// whether the pointer was nil.
func derefValue(t reflect.Type, v reflect.Value) (reflect.Type, reflect.Value, bool) {
	for t.Kind() == reflect.Pointer {
		if v.IsNil() {
			return t, v, true
		}
		t = t.Elem()
		v = v.Elem()
	}
	return t, v, false
}

// derefType unwraps pointer types until a non-pointer type is reached.
func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// fullKey joins prefix and name with the delimiter, or returns name alone
// when prefix is empty.
func fullKey(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + delim + name
}

// isKnown reports whether key is a valid config key: either an exact match
// or a sub-key of a known collection root (e.g. "servers.0.port" under "servers").
func isKnown(key string, allowed map[string]struct{}) bool {
	if _, ok := allowed[key]; ok {
		return true
	}
	for a := range allowed {
		if strings.HasPrefix(key, a+delim) {
			return true
		}
	}
	return false
}
