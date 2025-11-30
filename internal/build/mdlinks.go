package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

var SourceFileKey = parser.NewContextKey()

type LinkRewriter struct {
	SiteDir    string
	ContentDir string
}

func (r LinkRewriter) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	// Location of the markdown source file relative to ContentDir.
	srcFile := pc.Get(SourceFileKey).(string)

	walkErr := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		link, ok := n.(*ast.Link)
		if !ok {
			return ast.WalkContinue, nil
		}

		// Skip images entirely.
		if link.Parent() != nil {
			if _, isImg := link.Parent().(*ast.Image); isImg {
				return ast.WalkContinue, nil
			}
		}

		dest := string(link.Destination)

		// Skip external links
		if isExternal(dest) {
			return ast.WalkContinue, nil
		}

		// Split anchor for .md#section pattern
		base, anchor := splitAnchor(dest)

		// Only rewrite links pointing to .md files
		if !strings.HasSuffix(base, ".md") {
			return ast.WalkContinue, nil
		}

		// destPath is relative to ContentDir
		destPath := ""
		if filepath.IsAbs(base) {
			destPath = filepath.Clean(base)
		} else {
			destPath = filepath.Clean(filepath.Join(filepath.Dir(srcFile), base))
		}
		fullDestPath := filepath.Join(r.ContentDir, destPath)

		// println("dest", dest)
		// println("destPath", destPath)
		// println("fullDestPath", fullDestPath)
		if _, err := os.Stat(fullDestPath); err != nil {
			fmt.Fprintf(os.Stderr, "WARN: broken link → %s (resolved as %s)\n", dest, destPath)
		}

		// Now rewrite the URL to .html (keeping the resolved path)
		htmlPath := strings.TrimSuffix(base, ".md") + ".html"
		if anchor != "" {
			htmlPath += "#" + anchor
		}

		link.Destination = []byte(htmlPath)
		return ast.WalkContinue, nil
	})
	if walkErr != nil {
		fmt.Fprintf(os.Stderr, "WARN: error processing Markdown links: %s\n", walkErr)
	}
}

func isExternal(s string) bool {
	return strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "//")
}

// splitAnchor("docs/x.md#sec") → ("docs/x.md", "sec")
func splitAnchor(dest string) (string, string) {
	parts := strings.SplitN(dest, "#", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return dest, ""
}
