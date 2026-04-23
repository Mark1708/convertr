package pandoc

import (
	"slices"
	"strings"
	"testing"

	"github.com/Mark1708/convertr/internal/backend"
	"github.com/Mark1708/convertr/internal/fonts"
)

// containsPair reports whether "-V name=value" (as two adjacent tokens)
// appears anywhere in args.
func containsPair(args []string, flag, value string) bool {
	for i := range len(args) - 1 {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}

// countVariable returns how many times a pandoc variable with the given
// prefix ("name=" or "name:") appears in args across all -V / --variable
// shapes. Used to assert that defaults do not duplicate user-supplied values.
func countVariable(args []string, name string) int {
	n := 0
	for i, a := range args {
		switch {
		case (a == "-V" || a == "--variable") && i+1 < len(args):
			if strings.HasPrefix(args[i+1], name+"=") || strings.HasPrefix(args[i+1], name+":") {
				n++
			}
		case strings.HasPrefix(a, "-V"):
			rest := strings.TrimPrefix(a, "-V")
			if strings.HasPrefix(rest, name+"=") || strings.HasPrefix(rest, name+":") {
				n++
			}
		case strings.HasPrefix(a, "--variable="):
			rest := strings.TrimPrefix(a, "--variable=")
			if strings.HasPrefix(rest, name+"=") || strings.HasPrefix(rest, name+":") {
				n++
			}
		}
	}
	return n
}

func testFonts() fonts.Config {
	return fonts.Config{Mainfont: "PT Serif", Monofont: "Menlo", Sansfont: "Helvetica Neue"}
}

func TestBuildArgs_PDFWithXelatexInjectsFontDefaults(t *testing.T) {
	opts := backend.Options{
		Named: map[string]string{
			"step.from":          "md",
			"step.to":            "pdf",
			"pandoc.pdf_engine":  "xelatex",
		},
		Fonts: testFonts(),
	}
	args := buildArgs("in.md", "out.pdf", opts)

	if !containsPair(args, "-V", "mainfont=PT Serif") {
		t.Errorf("expected -V mainfont=PT Serif, got %v", args)
	}
	if !containsPair(args, "-V", "monofont=Menlo") {
		t.Errorf("expected -V monofont=Menlo, got %v", args)
	}
	if !containsPair(args, "-V", "sansfont=Helvetica Neue") {
		t.Errorf("expected -V sansfont=Helvetica Neue, got %v", args)
	}
	if !containsPair(args, "-V", "geometry:margin=2cm") {
		t.Errorf("expected default geometry, got %v", args)
	}
}

func TestBuildArgs_NonPDFSkipsFontDefaults(t *testing.T) {
	opts := backend.Options{
		Named: map[string]string{
			"step.from": "md",
			"step.to":   "docx",
		},
		Fonts: testFonts(),
	}
	args := buildArgs("in.md", "out.docx", opts)

	if countVariable(args, "mainfont") != 0 {
		t.Errorf("non-PDF output should not inject mainfont, got %v", args)
	}
	if countVariable(args, "geometry") != 0 {
		t.Errorf("non-PDF output should not inject geometry, got %v", args)
	}
}

func TestBuildArgs_NamedOverrideBeatsDefault(t *testing.T) {
	opts := backend.Options{
		Named: map[string]string{
			"step.from":          "md",
			"step.to":            "pdf",
			"pandoc.pdf_engine":  "xelatex",
			"pandoc.mainfont":    "Georgia",
		},
		Fonts: testFonts(),
	}
	args := buildArgs("in.md", "out.pdf", opts)

	if !containsPair(args, "-V", "mainfont=Georgia") {
		t.Errorf("expected mainfont=Georgia, got %v", args)
	}
	if containsPair(args, "-V", "mainfont=PT Serif") {
		t.Errorf("default mainfont should not appear alongside user override, got %v", args)
	}
	if countVariable(args, "mainfont") != 1 {
		t.Errorf("mainfont should appear exactly once, got %d times in %v", countVariable(args, "mainfont"), args)
	}
}

func TestBuildArgs_ExtraArgsVariableSuppressesDefault(t *testing.T) {
	opts := backend.Options{
		Named: map[string]string{
			"step.from":         "md",
			"step.to":           "pdf",
			"pandoc.pdf_engine": "xelatex",
		},
		ExtraArgs: []string{"-V", "mainfont=Bar"},
		Fonts:     testFonts(),
	}
	args := buildArgs("in.md", "out.pdf", opts)

	if countVariable(args, "mainfont") != 1 {
		t.Errorf("mainfont should appear exactly once (user-supplied), got %d in %v", countVariable(args, "mainfont"), args)
	}
	// Other defaults still injected.
	if countVariable(args, "monofont") != 1 {
		t.Errorf("monofont default should still be injected, got %d in %v", countVariable(args, "monofont"), args)
	}
}

func TestBuildArgs_ExtraArgsGeometrySuppressesDefault(t *testing.T) {
	opts := backend.Options{
		Named: map[string]string{
			"step.from":         "md",
			"step.to":           "pdf",
			"pandoc.pdf_engine": "xelatex",
		},
		ExtraArgs: []string{"--variable=geometry:margin=1in"},
		Fonts:     testFonts(),
	}
	args := buildArgs("in.md", "out.pdf", opts)

	if countVariable(args, "geometry") != 1 {
		t.Errorf("geometry should appear exactly once (user-supplied), got %d in %v", countVariable(args, "geometry"), args)
	}
}

func TestBuildArgs_PdflatexEngineSkipsFontspec(t *testing.T) {
	opts := backend.Options{
		Named: map[string]string{
			"step.from":         "md",
			"step.to":           "pdf",
			"pandoc.pdf_engine": "pdflatex",
		},
		Fonts: testFonts(),
	}
	args := buildArgs("in.md", "out.pdf", opts)

	if countVariable(args, "mainfont") != 0 {
		t.Errorf("pdflatex should not receive mainfont (no fontspec support), got %v", args)
	}
	if countVariable(args, "monofont") != 0 {
		t.Errorf("pdflatex should not receive monofont, got %v", args)
	}
	if countVariable(args, "geometry") != 0 {
		t.Errorf("pdflatex should not receive geometry default, got %v", args)
	}
	if !containsPair(args, "--pdf-engine=pdflatex", "") && !argsContain(args, "--pdf-engine=pdflatex") {
		t.Errorf("expected --pdf-engine=pdflatex in %v", args)
	}
}

func TestBuildArgs_HasVariableShapes(t *testing.T) {
	cases := []struct {
		name    string
		extra   []string
		lookFor string
		want    bool
	}{
		{"separate -V name=val", []string{"-V", "mainfont=X"}, "mainfont", true},
		{"separate -V name:val (colon-style)", []string{"-V", "geometry:margin=1cm"}, "geometry", true},
		{"glued -Vname=val", []string{"-Vmainfont=X"}, "mainfont", true},
		{"separate --variable", []string{"--variable", "mainfont=X"}, "mainfont", true},
		{"combined --variable=name=val", []string{"--variable=mainfont=X"}, "mainfont", true},
		{"absent", []string{"-V", "other=X"}, "mainfont", false},
		{"prefix collision", []string{"-V", "mainfontextra=X"}, "mainfont", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasVariable(tc.extra, tc.lookFor); got != tc.want {
				t.Errorf("hasVariable(%v, %q) = %v, want %v", tc.extra, tc.lookFor, got, tc.want)
			}
		})
	}
}

func TestBuildArgs_AppliesFromAndToFlags(t *testing.T) {
	opts := backend.Options{
		Named: map[string]string{
			"step.from": "md",
			"step.to":   "html",
		},
	}
	args := buildArgs("in.md", "out.html", opts)

	if !containsPair(args, "--from", "markdown") {
		t.Errorf("expected --from markdown, got %v", args)
	}
	if !containsPair(args, "--to", "html") {
		t.Errorf("expected --to html, got %v", args)
	}
}

func TestBuildArgs_ExtraArgsAppendedLast(t *testing.T) {
	opts := backend.Options{
		Named: map[string]string{
			"step.from": "md",
			"step.to":   "html",
		},
		ExtraArgs: []string{"--wrap=none"},
	}
	args := buildArgs("in.md", "out.html", opts)

	if args[len(args)-1] != "--wrap=none" {
		t.Errorf("ExtraArgs should be appended last, got %v", args)
	}
}

// argsContain reports whether a single token appears anywhere in args.
func argsContain(args []string, token string) bool {
	return slices.Contains(args, token)
}
