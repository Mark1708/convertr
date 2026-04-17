package sink

import (
	"os"
	"strings"
)

// ResolveSink determines the sink type from the output path string.
//   - "-"  or empty → SinkTypeStdout
//   - path ending in "/" or existing directory → SinkTypeDir
//   - everything else → SinkTypeFile
func ResolveSink(output, toFormat string) *Sink {
	if output == "-" || output == "" {
		return &Sink{Type: SinkTypeStdout, Format: toFormat}
	}
	if strings.HasSuffix(output, "/") {
		return &Sink{Type: SinkTypeDir, Path: strings.TrimRight(output, "/"), Format: toFormat}
	}
	if fi, err := os.Stat(output); err == nil && fi.IsDir() {
		return &Sink{Type: SinkTypeDir, Path: output, Format: toFormat}
	}
	return &Sink{Type: SinkTypeFile, Path: output, Format: toFormat}
}
