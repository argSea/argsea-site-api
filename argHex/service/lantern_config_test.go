package service_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/argSea/argsea-site-api/argHex/service"
	"github.com/spf13/viper"
)

// loadLanternConfig mirrors main.go's wiring exactly: same keys, same viper
// getters, over a real config file, so this test exercises the actual
// config-load path the server uses. (It can't live in package main: main's
// init() exits without --config/--log flags, which a test binary never has.)
func loadLanternConfig(t *testing.T, configJSON string) service.LanternConfig {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.json")

	if err := os.WriteFile(path, []byte(configJSON), 0644); nil != err {
		t.Fatalf("could not write config: %v", err)
	}

	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); nil != err {
		t.Fatalf("could not load config: %v", err)
	}

	return service.LanternConfig{
		SiteDir:  v.GetString("lantern.site_dir"),
		BuildCmd: v.GetStringSlice("lantern.build_cmd"),
		DistDir:  v.GetString("lantern.dist_dir"),
		Keep:     v.GetInt("lantern.keep"),
		Timeout:  time.Duration(v.GetInt("lantern.timeout_seconds")) * time.Second,
		Env:      v.GetStringSlice("lantern.env"),
	}
}

func TestLanternEnvKeyCaseSurvivesConfigLoad(t *testing.T) {
	cfg := loadLanternConfig(t, `{
		"lantern": {
			"site_dir":        "/srv/argsea/site",
			"build_cmd":       ["npm", "run", "build"],
			"dist_dir":        "dist",
			"keep":            2,
			"timeout_seconds": 600,
			"env":             ["ARGSEA_API_URL=http://127.0.0.1:8181"]
		}
	}`)

	if 1 != len(cfg.Env) || "ARGSEA_API_URL=http://127.0.0.1:8181" != cfg.Env[0] {
		t.Fatalf("the env entry must survive the config load byte-for-byte, got %+v", cfg.Env)
	}

	if 3 != len(cfg.BuildCmd) || "npm" != cfg.BuildCmd[0] {
		t.Fatalf("build_cmd must load as an argv array, got %+v", cfg.BuildCmd)
	}
}

// This pins the viper behavior that forced env to be an array of KEY=VALUE
// strings: nested JSON object keys are insensitivised (lowercased) on load, so
// a map-shaped env would silently corrupt ARGSEA_API_URL. If this test ever
// fails, the array workaround can be revisited.
func TestViperLowercasesMapKeysWhichIsWhyEnvIsAnArray(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := os.WriteFile(path, []byte(`{"lantern":{"env":{"ARGSEA_API_URL":"x"}}}`), 0644); nil != err {
		t.Fatalf("could not write config: %v", err)
	}

	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); nil != err {
		t.Fatalf("could not load config: %v", err)
	}

	env := v.GetStringMapString("lantern.env")

	if _, lowercased := env["argsea_api_url"]; !lowercased {
		t.Fatalf("expected viper to lowercase map keys (the reason env is an array); got %+v", env)
	}
}
