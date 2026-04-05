package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Tags []string

func TestStringToCleanSliceHookFunc(t *testing.T) {
	t.Parallel()

	hook := stringToCleanSliceHookFunc(",")

	tString := reflect.TypeFor[string]()
	tSliceString := reflect.TypeFor[[]string]()
	tSliceInt := reflect.TypeFor[[]int]()
	tTags := reflect.TypeFor[Tags]()

	testCases := [...]struct {
		name  string
		from  reflect.Type
		to    reflect.Type
		input any
		want  any
	}{
		{"single element", tString, tSliceString, "foo", []string{"foo"}},
		{"multiple elements", tString, tSliceString, "a,b,c", []string{"a", "b", "c"}},
		{"whitespace trimmed", tString, tSliceString, " foo , bar ", []string{"foo", "bar"}},
		{"empty string produces empty slice", tString, tSliceString, "", []string{}},
		{"consecutive delimiters dropped", tString, tSliceString, "a,,b", []string{"a", "b"}},
		{"leading delimiter dropped", tString, tSliceString, ",a,b", []string{"a", "b"}},
		{"trailing delimiter dropped", tString, tSliceString, "a,b,", []string{"a", "b"}},
		{"whitespace-only element dropped", tString, tSliceString, "a,   ,b", []string{"a", "b"}},
		{"all-whitespace produces empty slice", tString, tSliceString, "   ", []string{}},
		{"named Tags type supported", tString, tTags, "x,y", Tags{"x", "y"}},
		{"named Tags empty input", tString, tTags, "", Tags{}},
		{"non-string source passed through", reflect.TypeFor[int](), tSliceString, 42, 42},
		{"non-string-slice target passed through", tString, tSliceInt, "1,2,3", "1,2,3"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fn, ok := hook.(func(reflect.Type, reflect.Type, any) (any, error))
			require.True(t, ok, "hook must implement the expected function signature")

			got, err := fn(tc.from, tc.to, tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
