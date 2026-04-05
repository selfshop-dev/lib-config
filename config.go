package config

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/v2"
)

const (
	delim = "."
	tag   = "koanf"
)

// New loads, validates, and returns a fully-populated config struct of type T.
//
// Loading order (last writer wins):
//  1. Hardcoded defaults supplied by the caller.
//  2. Environment variables with the given prefix.
//
// envPrefix is normalised internally: trailing underscores are stripped and
// exactly one is appended, so "APP" and "APP_" are equivalent.
// envPrefix must not contain double underscores (__) — they are reserved
// as the hierarchy separator. Use a simple uppercase name: "APP", "SVC", "INIT".
//
// Returns a descriptive error for each class of misconfiguration; see package
// documentation for the five validation phases.
//
//nolint:funlen // five-phase validation pipeline — each phase is a distinct step, extraction would obscure the flow
func New[T any](envPrefix string, defaults map[string]any) (*T, error) {
	var zero T
	rt := derefType(reflect.TypeOf(zero))
	if rt.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type parameter T must be a struct, got %s", rt.Kind())
	}

	// Phase 1: struct contract — every exported field must have a koanf tag.
	// Detected before any I/O so the error is a fast, clear programming error.
	allowed, err := extractKnownKeys(rt, "")
	if err != nil {
		return nil, err
	}

	// Normalise prefix: always ensure it ends with exactly one "_".
	envPrefix = strings.TrimRight(envPrefix, "_") + "_"

	k := koanf.New(delim)

	// Load defaults first so env vars can override them selectively.
	// confmap.Provider never returns an error.
	_ = k.Load(confmap.Provider(defaults, delim), nil) //nolint:errcheck // confmap.Provider never returns an error

	// Load environment variables, stripping the prefix and normalising names.
	// env.Provider never returns an error.
	_ = k.Load(env.Provider(delim, env.Opt{ //nolint:errcheck // env.Provider never returns an error
		Prefix: envPrefix,
		TransformFunc: func(k, v string) (string, any) {
			k = strings.TrimPrefix(k, envPrefix)
			k = strings.ToLower(k)
			k = strings.TrimLeft(k, "_")
			k = strings.ReplaceAll(k, "__", delim)
			return k, v
		},
	}), nil)

	// Phase 2: unknown keys — collect all and report together via errors.Join.
	var unknownErrs []error
	for _, k := range k.Keys() {
		if !isKnown(k, allowed) {
			unknownErrs = append(unknownErrs, fmt.Errorf("unknown configuration key: %q", k))
		}
	}
	if len(unknownErrs) > 0 {
		return nil, errors.Join(unknownErrs...)
	}

	// Phase 3: decode — type mismatches surface here.
	var conf T
	if err := k.UnmarshalWithConf("", &conf, koanf.UnmarshalConf{
		Tag: tag,
		DecoderConfig: &mapstructure.DecoderConfig{
			WeaklyTypedInput:     false,
			ErrorUnused:          true,
			IgnoreUntaggedFields: false,
			Result:               &conf,
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToTimeHookFunc(time.RFC3339),
				stringToCleanSliceHookFunc(","),
			),
		},
	}); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	// Phase 4: struct tag validation — field-level constraints.
	// Each violation becomes a separate error in errors.Join so the operator
	// sees all problems at once, not just the first.
	val := validator.New(validator.WithRequiredStructEnabled())
	if err := val.Struct(&conf); err != nil {
		return nil, formatValidationError(err)
	}

	// Phase 5: semantic validation — cross-field constraints via Validate() on *T.
	if v, ok := any(&conf).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return nil, fmt.Errorf("semantic validation: %w", err)
		}
	}

	return &conf, nil
}

// formatValidationError converts go-playground/validator errors into
// human-readable messages joined with errors.Join so every violation
// is visible at once.
//
// Output format:
//
//	config: App.Runmode: must satisfy oneof=dev prod (got "staging")
//	config: Entry.HTTP.Port: must satisfy min=1024 (got 80)
//	config: Entry.HTTP.WriteTimeout: must satisfy gtfield=RequestTimeout (got 5s)
func formatValidationError(err error) error {
	var ves validator.ValidationErrors
	if !errors.As(err, &ves) {
		return fmt.Errorf("validation: %w", err)
	}

	errs := make([]error, 0, len(ves))
	for _, fe := range ves {
		// Namespace includes the root struct name (e.g. "SimpleConf.App.Runmode").
		// Strip the first segment to get the caller-facing path ("App.Runmode").
		ns := fe.Namespace()
		if _, rest, ok := strings.Cut(ns, "."); ok {
			ns = rest
		}

		// Include the param when present so the message is self-explanatory:
		//   "oneof=dev prod", "min=1024", "gtfield=RequestTimeout"
		// Without a param, just the tag name: "required".
		rule := fe.Tag()
		if p := fe.Param(); p != "" {
			rule = fe.Tag() + "=" + p
		}

		errs = append(errs, fmt.Errorf(
			"config: %s: must satisfy %s (got %v)",
			ns, rule, fe.Value(),
		))
	}
	return errors.Join(errs...)
}
