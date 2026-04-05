package config

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTagOptions(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name     string
		raw      string
		wantOpts tagOptions
		wantSkip bool
	}{
		{"plain name", "host", tagOptions{name: "host"}, false},
		{"dash means skip", "-", tagOptions{}, true},
		{"squash option only", ",squash", tagOptions{squash: true}, false},
		{"name with squash", "base,squash", tagOptions{name: "base", squash: true}, false},
		{"empty raw tag", "", tagOptions{name: ""}, false},
		{"unknown option ignored", "field,omitempty", tagOptions{name: "field"}, false},
		{"multiple options squash present", "field,omitempty,squash", tagOptions{name: "field", squash: true}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gotOpts, gotSkip := parseTagOptions(tc.raw)
			assert.Equal(t, tc.wantSkip, gotSkip)
			if !tc.wantSkip {
				assert.Equal(t, tc.wantOpts, gotOpts)
			}
		})
	}
}

type flatFields struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

type nestedFixture struct {
	Name string     `koanf:"name"`
	DB   flatFields `koanf:"db"`
}

type SquashBase struct {
	Debug bool `koanf:"debug"`
}

type withSquashFixture struct {
	SquashBase `koanf:",squash"`
	Name       string `koanf:"name"`
}

type withUnexportedField struct {
	Name    string `koanf:"name"`
	private string //nolint:unused // unexported field used only to test that LogFields skips it
}

type withDashFixture struct {
	Name    string `koanf:"name"`
	Ignored string `koanf:"-"`
}

type withSliceFixture struct {
	Tags []string     `koanf:"tags"`
	Svcs []flatFields `koanf:"svcs"`
}

type withSliceOfPtrFixture struct {
	Nodes []*flatFields `koanf:"nodes"`
}

type withMapFixture struct {
	Labels map[string]string `koanf:"labels"`
}

type withPointerFixture struct {
	Sub *flatFields `koanf:"sub"`
}

type missingTagDirect struct {
	Good string `koanf:"good"`
	Bad  string
}

type squashOnNonStructFixture struct {
	Value string `koanf:",squash"`
}

type BadInnerExported struct {
	Good string `koanf:"good"`
	Bad  string
}

type SquashWithExportedBadInner struct {
	BadInnerExported `koanf:",squash"`
}

type badElemStruct struct {
	OK  string `koanf:"ok"`
	Bad string
}

type sliceWithBadElem struct {
	Items []badElemStruct `koanf:"items"`
}

type nestedStructWithError struct {
	Sub missingTagDirect `koanf:"sub"`
}

type missingTagInner struct {
	Good string `koanf:"good"`
	Bad  string
}

type nestedWithMissingTag struct {
	Inner missingTagInner `koanf:"inner"`
}

func TestExtractKnownKeys(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name    string
		typ     reflect.Type
		prefix  string
		want    map[string]struct{}
		wantErr bool
	}{
		{
			name:   "flat struct no prefix",
			typ:    reflect.TypeFor[flatFields](),
			prefix: "",
			want:   map[string]struct{}{"host": {}, "port": {}},
		},
		{
			name:   "flat struct with prefix",
			typ:    reflect.TypeFor[flatFields](),
			prefix: "db",
			want:   map[string]struct{}{"db.host": {}, "db.port": {}},
		},
		{
			name:   "nested struct recurses",
			typ:    reflect.TypeFor[nestedFixture](),
			prefix: "",
			want:   map[string]struct{}{"name": {}, "db.host": {}, "db.port": {}},
		},
		{
			name:   "squash flattens exported embedded struct",
			typ:    reflect.TypeFor[withSquashFixture](),
			prefix: "",
			want:   map[string]struct{}{"debug": {}, "name": {}},
		},
		{
			name:   "unexported field skipped",
			typ:    reflect.TypeFor[withUnexportedField](),
			prefix: "",
			want:   map[string]struct{}{"name": {}},
		},
		{
			name:   "dash-tagged field excluded",
			typ:    reflect.TypeFor[withDashFixture](),
			prefix: "",
			want:   map[string]struct{}{"name": {}},
		},
		{
			name:   "slice of scalars and structs",
			typ:    reflect.TypeFor[withSliceFixture](),
			prefix: "",
			want:   map[string]struct{}{"tags": {}, "svcs": {}, "svcs.host": {}, "svcs.port": {}},
		},
		{
			name:   "slice of pointer-to-struct unwraps element",
			typ:    reflect.TypeFor[withSliceOfPtrFixture](),
			prefix: "",
			want:   map[string]struct{}{"nodes": {}, "nodes.host": {}, "nodes.port": {}},
		},
		{
			name:   "map field registers collection key only",
			typ:    reflect.TypeFor[withMapFixture](),
			prefix: "",
			want:   map[string]struct{}{"labels": {}},
		},
		{
			name:   "pointer-to-struct field unwrapped",
			typ:    reflect.TypeFor[withPointerFixture](),
			prefix: "",
			want:   map[string]struct{}{"sub.host": {}, "sub.port": {}},
		},
		{
			name:   "pointer-to-struct root type unwrapped in prologue",
			typ:    reflect.TypeFor[*flatFields](),
			prefix: "",
			want:   map[string]struct{}{"host": {}, "port": {}},
		},
		{
			name:   "non-struct type returns nil without error",
			typ:    reflect.TypeFor[string](),
			prefix: "",
			want:   map[string]struct{}{},
		},
		{
			name:    "missing koanf tag at top level returns error",
			typ:     reflect.TypeFor[missingTagDirect](),
			prefix:  "",
			wantErr: true,
		},
		{
			name:    "squash on non-struct returns error",
			typ:     reflect.TypeFor[squashOnNonStructFixture](),
			prefix:  "",
			wantErr: true,
		},
		{
			name:    "squash recursion propagates error from bad inner struct",
			typ:     reflect.TypeFor[SquashWithExportedBadInner](),
			prefix:  "",
			wantErr: true,
		},
		{
			name:    "nested struct with missing tag propagates error",
			typ:     reflect.TypeFor[nestedStructWithError](),
			prefix:  "",
			wantErr: true,
		},
		{
			name:    "slice with bad element struct propagates error",
			typ:     reflect.TypeFor[sliceWithBadElem](),
			prefix:  "",
			wantErr: true,
		},
		{
			name:    "nested missing tag through nested path",
			typ:     reflect.TypeFor[nestedWithMissingTag](),
			prefix:  "",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := extractKnownKeys(tc.typ, tc.prefix)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestIsKnown(t *testing.T) {
	t.Parallel()

	allowed := map[string]struct{}{
		"app.name":     {},
		"db.host":      {},
		"db.port":      {},
		"servers":      {},
		"servers.host": {},
	}

	testCases := [...]struct {
		name string
		key  string
		want bool
	}{
		{"exact scalar match", "app.name", true},
		{"exact nested match", "db.host", true},
		{"collection root exact match", "servers", true},
		{"collection sub-key via prefix", "servers.0.port", true},
		{"unknown key rejected", "app.typo", false},
		{"empty key rejected", "", false},
		{"partial name without dot not matched", "server", false},
		{"proper prefix of allowed key without dot", "db", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, isKnown(tc.key, allowed))
		})
	}
}
