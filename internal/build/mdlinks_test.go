package build

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

func renderWith(t *testing.T, site string, mdText string) string {
	mdPath := ""
	return renderWithPath(t, site, mdPath, mdText)
}

func renderWithPath(t *testing.T, site string, mdPath string, mdText string) string {
	t.Helper()

	mr := LinkRewriter{
		SiteDir:    site,
		ContentDir: site + "/content",
	}

	md := goldmark.New(
		goldmark.WithRendererOptions(html.WithUnsafe()), // Allow raw HTML in markdown
		goldmark.WithParserOptions(
			parser.WithASTTransformers(
				util.Prioritized(mr, 100),
			),
		),
	)

	pc := parser.NewContext()
	pc.Set(SourceFileKey, mdPath)

	var buf bytes.Buffer
	if err := md.Convert([]byte(mdText), &buf, parser.WithContext(pc)); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

func TestRewriteSimple(t *testing.T) {
	tmp := t.TempDir()
	os.Mkdir(filepath.Join(tmp, "content"), 0755)
	os.WriteFile(filepath.Join(tmp, "content", "page.md"), []byte("hi"), 0644)

	out := renderWith(t, tmp, `[x](page.md)`)
	expected := `href="page.html"`
	if !strings.Contains(out, expected) {
		t.Fatalf("expected rewritten link: %s got: %s", expected, out)
	}
}

func TestPreserveAnchor(t *testing.T) {
	tmp := t.TempDir()
	os.Mkdir(filepath.Join(tmp, "content"), 0755)
	os.WriteFile(filepath.Join(tmp, "content", "a.md"), []byte("hi"), 0644)

	out := renderWith(t, tmp, `[x](a.md#sec)`)
	expected := `href="a.html#sec"`
	if !strings.Contains(out, expected) {
		t.Fatalf("expected anchor preserved: %s got: %s", expected, out)
	}
}

func TestSkipImages(t *testing.T) {
	tmp := t.TempDir()
	os.Mkdir(filepath.Join(tmp, "content"), 0755)
	os.WriteFile(filepath.Join(tmp, "content", "pic.md"), []byte("hi"), 0644)

	out := renderWith(t, tmp, `![alt](pic.md)`)
	expected := `src="pic.md"`
	if !strings.Contains(out, expected) {
		t.Fatalf("expected image link untouched: %s got: %s", expected, out)
	}
}

func TestBrokenLinksWarn(t *testing.T) {
	tmp := t.TempDir()
	os.Mkdir(filepath.Join(tmp, "content"), 0755)

	// Simulating a broken link: nosuch.md does not exist
	out := renderWith(t, tmp, `[x](expect-broken-link-warning.md)`)

	// FIXME: This doesn't actually test that we sent a warning to stderr.
	// we expect it to be rewritten to .html anyway
	expected := `href="expect-broken-link-warning.html"`
	if !strings.Contains(out, expected) {
		t.Fatalf("expected broken link rewritten: %s got: %s", expected, out)
	}
}
func TestRelativeLinkResolution(t *testing.T) {
	tmp := t.TempDir()

	// Create content structure
	os.MkdirAll(filepath.Join(tmp, "content/docs/apps"), 0755)
	os.WriteFile(filepath.Join(tmp, "content/docs/intro.md"), []byte("hi"), 0644)
	os.WriteFile(filepath.Join(tmp, "content/docs/apps/page.md"), []byte("hi"), 0644)

	mdPath := "docs/apps/page.md"
	md := `[intro](../intro.md)`

	out := renderWithPath(t, tmp, mdPath, md)
	expected := `href="../intro.html"`
	if !strings.Contains(out, expected) {
		t.Fatalf("expected: %s got: %s", expected, out)
	}

	md = `[intro](/docs/intro.md)`

	out = renderWithPath(t, tmp, mdPath, md)
	expected = `href="/docs/intro.html"`
	if !strings.Contains(out, expected) {
		t.Fatalf("expected: %s got: %s", expected, out)
	}
}
