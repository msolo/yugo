package build

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ianbruene/go-difflib/difflib"
	"github.com/msolo/yugo/internal/htmltidy"
)

func TestRenderFile(t *testing.T) {
	tmp := t.TempDir()

	// Create directory structure
	contentDir := filepath.Join(tmp, "content")
	templateDir := filepath.Join(tmp, "templates")
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatal(err)
	}

	input := `# Test Header

Some ~~fine~~ content.`
	expected := `<html><head></head><body><h1 id="test-header">Test Header</h1><p>Some <del>fine</del> content.</p></body></html>`

	// Create markdown file
	mdPath := filepath.Join(contentDir, "test.md")
	if err := os.WriteFile(mdPath, []byte(input), 0644); err != nil {
		t.Fatal(err)
	}

	// Create simple template as internally required.
	baseTemplate := `<html><head></head><body>{{.Content}}</body></html>`
	templatePath := filepath.Join(templateDir, "base.html")
	if err := os.WriteFile(templatePath, []byte(baseTemplate), 0644); err != nil {
		t.Fatal(err)
	}

	opts := &Options{&RawOptions{
		SiteDir: tmp,
	}}

	tmpl, err := loadTemplates(opts)
	if err != nil {
		t.Fatal(err)
	}

	siteConfig := map[string]any{}

	relPath := "test.md"

	result, err := renderFile(mdPath, relPath, tmpl, opts, siteConfig)
	if err != nil {
		t.Fatal(err)
	}

	result = normalize(result)
	expected = normalize(expected)
	if expected != result {
		diff := strDiff(expected, result)
		t.Fatalf("output diff:\n--- expected ---\n%s\n--- output ---\n%s\n\n%s",
			diff, expected, result)
	}
}

// Normalize whitespace for comparison
func normalize(s string) string {
	out, _ := htmltidy.NormalizeHTML(s)
	return out
}

func strDiff(expected, out string) string {
	// Pick a character to help spot whitespace issues in black and white.
	const visualSpace = "⋅" // "∘" "∙"
	const visualTab = "→   "

	expected = strings.ReplaceAll(expected, " ", visualSpace)
	out = strings.ReplaceAll(out, " ", visualSpace)
	expected = strings.ReplaceAll(expected, "\t", visualTab)
	out = strings.ReplaceAll(out, "\t", visualTab)

	diff, err := difflib.GetUnifiedDiffString(difflib.LineDiffParams{
		A:        slices.Collect(strings.Lines(expected)),
		FromFile: "expected",
		B:        slices.Collect(strings.Lines(out)),
		ToFile:   "out",
	})
	if err != nil {
		panic(err)
	}
	return diff
}
