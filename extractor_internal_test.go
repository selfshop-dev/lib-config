package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFullKey(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "name", fullKey("", "name"))
	assert.Equal(t, "db.host", fullKey("db", "host"))
	assert.Equal(t, "a.b.c", fullKey("a.b", "c"))
}

func TestDerefValue(t *testing.T) {
	t.Parallel()

	t.Run("non-pointer passthrough", func(t *testing.T) {
		t.Parallel()
		s := "hello"
		rt, rv, ok := derefValue(reflect.TypeFor[string](), reflect.ValueOf(s))
		assert.Equal(t, reflect.TypeFor[string](), rt)
		assert.Equal(t, s, rv.Interface())
		assert.False(t, ok)
	})

	t.Run("non-nil pointer unwrapped", func(t *testing.T) {
		t.Parallel()
		s := "hello"
		rt, rv, ok := derefValue(reflect.TypeFor[*string](), reflect.ValueOf(&s))
		assert.Equal(t, reflect.TypeFor[string](), rt)
		assert.Equal(t, s, rv.Interface())
		assert.False(t, ok)
	})

	t.Run("nil pointer returns ok=true", func(t *testing.T) {
		t.Parallel()
		var p *string
		_, _, ok := derefValue(reflect.TypeFor[*string](), reflect.ValueOf(p))
		assert.True(t, ok)
	})
}

func TestDerefType(t *testing.T) {
	t.Parallel()

	assert.Equal(t, reflect.TypeFor[string](), derefType(reflect.TypeFor[string]()))
	assert.Equal(t, reflect.TypeFor[string](), derefType(reflect.TypeFor[*string]()))
	assert.Equal(t, reflect.TypeFor[string](), derefType(reflect.TypeFor[**string]()))
}
