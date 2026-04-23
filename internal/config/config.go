// Package config loads and merges convertr configuration.
package config

import "github.com/Mark1708/convertr/internal/fonts"

// Config is the top-level configuration structure.
type Config struct {
	Defaults DefaultsConfig           `toml:"defaults"`
	Fonts    fonts.Config             `toml:"fonts"`
	Backend  map[string]BackendConfig `toml:"backend"`
	Profile  map[string]ProfileConfig `toml:"profile"`
}

// DefaultsConfig holds global defaults applied to every conversion.
type DefaultsConfig struct {
	Quality    int    `toml:"quality"`     // 0 = unset
	Workers    int    `toml:"workers"`     // 0 = GOMAXPROCS
	OnError    string `toml:"on_error"`    // skip|stop|retry
	OnConflict string `toml:"on_conflict"` // overwrite|skip|rename|error
}

// BackendConfig holds per-backend overrides (extra_args, etc.).
type BackendConfig struct {
	ExtraArgs []string `toml:"extra_args"`
}

// ProfileConfig holds named profile overrides that merge over Defaults.
type ProfileConfig struct {
	Quality    int    `toml:"quality"`
	Workers    int    `toml:"workers"`
	OnError    string `toml:"on_error"`
	OnConflict string `toml:"on_conflict"`
}

// hardcoded defaults — used when no config file is present.
var hardcodedDefaults = Config{
	Defaults: DefaultsConfig{
		Quality:    85,
		Workers:    0,
		OnError:    "skip",
		OnConflict: "overwrite",
	},
}

// Default returns the hardcoded default configuration.
func Default() Config { return hardcodedDefaults }
