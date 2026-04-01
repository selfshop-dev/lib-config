package example_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	example "github.com/selfshop-dev/tmpl-seed"
)

func TestCalc_Sum(t *testing.T) {
	t.Parallel()

	c := example.Calc{}

	testCases := [...]struct {
		name string
		a    int
		b    int
		want int
	}{
		{name: "positive numbers", a: 2, b: 3, want: 5},
		{name: "zero values", a: 0, b: 0, want: 0},
		{name: "negative numbers", a: -2, b: -3, want: -5},
		{name: "mixed signs", a: -1, b: 1, want: 0},
		{name: "large values", a: 1_000_000, b: 2_000_000, want: 3_000_000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, c.Sum(tc.a, tc.b))
		})
	}
}

func TestCalc_Diff(t *testing.T) {
	t.Parallel()

	c := example.Calc{}

	testCases := [...]struct {
		name string
		a    int
		b    int
		want int
	}{
		{name: "positive numbers", a: 5, b: 3, want: 2},
		{name: "zero values", a: 0, b: 0, want: 0},
		{name: "negative numbers", a: -2, b: -3, want: 1},
		{name: "mixed signs", a: -1, b: 1, want: -2},
		{name: "large values", a: 3_000_000, b: 1_000_000, want: 2_000_000},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, c.Diff(tc.a, tc.b))
		})
	}
}
