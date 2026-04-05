package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	config "github.com/selfshop-dev/lib-config"
)

func baseWith(runmode, logFormat, logLevel string, debug bool) config.Base {
	return config.Base{
		App: config.App{Name: "test-service", Runmode: runmode},
		Log: config.Log{Format: logFormat, MinLevel: logLevel},
		Entry: config.Entry{
			HTTP: config.HTTP{
				Port:           8080,
				ReadTimeout:    5 * time.Second,
				RequestTimeout: 10 * time.Second,
				WriteTimeout:   15 * time.Second,
				IdleTimeout:    30 * time.Second,
			},
		},
		Debug: debug,
	}
}

func TestBase_IsDev(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name    string
		runmode string
		want    bool
	}{
		{"dev runmode", config.AppRunmodeDev, true},
		{"prod runmode", config.AppRunmodeProd, false},
		{"empty runmode", "", false},
		{"unknown runmode", "staging", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b := baseWith(tc.runmode, config.LogFormatConsole, config.LogMinLevelInfo, false)
			assert.Equal(t, tc.want, b.IsDev())
		})
	}
}

func TestBase_IsProd(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name    string
		runmode string
		want    bool
	}{
		{"prod runmode", config.AppRunmodeProd, true},
		{"dev runmode", config.AppRunmodeDev, false},
		{"empty runmode", "", false},
		{"unknown runmode", "canary", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b := baseWith(tc.runmode, config.LogFormatConsole, config.LogMinLevelInfo, false)
			assert.Equal(t, tc.want, b.IsProd())
		})
	}
}

func TestBase_LogFormat(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name    string
		runmode string
		format  string
		want    string
	}{
		{"explicit json returned as-is", config.AppRunmodeDev, config.LogFormatJSON, config.LogFormatJSON},
		{"explicit console returned as-is", config.AppRunmodeProd, config.LogFormatConsole, config.LogFormatConsole},
		{"auto in prod resolves to json", config.AppRunmodeProd, config.LogFormatAuto, config.LogFormatJSON},
		{"auto in dev resolves to console", config.AppRunmodeDev, config.LogFormatAuto, config.LogFormatConsole},
		{"auto with unknown runmode resolves to console", "staging", config.LogFormatAuto, config.LogFormatConsole},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			b := baseWith(tc.runmode, tc.format, config.LogMinLevelInfo, false)
			assert.Equal(t, tc.want, b.LogFormat())
		})
	}
}

func TestBase_Validate(t *testing.T) {
	t.Parallel()

	testCases := [...]struct {
		name    string
		base    config.Base
		wantErr bool
	}{
		{
			name:    "valid dev without debug info level",
			base:    baseWith(config.AppRunmodeDev, config.LogFormatConsole, config.LogMinLevelInfo, false),
			wantErr: false,
		},
		{
			name:    "valid dev with debug any log level",
			base:    baseWith(config.AppRunmodeDev, config.LogFormatConsole, config.LogMinLevelDebug, true),
			wantErr: false,
		},
		{
			name:    "valid dev with debug info level — log level not enforced in dev",
			base:    baseWith(config.AppRunmodeDev, config.LogFormatConsole, config.LogMinLevelInfo, true),
			wantErr: false,
		},
		{
			name:    "valid prod without debug info level",
			base:    baseWith(config.AppRunmodeProd, config.LogFormatJSON, config.LogMinLevelInfo, false),
			wantErr: false,
		},
		{
			name:    "prod with debug mode is forbidden",
			base:    baseWith(config.AppRunmodeProd, config.LogFormatJSON, config.LogMinLevelInfo, true),
			wantErr: true,
		},
		{
			name:    "prod with debug log level is forbidden",
			base:    baseWith(config.AppRunmodeProd, config.LogFormatJSON, config.LogMinLevelDebug, false),
			wantErr: true,
		},
		{
			name:    "prod with debug mode and debug log level — both rules fire",
			base:    baseWith(config.AppRunmodeProd, config.LogFormatJSON, config.LogMinLevelDebug, true),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.base.Validate()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
