package config_test

import (
	"errors"
	"fmt"
	"os"

	config "github.com/selfshop-dev/lib-config"
)

type exampleConfig struct {
	Host string `koanf:"host" validate:"required"`
	Port int    `koanf:"port" validate:"required,min=1024"`
}

func ExampleNew() {
	cfg, err := config.New[exampleConfig]("APP", map[string]any{
		"host": "localhost",
		"port": 8080,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cfg.Host)
	fmt.Println(cfg.Port)

	// Output:
	// localhost
	// 8080
}

func ExampleNew_unknownKey() {
	_, err := config.New[exampleConfig]("APP", map[string]any{
		"host": "localhost",
		"port": 8080,
		"typo": "bad",
	})
	fmt.Println(err != nil)

	// Output:
	// true
}

type exampleWithValidate struct {
	Mode string `koanf:"mode" validate:"required"`
}

func (c *exampleWithValidate) Validate() error {
	if c.Mode != "dev" && c.Mode != "prod" {
		return errors.New("mode must be dev or prod")
	}
	return nil
}

func ExampleNew_semanticValidation() {
	_, err := config.New[exampleWithValidate]("APP", map[string]any{
		"mode": "staging",
	})
	fmt.Println(err != nil)

	// Output:
	// true
}

func ExampleNew_envOverride() {
	//nolint:gosec // os.Setenv used instead of t.Setenv — Example functions have no *testing.T
	os.Setenv("APP_HOST", "db.internal")
	defer os.Unsetenv("APP_HOST")

	cfg, err := config.New[exampleConfig]("APP", map[string]any{
		"host": "localhost",
		"port": 8080,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(cfg.Host)
	fmt.Println(cfg.Port)

	// Output:
	// db.internal
	// 8080
}

func ExampleNew_validationError() {
	_, err := config.New[exampleConfig]("APP", map[string]any{
		"host": "localhost",
		"port": 80, // below min=1024
	})
	fmt.Println(err != nil)

	// Output:
	// true
}

func ExampleBase_LogFormat() {
	b := config.Base{
		App: config.App{Name: "svc", Runmode: config.AppRunmodeProd},
		Log: config.Log{Format: config.LogFormatAuto, MinLevel: config.LogMinLevelInfo},
	}
	fmt.Println(b.LogFormat())

	// Output:
	// json
}
