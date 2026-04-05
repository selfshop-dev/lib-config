# lib-config

[![CI](https://github.com/selfshop-dev/lib-config/actions/workflows/ci.yml/badge.svg)](https://github.com/selfshop-dev/lib-config/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/selfshop-dev/lib-config/branch/main/graph/badge.svg)](https://codecov.io/gh/selfshop-dev/lib-config)
[![Go Report Card](https://goreportcard.com/badge/github.com/selfshop-dev/lib-config)](https://goreportcard.com/report/github.com/selfshop-dev/lib-config)
[![Go version](https://img.shields.io/github/go-mod/go-version/selfshop-dev/lib-config)](go.mod)
[![License](https://img.shields.io/github/license/selfshop-dev/lib-config)](LICENSE)

Строгий типобезопасный загрузчик конфигурации из переменных окружения для Go-сервисов. Проект организации [selfshop-dev](https://github.com/selfshop-dev).

### Installation

```bash
go get -u github.com/selfshop-dev/lib-config
```

## Overview

`lib-config` загружает конфигурацию из двух источников — хардкодных дефолтов и переменных окружения — и проверяет её за один вызов `New[T]`. Все ошибки выдаются сразу через `errors.Join`, а не по одной. Пакет намеренно ограничен: нет файловых источников, нет слабой типизации, нет молчаливого приёма некорректных значений.

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

### Быстрый старт

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

## Переменные окружения

Двойное подчёркивание (`__`) используется как разделитель вложенности, потому что одиночное подчёркивание слишком часто встречается внутри имён сегментов. Префикс нормализуется автоматически: `"APP"` и `"APP_"` эквивалентны.

```bash
APP_HOST=localhost         # → host
APP_DB__HOST=localhost     # → db.host
APP_ENTRY__HTTP__PORT=8080 # → entry.http.port
```

Переменная `APP_DB__HOST` читается как: убрать префикс `APP_`, перевести в нижний регистр, заменить `__` на `.`.

## Struct tag contract

Каждое экспортируемое поле обязано иметь явный тег `koanf`. Отсутствие тега — ошибка, обнаруживаемая до любого I/O.

```go
type Config struct {
    Name    string `koanf:"name"`    // скалярное поле
    Sub     Inner  `koanf:"sub"`     // вложенная структура
    Base           `koanf:",squash"` // flatten в родительское пространство имён
    Ignored string `koanf:"-"`       // исключено из загрузки
}
```

Поддерживается только опция `squash`. Остальные опции игнорируются.

## Фазы валидации

`New[T]` прогоняет пять фаз последовательно. Каждая фаза ловит свой класс ошибок, и все нарушения в пределах одной фазы выдаются вместе через `errors.Join`.

Первая фаза — **struct contract**: все экспортируемые поля должны иметь тег `koanf`. Обнаруживается до любого I/O через рефлексию.

Вторая фаза — **unknown keys**: все нераспознанные ключи репортируются сразу. Ловит опечатки в именах переменных окружения.

Третья фаза — **decode**: несоответствия типов (например, `"abc"` в `uint16`) репортируются через mapstructure.

Четвёртая фаза — **struct tags**: field-level ограничения через go-playground/validator. Ошибки форматируются как `"config: Field: must satisfy rule=param (got value)"`.

Пятая фаза — **semantic**: кросс-полевые ограничения через опциональный метод `Validate() error` на `*T`. Используй `errors.Join` внутри `Validate()`, чтобы репортировать все нарушения сразу.

```go
func (c *Config) Validate() error {
    return errors.Join(
        validatePair(c.MinConn, c.MaxConn, "min_conn must be less than max_conn"),
    )
}
```

## Base

`Base` содержит конфигурационные поля, общие для всех сервисов. Встраивай с `koanf:",squash"`, чтобы поля оказались на корневом уровне.

```go
type AppConfig struct {
    config.Base `koanf:",squash"`
    DB          DBConfig `koanf:"db" validate:"required"`
}
```

`Base` включает `App` (имя и runmode), `Log` (формат и уровень), `Entry.HTTP` (порт и таймауты) и флаг `Debug`.

```go
cfg.IsProd()    // runmode == "prod"
cfg.IsDev()     // runmode == "dev"
cfg.LogFormat() // разрешает "auto" в "json" или "console" по runmode
```

`Base.Validate()` автоматически вызывается через пятую фазу и запрещает `debug=true` в prod, а также требует `log.min_level=debug` при включённом debug-режиме.

### HTTP-таймауты

Порядок таймаутов зафиксирован в validator-тегах: `ReadTimeout < RequestTimeout < WriteTimeout` и `ReadTimeout < RequestTimeout < IdleTimeout`.

| Поле | Диапазон | Описание |
|---|---|---|
| `port` | 1024–65535 | TCP-порт сервера |
| `read_timeout` | 5s–60s | Максимальное время чтения запроса |
| `request_timeout` | 10s–120s | Дедлайн контекста обработчика |
| `write_timeout` | 5s–90s | Максимальное время записи ответа |
| `idle_timeout` | 30s–180s | Время жизни idle keep-alive соединения |

## Лицензия

[`MIT`](LICENSE) © 2026-present [`selfshop-dev`](https://github.com/selfshop-dev)