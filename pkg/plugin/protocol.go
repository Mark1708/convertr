// Package plugin defines the public protocol for convertr external plugins.
//
// A plugin is an executable named convertr-<NAME> that implements two sub-commands:
//
//	convertr-NAME capabilities
//	    Writes a JSON array of Capability objects to stdout and exits 0.
//
//	convertr-NAME convert --from FROM --to TO --input IN --output OUT [--opt key=value ...]
//	    Converts IN to OUT. Exits 0 on success, non-zero on failure.
//	    On failure, writes a single line error message to stderr.
package plugin

// Capability describes one conversion edge a plugin can perform.
type Capability struct {
	From string `json:"from"` // source format ID (e.g. "md")
	To   string `json:"to"`   // target format ID (e.g. "html")
	Cost int    `json:"cost"` // routing cost (1–10; lower = preferred)
}
