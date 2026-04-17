package config

// MergeProfile applies profile overrides on top of cfg.
// Zero/empty fields in the profile leave the corresponding field unchanged.
func MergeProfile(cfg Config, profileName string) Config {
	p, ok := cfg.Profile[profileName]
	if !ok {
		return cfg
	}
	if p.Quality != 0 {
		cfg.Defaults.Quality = p.Quality
	}
	if p.Workers != 0 {
		cfg.Defaults.Workers = p.Workers
	}
	if p.OnError != "" {
		cfg.Defaults.OnError = p.OnError
	}
	if p.OnConflict != "" {
		cfg.Defaults.OnConflict = p.OnConflict
	}
	return cfg
}
