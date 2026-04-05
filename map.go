package config

import (
	"errors"
	"time"
)

// Runmode constants define the supported application run modes.
const (
	AppRunmodeDev  = "dev"
	AppRunmodeProd = "prod"
)

// Log format constants define the supported log output formats.
const (
	LogFormatJSON    = "json"
	LogFormatConsole = "console"
	// LogFormatAuto resolves to [LogFormatConsole] in dev and [LogFormatJSON]
	// in prod. Resolution happens via [Base.LogFormat]; never pass "auto" to
	// the logger directly.
	LogFormatAuto = "auto"
)

// Log level constants define the supported minimum log severity levels.
const (
	LogMinLevelDebug = "debug"
	LogMinLevelInfo  = "info"
	LogMinLevelWarn  = "warn"
	LogMinLevelError = "error"
	LogMinLevelPanic = "panic"
	LogMinLevelFatal = "fatal"
)

// MinUnprivilegedPort is the lowest TCP/UDP port bindable without root on Linux.
const MinUnprivilegedPort = 1024

// Base holds configuration fields shared by every service.
// Embed with koanf:",squash" so fields appear at root level:
//
//	type AppConfig struct {
//	    config.Base `koanf:",squash"`
//	    DB          DBConfig `koanf:"db" validate:"required"`
//	}
type Base struct {
	App   App   `koanf:"app"   validate:"required"`
	Log   Log   `koanf:"log"   validate:"required"`
	Entry Entry `koanf:"entry" validate:"required"`
	// Debug enables verbose developer tooling. When true:
	//   - runmode must be "dev" (forbidden in prod by [Base.Validate])
	//   - log.min_level must be "debug"
	Debug bool `koanf:"debug"`
}

// IsDev reports whether the application is running in development mode.
func (b Base) IsDev() bool { return b.App.Runmode == AppRunmodeDev }

// IsProd reports whether the application is running in production mode.
func (b Base) IsProd() bool { return b.App.Runmode == AppRunmodeProd }

// LogFormat resolves the effective log format, expanding [LogFormatAuto] to a
// concrete value based on runmode. Use this when constructing the logger —
// never read Log.Format directly.
func (b Base) LogFormat() string {
	if b.Log.Format != LogFormatAuto {
		return b.Log.Format
	}
	if b.IsProd() {
		return LogFormatJSON
	}
	return LogFormatConsole
}

// Validate enforces cross-field constraints called automatically by [New].
//
// Rules:
//   - Debug mode is forbidden in prod: it exposes internals and degrades performance.
//   - log.min_level must not be "debug" in prod: it leaks internal details and
//     generates excessive noise under production load.
func (b Base) Validate() error {
	var errs []error

	if b.IsProd() && b.Debug {
		errs = append(errs, errors.New("debug mode must be disabled in prod runmode"))
	}
	if b.IsProd() && b.Log.MinLevel == LogMinLevelDebug {
		errs = append(errs, errors.New("log.min_level must not be 'debug' in prod runmode"))
	}

	return errors.Join(errs...)
}

// App holds identity and runtime-mode configuration for the service.
type App struct {
	// Name is the human-readable service identifier used in logs and traces.
	Name string `koanf:"name" validate:"required"`
	// Runmode controls environment-specific behaviour. Must be one of
	// [AppRunmodeDev] or [AppRunmodeProd].
	Runmode string `koanf:"runmode" validate:"required,oneof=dev prod"`
}

// Log holds logging configuration.
type Log struct {
	// MinLevel is the minimum severity level at which log entries are emitted.
	MinLevel string `koanf:"min_level" validate:"required,oneof=debug info warn error panic fatal"`
	// Format controls the log output encoding. Use [LogFormatAuto] to let
	// runmode decide; read the resolved value via [Base.LogFormat].
	Format string `koanf:"format" validate:"required,oneof=json auto console"`
}

// Entry holds inbound traffic entrypoint configuration.
type Entry struct {
	HTTP HTTP `koanf:"http" validate:"required"`
}

// HTTP holds configuration for the HTTP server entrypoint.
//
// Timeout ordering enforced by validator tags:
//
//	ReadTimeout < RequestTimeout < WriteTimeout
//	ReadTimeout < RequestTimeout < IdleTimeout
type HTTP struct {
	// Port is the TCP port the server listens on (unprivileged range).
	Port uint16 `koanf:"port" validate:"required,min=1024,max=65535"`
	// ReadTimeout is the maximum time to read a complete request including body.
	ReadTimeout time.Duration `koanf:"read_timeout" validate:"required,gte=5s,lte=60s"`
	// RequestTimeout is the context deadline injected into each request handler.
	// Must exceed ReadTimeout so the handler has a meaningful budget.
	RequestTimeout time.Duration `koanf:"request_timeout" validate:"required,gte=10s,lte=120s,gtfield=ReadTimeout"`
	// WriteTimeout is the maximum time to write a complete response.
	// Must exceed RequestTimeout so the server can flush after the handler completes.
	WriteTimeout time.Duration `koanf:"write_timeout" validate:"required,gte=5s,lte=90s,gtfield=RequestTimeout"`
	// IdleTimeout is the maximum time to keep an idle keep-alive connection open.
	// Must exceed RequestTimeout so a connection is not closed mid-request.
	IdleTimeout time.Duration `koanf:"idle_timeout" validate:"required,gte=30s,lte=180s,gtfield=RequestTimeout"`
}
