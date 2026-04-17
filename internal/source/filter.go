package source

// DirOpts configures DirSource behaviour.
type DirOpts struct {
	MaxDepth     int      // 0 = unlimited
	IncludeGlobs []string // only include matching basenames; empty = include all
	ExcludeGlobs []string // exclude matching basenames
	MinSize      int64    // bytes; 0 = no minimum
	MaxSize      int64    // bytes; 0 = no maximum
}
