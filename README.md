# lib-config

[![CI](https://github.com/selfshop-dev/lib-config/actions/workflows/ci.yml/badge.svg)](https://github.com/selfshop-dev/lib-config/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/selfshop-dev/lib-config/branch/main/graph/badge.svg)](https://codecov.io/gh/selfshop-dev/lib-config)
[![Go Report Card](https://goreportcard.com/badge/github.com/selfshop-dev/lib-config)](https://goreportcard.com/report/github.com/selfshop-dev/lib-config)
[![Go version](https://img.shields.io/github/go-mod/go-version/selfshop-dev/lib-config)](go.mod)
[![License](https://img.shields.io/github/license/selfshop-dev/lib-config)](LICENSE)

Strict, type-safe configuration loader from environment variables for Go services. A project by [selfshop-dev](https://github.com/selfshop-dev).

### Installation

```bash
go get -u github.com/selfshop-dev/lib-config
```

## Overview

`lib-config` loads configuration from two sources — hardcoded defaults and environment variables — and validates it in a single `New[T]` call. All errors are reported at once via `errors.Join`, not one at a time. The package is intentionally narrow: no file sources, no weak typing, no silent acceptance of invalid values.

```go
type AppConfig struct {
    config.Base `koanf:",squash"`
    DB          DBConfig `koanf:"db" validate:"required"`
}

cfg, err := config.New[AppConfig]("APP", map[string]any{
    "app.name":    "my-service",
    "app.runmode": "prod",
})
if err != nil {
    log.Fatal(err)
}
```

### Quick Start

```go
import config "github.com/selfshop-dev/lib-config"

type Config struct {
    Host string `koanf:"host" validate:"required"`
    Port int    `koanf:"port" validate:"required,min=1024"`
}

cfg, err := config.New[Config]("SVC", map[string]any{
    "host": "localhost",
    "port": 8080,
})
```

## Environment Variables

Double underscore (`__`) is used as the nesting separator because a single underscore appears too frequently within segment names. The prefix is normalized automatically: `"APP"` and `"APP_"` are equivalent.

```bash
APP_HOST=localhost         # → host
APP_DB__HOST=localhost     # → db.host
APP_ENTRY__HTTP__PORT=8080 # → entry.http.port
```

`APP_DB__HOST` is read as: strip the `APP_` prefix, lowercase, replace `__` with `.`.

## Struct Tag Contract

Every exported field must have an explicit `koanf` tag. A missing tag is an error detected before any I/O.

```go
type Config struct {
    Name    string `koanf:"name"`    // scalar field
    Sub     Inner  `koanf:"sub"`     // nested struct
    Base           `koanf:",squash"` // flattened into the parent namespace
    Ignored string `koanf:"-"`       // excluded from loading
}
```

Only the `squash` option is supported. All other options are ignored.

## Validation Phases

`New[T]` runs five phases in sequence. Each phase catches its own class of errors, and all violations within a phase are reported together via `errors.Join`.

**Phase 1 — struct contract**: all exported fields must have a `koanf` tag. Detected before any I/O via reflection.

**Phase 2 — unknown keys**: all unrecognized keys are reported at once. Catches typos in environment variable names.

**Phase 3 — decode**: type mismatches (e.g. `"abc"` into `uint16`) are reported via mapstructure.

**Phase 4 — struct tags**: field-level constraints via go-playground/validator. Errors are formatted as `"config: Field: must satisfy rule=param (got value)"`.

**Phase 5 — semantic**: cross-field constraints via an optional `Validate() error` method on `*T`. Use `errors.Join` inside `Validate()` to report all violations at once.

```go
func (c *Config) Validate() error {
    return errors.Join(
        validatePair(c.MinConn, c.MaxConn, "min_conn must be less than max_conn"),
    )
}
```

## Base

`Base` contains configuration fields common to all services. Embed with `koanf:",squash"` to place the fields at the root level.

```go
type AppConfig struct {
    config.Base `koanf:",squash"`
    DB          DBConfig `koanf:"db" validate:"required"`
}
```

`Base` includes `App` (name and runmode), `Log` (format and level), `Entry.HTTP` (port and timeouts), and a `Debug` flag.

```go
cfg.IsProd()    // runmode == "prod"
cfg.IsDev()     // runmode == "dev"
cfg.LogFormat() // resolves "auto" to "json" or "console" based on runmode
```

`Base.Validate()` is invoked automatically via phase 5 and disallows `debug=true` in prod, and requires `log.min_level=debug` when debug mode is enabled.

### HTTP Timeouts

Timeout ordering is enforced via validator tags: `ReadTimeout < RequestTimeout < WriteTimeout` and `ReadTimeout < RequestTimeout < IdleTimeout`.

| Field | Range | Description |
|---|---|---|
| `port` | 1024–65535 | Server TCP port |
| `read_timeout` | 5s–60s | Maximum time to read the request |
| `request_timeout` | 10s–120s | Handler context deadline |
| `write_timeout` | 5s–90s | Maximum time to write the response |
| `idle_timeout` | 30s–180s | Keep-alive idle connection lifetime |

## License

[`MIT`](LICENSE) © 2026-present [`selfshop-dev`](https://github.com/selfshop-dev)