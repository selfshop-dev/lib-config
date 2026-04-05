// Package config provides a strict, type-safe configuration loader that
// sources values from hardcoded defaults and environment variables.
//
// # Design philosophy
//
// This package is an opinionated infra-layer tool, not a general-purpose
// configuration library. Every decision prioritises catching misconfiguration
// at startup rather than silently accepting a degraded runtime state.
//
// Key properties:
//   - Two sources only: defaults (map[string]any) + env vars. No file sources.
//     In containerised / twelve-factor deployments env vars are the standard
//     runtime knob; files add operational complexity (mounts, permissions, drift).
//   - Strict typing: WeaklyTypedInput is false. Accepting "1" where int is
//     expected hides bugs; explicit conversion must happen at the source.
//   - Unknown keys are reported all at once via errors.Join before decoding.
//     A typo in an env var name is caught immediately with a clear error
//     instead of being silently ignored.
//   - Every exported struct field must carry an explicit koanf tag.
//     Relying on lowercase fallback names is fragile and hides intent.
//   - Validator errors are formatted as human-readable field + rule + value
//     messages — not raw go-playground error strings.
//   - Cross-field validation is supported via an optional Validate() error
//     method on *T (pointer receiver).
//
// # Environment variable naming
//
// Double-underscore (__) is used as a hierarchy separator because a single
// underscore is too common inside segment names (e.g. DB_MAX_CONNS would be
// ambiguous). Examples with prefix "APP":
//
//	APP__DEBUG=true             → debug
//	APP__DB__HOST=localhost     → db.host
//	APP__ENTRY__HTTP__PORT=8080 → entry.http.port
//
// The prefix is normalised internally: "APP" and "APP_" are equivalent.
// The prefix itself must not contain double underscores (__) — they are
// reserved as the hierarchy separator. Use a simple uppercase name without
// double underscores: "APP", "SVC", "INIT".
//
// # Struct tag contract
//
// All exported fields must define a koanf tag explicitly:
//
//	Name    string `koanf:"name"`    // scalar leaf
//	Sub     Inner  `koanf:"sub"`     // nested struct
//	Base           `koanf:",squash"` // flatten into parent namespace
//	ignored string `koanf:"-"`       // excluded from config loading
//
// Only the "squash" option is supported.
//
// # Validation
//
// Validation runs in five phases, each surfacing a different class of error:
//
//  1. Struct contract: every exported field must have an explicit koanf tag.
//     Detected at New[T] call time via reflection — programming error.
//  2. Unknown keys: all unrecognised keys reported together via errors.Join.
//     Catches typos in env var names and stale defaults.
//  3. Decode: type mismatches (e.g. "abc" into uint16) reported by mapstructure.
//  4. Struct tags: field-level constraints via go-playground/validator.
//     Formatted as human-readable "field: rule (got value)" messages.
//  5. Semantic: cross-field constraints via Validate() error on *T.
//     Use errors.Join inside Validate() to surface all violations at once.
//
// # Startup logging
//
// [LogFields] returns a flat slice of key-value pairs safe for structured
// logging. Keys containing sensitive substrings (password, secret, dsn, token,
// key, credential, auth) are replaced with "[redacted]":
//
//	logger.Info("config loaded", config.LogFields(cfg)...)
package config
