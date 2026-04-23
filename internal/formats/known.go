package formats

// known is the authoritative list of formats convertr understands.
// Backends declare Capabilities using these IDs.
var known = []Format{
	// ── Documents ──────────────────────────────────────────────────────────
	{ID: "doc", Category: CategoryDocument, Extensions: []string{".doc"}, MimeTypes: []string{"application/msword"}},
	{ID: "docx", Category: CategoryDocument, Extensions: []string{".docx"}, MimeTypes: []string{"application/vnd.openxmlformats-officedocument.wordprocessingml.document"}},
	{ID: "odt", Category: CategoryDocument, Extensions: []string{".odt"}, MimeTypes: []string{"application/vnd.oasis.opendocument.text"}},
	{ID: "rtf", Category: CategoryDocument, Extensions: []string{".rtf"}, MimeTypes: []string{"application/rtf", "text/rtf"}},
	{ID: "epub", Category: CategoryDocument, Extensions: []string{".epub"}, MimeTypes: []string{"application/epub+zip"}},
	{ID: "fb2", Category: CategoryDocument, Extensions: []string{".fb2"}, MimeTypes: []string{"application/x-fictionbook+xml"}, Text: true},
	{ID: "pdf", Category: CategoryDocument, Extensions: []string{".pdf"}, MimeTypes: []string{"application/pdf"}},
	{ID: "pptx", Category: CategoryDocument, Extensions: []string{".pptx"}, MimeTypes: []string{"application/vnd.openxmlformats-officedocument.presentationml.presentation"}},
	{ID: "odp", Category: CategoryDocument, Extensions: []string{".odp"}, MimeTypes: []string{"application/vnd.oasis.opendocument.presentation"}},

	// ── Markup ─────────────────────────────────────────────────────────────
	{ID: "md", Category: CategoryMarkup, Extensions: []string{".md", ".markdown", ".mdown"}, MimeTypes: []string{"text/markdown"}, Text: true},
	{ID: "html", Category: CategoryMarkup, Extensions: []string{".html", ".htm"}, MimeTypes: []string{"text/html"}, Text: true},
	{ID: "rst", Category: CategoryMarkup, Extensions: []string{".rst"}, MimeTypes: []string{"text/x-rst"}, Text: true},
	{ID: "adoc", Category: CategoryMarkup, Extensions: []string{".adoc", ".asciidoc", ".asc"}, MimeTypes: []string{"text/asciidoc"}, Text: true},
	{ID: "tex", Category: CategoryMarkup, Extensions: []string{".tex"}, MimeTypes: []string{"application/x-tex"}, Text: true},
	{ID: "org", Category: CategoryMarkup, Extensions: []string{".org"}, MimeTypes: []string{}, Text: true},
	{ID: "txt", Category: CategoryMarkup, Extensions: []string{".txt", ".text"}, MimeTypes: []string{"text/plain"}, Text: true},
	{ID: "typst", Category: CategoryMarkup, Extensions: []string{".typ"}, MimeTypes: []string{"text/x-typst"}, Text: true},
	{ID: "ipynb", Category: CategoryMarkup, Extensions: []string{".ipynb"}, MimeTypes: []string{"application/x-ipynb+json"}, Text: true},
	{ID: "mediawiki", Category: CategoryMarkup, Extensions: []string{".mediawiki", ".wiki"}, MimeTypes: []string{}, Text: true},
	{ID: "dokuwiki", Category: CategoryMarkup, Extensions: []string{".dokuwiki"}, MimeTypes: []string{}, Text: true},
	{ID: "jira", Category: CategoryMarkup, Extensions: []string{".jira"}, MimeTypes: []string{}, Text: true},
	{ID: "textile", Category: CategoryMarkup, Extensions: []string{".textile"}, MimeTypes: []string{"text/x-textile"}, Text: true},
	{ID: "docbook", Category: CategoryMarkup, Extensions: []string{".docbook", ".dbk"}, MimeTypes: []string{"application/docbook+xml"}, Text: true},

	// ── Data ───────────────────────────────────────────────────────────────
	{ID: "json", Category: CategoryData, Extensions: []string{".json"}, MimeTypes: []string{"application/json"}, Text: true},
	{ID: "yaml", Category: CategoryData, Extensions: []string{".yaml", ".yml"}, MimeTypes: []string{"application/yaml", "text/yaml"}, Text: true},
	{ID: "toml", Category: CategoryData, Extensions: []string{".toml"}, MimeTypes: []string{"application/toml"}, Text: true},
	{ID: "csv", Category: CategoryData, Extensions: []string{".csv"}, MimeTypes: []string{"text/csv"}, Text: true},
	{ID: "tsv", Category: CategoryData, Extensions: []string{".tsv"}, MimeTypes: []string{"text/tab-separated-values"}, Text: true},
	{ID: "xlsx", Category: CategoryData, Extensions: []string{".xlsx"}, MimeTypes: []string{"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"}},
	{ID: "ods", Category: CategoryData, Extensions: []string{".ods"}, MimeTypes: []string{"application/vnd.oasis.opendocument.spreadsheet"}},
	{ID: "bibtex", Category: CategoryData, Extensions: []string{".bib", ".bibtex"}, MimeTypes: []string{"application/x-bibtex"}, Text: true},
	{ID: "csljson", Category: CategoryData, Extensions: []string{".csljson", ".csl.json"}, MimeTypes: []string{"application/vnd.citationstyles.csl+json"}, Text: true},
	{ID: "opml", Category: CategoryData, Extensions: []string{".opml"}, MimeTypes: []string{"text/x-opml"}, Text: true},

	// ── Image ──────────────────────────────────────────────────────────────
	{ID: "jpg", Category: CategoryImage, Extensions: []string{".jpg", ".jpeg"}, MimeTypes: []string{"image/jpeg"}},
	{ID: "png", Category: CategoryImage, Extensions: []string{".png"}, MimeTypes: []string{"image/png"}},
	{ID: "gif", Category: CategoryImage, Extensions: []string{".gif"}, MimeTypes: []string{"image/gif"}},
	{ID: "webp", Category: CategoryImage, Extensions: []string{".webp"}, MimeTypes: []string{"image/webp"}},
	{ID: "avif", Category: CategoryImage, Extensions: []string{".avif"}, MimeTypes: []string{"image/avif"}},
	{ID: "tiff", Category: CategoryImage, Extensions: []string{".tiff", ".tif"}, MimeTypes: []string{"image/tiff"}},
	{ID: "bmp", Category: CategoryImage, Extensions: []string{".bmp"}, MimeTypes: []string{"image/bmp"}},
	{ID: "heic", Category: CategoryImage, Extensions: []string{".heic", ".heif"}, MimeTypes: []string{"image/heic"}},
	{ID: "svg", Category: CategoryImage, Extensions: []string{".svg"}, MimeTypes: []string{"image/svg+xml"}, Text: true},

	// ── Audio ──────────────────────────────────────────────────────────────
	{ID: "mp3", Category: CategoryAudio, Extensions: []string{".mp3"}, MimeTypes: []string{"audio/mpeg"}},
	{ID: "flac", Category: CategoryAudio, Extensions: []string{".flac"}, MimeTypes: []string{"audio/flac"}},
	{ID: "aac", Category: CategoryAudio, Extensions: []string{".aac"}, MimeTypes: []string{"audio/aac"}},
	{ID: "ogg", Category: CategoryAudio, Extensions: []string{".ogg"}, MimeTypes: []string{"audio/ogg"}},
	{ID: "wav", Category: CategoryAudio, Extensions: []string{".wav"}, MimeTypes: []string{"audio/wav"}},
	{ID: "m4a", Category: CategoryAudio, Extensions: []string{".m4a"}, MimeTypes: []string{"audio/mp4"}},
	{ID: "opus", Category: CategoryAudio, Extensions: []string{".opus"}, MimeTypes: []string{"audio/opus"}},

	// ── Video ──────────────────────────────────────────────────────────────
	{ID: "mp4", Category: CategoryVideo, Extensions: []string{".mp4"}, MimeTypes: []string{"video/mp4"}},
	{ID: "mkv", Category: CategoryVideo, Extensions: []string{".mkv"}, MimeTypes: []string{"video/x-matroska"}},
	{ID: "webm", Category: CategoryVideo, Extensions: []string{".webm"}, MimeTypes: []string{"video/webm"}},
	{ID: "mov", Category: CategoryVideo, Extensions: []string{".mov"}, MimeTypes: []string{"video/quicktime"}},
	{ID: "avi", Category: CategoryVideo, Extensions: []string{".avi"}, MimeTypes: []string{"video/x-msvideo"}},

	// ── Text Art ───────────────────────────────────────────────────────────
	{ID: "ascii", Category: CategoryTextArt, Extensions: []string{".asc"}, MimeTypes: []string{}, Text: true},
}
