package xdg

import (
	"os"
	"path/filepath"
)

// ConfigPath returns the path to convertr's config file.
// Priority: $CONVERTR_CONFIG > $XDG_CONFIG_HOME/convertr/config.toml > ~/.config/convertr/config.toml
func ConfigPath() string {
	if v := os.Getenv("CONVERTR_CONFIG"); v != "" {
		return v
	}
	return filepath.Join(ConfigDir(), "config.toml")
}

func ConfigDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(home(), ".config")
	}
	return filepath.Join(base, "convertr")
}

func CacheDir() string {
	if v := os.Getenv("CONVERTR_CACHE_DIR"); v != "" {
		return v
	}
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		base = filepath.Join(home(), ".cache")
	}
	return filepath.Join(base, "convertr")
}

// WatchDBPath returns the sqlite db path used by convertr watch for idempotency.
func WatchDBPath() string {
	return filepath.Join(CacheDir(), "watch.db")
}

func home() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if h, err := os.UserHomeDir(); err == nil {
		return h
	}
	return "."
}
