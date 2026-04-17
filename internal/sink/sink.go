package sink

// SinkType classifies the output destination.
type SinkType int

const (
	SinkTypeFile   SinkType = iota // single output file
	SinkTypeDir                    // output directory (one file per input)
	SinkTypeStdout                 // write to stdout
)

// Sink describes where converted output is written.
type Sink struct {
	Type     SinkType
	Path     string         // empty for SinkTypeStdout
	Format   string         // canonical target format ID
	Policy   ConflictPolicy // how to handle existing files
	Template string         // optional naming template for SinkTypeDir
}
