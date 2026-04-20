// Package formats defines the canonical format registry used for routing.
package formats

// Category classifies a format by its broad media type.
type Category string

const (
	CategoryDocument Category = "document"
	CategoryMarkup   Category = "markup"
	CategoryData     Category = "data"
	CategoryImage    Category = "image"
	CategoryAudio    Category = "audio"
	CategoryVideo    Category = "video"
	CategoryTextArt  Category = "text-art"
	CategoryArchive  Category = "archive"
)

// Format describes a single file format.
type Format struct {
	ID         string // canonical id: "md", "docx", "mp4"
	Category   Category
	Extensions []string // includes leading dot: ".md", ".markdown"
	MimeTypes  []string
	Text       bool // true for text-based formats (safe to pipe via stdin)
}

// ID returns the canonical format id for the given extension (without dot).
// Returns empty string if not found.
func ByExtension(ext string) *Format {
	// normalise: remove leading dot, lowercase
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}
	for i := range known {
		for _, e := range known[i].Extensions {
			if len(e) > 0 && e[1:] == ext {
				return &known[i]
			}
		}
	}
	return nil
}

// ByID returns the Format for the given canonical id.
func ByID(id string) *Format {
	for i := range known {
		if known[i].ID == id {
			return &known[i]
		}
	}
	return nil
}

// All returns all registered formats.
func All() []Format {
	out := make([]Format, len(known))
	copy(out, known)
	return out
}

// AllByCategory returns formats filtered by category.
func AllByCategory(c Category) []Format {
	var out []Format
	for _, f := range known {
		if f.Category == c {
			out = append(out, f)
		}
	}
	return out
}
