package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

// Source identifies where a configuration value came from.
type Source string

const (
	SourceDefault Source = "default"
	SourceFile    Source = "file"
	SourceEnv     Source = "env"
	SourceFlag    Source = "flag"
)

// FieldSource records the origin of each top-level field.
type FieldSource struct {
	Quality    Source
	Workers    Source
	OnError    Source
	OnConflict Source
}

// Loaded is the resolved configuration together with field provenance.
type Loaded struct {
	Config
	Sources FieldSource
}

// Load reads the config file at path and applies env overrides.
// Missing file is not an error — defaults are used.
func Load(path string) (*Loaded, error) {
	cfg := hardcodedDefaults
	src := FieldSource{
		Quality:    SourceDefault,
		Workers:    SourceDefault,
		OnError:    SourceDefault,
		OnConflict: SourceDefault,
	}

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			var fileCfg Config
			if _, err := toml.DecodeFile(path, &fileCfg); err != nil {
				return nil, fmt.Errorf("parse config %s: %w", path, err)
			}
			if fileCfg.Defaults.Quality != 0 {
				cfg.Defaults.Quality = fileCfg.Defaults.Quality
				src.Quality = SourceFile
			}
			if fileCfg.Defaults.Workers != 0 {
				cfg.Defaults.Workers = fileCfg.Defaults.Workers
				src.Workers = SourceFile
			}
			if fileCfg.Defaults.OnError != "" {
				cfg.Defaults.OnError = fileCfg.Defaults.OnError
				src.OnError = SourceFile
			}
			if fileCfg.Defaults.OnConflict != "" {
				cfg.Defaults.OnConflict = fileCfg.Defaults.OnConflict
				src.OnConflict = SourceFile
			}
			if fileCfg.Fonts.Mainfont != "" {
				cfg.Fonts.Mainfont = fileCfg.Fonts.Mainfont
			}
			if fileCfg.Fonts.Monofont != "" {
				cfg.Fonts.Monofont = fileCfg.Fonts.Monofont
			}
			if fileCfg.Fonts.Sansfont != "" {
				cfg.Fonts.Sansfont = fileCfg.Fonts.Sansfont
			}
			if fileCfg.Backend != nil {
				cfg.Backend = fileCfg.Backend
			}
			if fileCfg.Profile != nil {
				cfg.Profile = fileCfg.Profile
			}
		}
	}

	// Environment overrides: CONVERTR_QUALITY, CONVERTR_WORKERS, etc.
	if v := os.Getenv("CONVERTR_QUALITY"); v != "" {
		var q int
		if _, err := fmt.Sscanf(v, "%d", &q); err == nil {
			cfg.Defaults.Quality = q
			src.Quality = SourceEnv
		}
	}
	if v := os.Getenv("CONVERTR_WORKERS"); v != "" {
		var w int
		if _, err := fmt.Sscanf(v, "%d", &w); err == nil {
			cfg.Defaults.Workers = w
			src.Workers = SourceEnv
		}
	}
	if v := os.Getenv("CONVERTR_ON_ERROR"); v != "" {
		cfg.Defaults.OnError = strings.ToLower(v)
		src.OnError = SourceEnv
	}
	if v := os.Getenv("CONVERTR_ON_CONFLICT"); v != "" {
		cfg.Defaults.OnConflict = strings.ToLower(v)
		src.OnConflict = SourceEnv
	}

	return &Loaded{Config: cfg, Sources: src}, nil
}
