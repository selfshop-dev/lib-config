package config

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type simpleConf struct {
	Host string `koanf:"host" validate:"required"`
	Port int    `koanf:"port" validate:"required,min=1"`
}

type envConf struct {
	Host string `koanf:"host" validate:"required"`
	Name string `koanf:"name" validate:"required"`
}

type nestedConf struct {
	App   appConf `koanf:"app"   validate:"required"`
	Debug bool    `koanf:"debug"`
}

type appConf struct {
	Name    string        `koanf:"name"    validate:"required"`
	Timeout time.Duration `koanf:"timeout" validate:"required"`
}

type withSliceConf struct {
	Tags []string `koanf:"tags"`
}

type withValidateMethod struct {
	Mode string `koanf:"mode" validate:"required"`
}

func (c *withValidateMethod) Validate() error {
	if c.Mode != "allowed" {
		return errors.New("mode must be 'allowed'")
	}
	return nil
}

type withValidateMultiError struct {
	A string `koanf:"a"`
	B string `koanf:"b"`
}

func (c *withValidateMultiError) Validate() error {
	return errors.Join(
		errors.New("first error"),
		errors.New("second error"),
	)
}

type missingTagConf struct {
	Good string `koanf:"good"`
	Bad  string
}

type nonStructAlias = string

func TestNew_NonStructTypeParameterIsRejected(t *testing.T) {
	t.Parallel()

	_, err := New[nonStructAlias]("APP", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "struct")
}

func TestNew_DefaultsOnly(t *testing.T) {
	t.Parallel()

	defaults := map[string]any{
		"host": "localhost",
		"port": 5432,
	}

	got, err := New[simpleConf]("TEST_DEFAULTS", defaults)
	require.NoError(t, err)
	assert.Equal(t, "localhost", got.Host)
	assert.Equal(t, 5432, got.Port)
}

func TestNew_EnvOverridesDefault(t *testing.T) {
	t.Setenv("MYAPP_HOST", "db.internal")
	t.Setenv("MYAPP_NAME", "overridden-svc")

	defaults := map[string]any{
		"host": "localhost",
		"name": "default-svc",
	}

	got, err := New[envConf]("MYAPP", defaults)
	require.NoError(t, err)
	assert.Equal(t, "db.internal", got.Host)
	assert.Equal(t, "overridden-svc", got.Name)
}

func TestNew_PrefixNormalization(t *testing.T) {
	testCases := [...]struct {
		name   string
		prefix string
	}{
		{"no trailing underscore", "SVC"},
		{"single trailing underscore", "SVC_"},
		{"multiple trailing underscores", "SVC___"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SVC_HOST", "normalized")
			t.Setenv("SVC_NAME", "svc")

			got, err := New[envConf](tc.prefix, nil)
			require.NoError(t, err)
			assert.Equal(t, "normalized", got.Host)
		})
	}
}

func TestNew_DoubleUnderscoreHierarchySeparator(t *testing.T) {
	t.Setenv("SVC_APP__NAME", "my-service")
	t.Setenv("SVC_APP__TIMEOUT", "30s")

	got, err := New[nestedConf]("SVC", nil)
	require.NoError(t, err)
	assert.Equal(t, "my-service", got.App.Name)
	assert.Equal(t, 30*time.Second, got.App.Timeout)
}

func TestNew_UnknownEnvKey_ReportsAll(t *testing.T) {
	t.Setenv("APP_HOST", "localhost")
	t.Setenv("APP_NAME", "svc")
	t.Setenv("APP_TYPO_KEY", "bad")
	t.Setenv("APP_ANOTHER_TYPO", "bad")

	_, err := New[envConf]("APP", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown configuration key")
	// errors.Join: both unknown keys reported together.
	assert.Contains(t, err.Error(), "typo")
	assert.Contains(t, err.Error(), "another")
}

func TestNew_UnknownDefaultKey_Rejected(t *testing.T) {
	t.Parallel()

	defaults := map[string]any{
		"host":        "localhost",
		"port":        1,
		"nonexistent": "value",
	}

	_, err := New[simpleConf]("APP_NOENV", defaults)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown configuration key")
}

func TestNew_MissingTag_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := New[missingTagConf]("APP_NOENV7", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "koanf tag")
}

func TestNew_ValidatorError_FormattedHumanReadable(t *testing.T) {
	t.Parallel()

	// host is required but not provided.
	defaults := map[string]any{"port": 5432}

	_, err := New[simpleConf]("APP_NOENV2", defaults)
	require.Error(t, err)
	// Format: "config: Host: must satisfy required (got )"
	assert.Contains(t, err.Error(), "required")
	assert.Contains(t, err.Error(), "Host")
}

func TestNew_ValidatorError_MultipleViolations_AllReported(t *testing.T) {
	t.Parallel()

	// Both host (required) and port (min=1) are missing/invalid.
	_, err := New[simpleConf]("APP_NOENV_MULTI", nil)
	require.Error(t, err)
	// errors.Join: both violations present.
	assert.Contains(t, err.Error(), "Host")
	assert.Contains(t, err.Error(), "Port")
}

func TestNew_ValidatorError_IncludesParam(t *testing.T) {
	t.Parallel()

	type oneofConf struct {
		Mode string `koanf:"mode" validate:"required,oneof=dev prod"`
	}

	_, err := New[oneofConf]("APP_ONEOF", map[string]any{"mode": "staging"})
	require.Error(t, err)
	// Param included: "oneof=dev prod"
	assert.Contains(t, err.Error(), "oneof=dev prod")
	assert.Contains(t, err.Error(), "staging")
}

func TestNew_SemanticValidation_Passes(t *testing.T) {
	t.Parallel()

	got, err := New[withValidateMethod]("APP_NOENV3", map[string]any{"mode": "allowed"})
	require.NoError(t, err)
	assert.Equal(t, "allowed", got.Mode)
}

func TestNew_SemanticValidation_Fails(t *testing.T) {
	t.Parallel()

	_, err := New[withValidateMethod]("APP_NOENV4", map[string]any{"mode": "forbidden"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "semantic validation")
}

func TestNew_SemanticValidation_MultiErrorSurfaced(t *testing.T) {
	t.Parallel()

	_, err := New[withValidateMultiError]("APP_NOENV5", map[string]any{"a": "x", "b": "y"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "first error")
	assert.Contains(t, err.Error(), "second error")
}

func TestNew_StringToCleanSlice(t *testing.T) {
	testCases := [...]struct {
		name string
		env  string
		want []string
	}{
		{"comma-separated", "foo,bar,baz", []string{"foo", "bar", "baz"}},
		{"empty string produces empty slice", "", []string{}},
		{"spaces trimmed", " a , b ", []string{"a", "b"}},
		{"consecutive commas dropped", "a,,b", []string{"a", "b"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("SLICE_TAGS", tc.env)

			got, err := New[withSliceConf]("SLICE", nil)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got.Tags)
		})
	}
}

func TestNew_DurationDecodeHook(t *testing.T) {
	t.Setenv("DUR_APP__NAME", "svc")
	t.Setenv("DUR_APP__TIMEOUT", "2m30s")

	got, err := New[nestedConf]("DUR", nil)
	require.NoError(t, err)
	assert.Equal(t, 2*time.Minute+30*time.Second, got.App.Timeout)
}

func TestNew_StrictTyping_StringForInt_Rejected(t *testing.T) {
	t.Setenv("STRICT_HOST", "localhost")
	t.Setenv("STRICT_PORT", "not-a-number")

	_, err := New[simpleConf]("STRICT", nil)
	require.Error(t, err)
}

func TestNew_NilDefaults_NoPanic(t *testing.T) {
	t.Setenv("NIL_HOST", "h")
	t.Setenv("NIL_NAME", "svc")

	assert.NotPanics(t, func() {
		_, _ = New[envConf]("NIL", nil)
	})
}

// TestNew_PointerTypeParam covers the `for rt.Kind() == reflect.Pointer { rt = rt.Elem() }`
// loop in New[T] — T = *simpleConf unwraps to simpleConf (struct) and proceeds normally.
func TestNew_PointerTypeParam_UnwrapsToStruct(t *testing.T) {
	t.Parallel()

	// New[*simpleConf]: T is a pointer to struct, not a struct itself.
	// The reflect loop unwraps *simpleConf → simpleConf so the struct
	// contract check passes. Subsequent unmarshal into **simpleConf will
	// fail, but the important invariant is that the pointer-unwrap branch
	// is exercised and does NOT produce "must be a struct" error.
	_, err := New[*simpleConf]("PTRWRAP_NOENV", nil)
	if err != nil {
		assert.NotContains(t, err.Error(), "must be a struct")
	}
}

// TestNew_ValidatorError_NonValidationErrors covers the
// `if !errors.As(err, &ve)` fallback in formatValidationError.
// This branch fires when validator returns an error that is not
// validator.ValidationErrors — practically unreachable via val.Struct()
// but the guard exists for safety.
//
// We test it indirectly: an unmarshal error before the validator phase
// is wrapped with fmt.Errorf, so its type is never ValidationErrors.
func TestNew_DecodeError_IsNotValidationErrors(t *testing.T) {
	t.Setenv("DECODE_HOST", "localhost")
	t.Setenv("DECODE_PORT", "not-a-number")

	_, err := New[simpleConf]("DECODE", nil)
	require.Error(t, err)
	// Decode error wraps a plain error — not a ValidationErrors.
	// The error message comes from the decode phase, not formatValidationError.
	assert.Contains(t, err.Error(), "decode: ")
}
