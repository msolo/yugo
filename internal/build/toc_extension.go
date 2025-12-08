package build

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// TOCItem represents an entry in the table of contents
// This doesn't have to be a well-formed tree, so record level rather than
// making the structure recursive.
type TOCItem struct {
	Level int
	ID    string
	Text  string
}

type TOCExtension struct {
	Items *[]TOCItem
}

func (e *TOCExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&tocTransformer{items: e.Items}, 0),
		),
	)
}

type tocTransformer struct {
	items *[]TOCItem
}

func (t *tocTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	if t.items == nil {
		return
	}

	// Clear existing items
	*t.items = (*t.items)[:0]

	// Walk the document tree looking for headings
	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		heading, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}

		text := extractText(heading, reader.Source())

		// Get heading ID - this needs to be enabled
		id := ""
		if heading.Attributes() != nil {
			if idAttr, ok := heading.Attribute([]byte("id")); ok {
				if idBytes, ok := idAttr.([]byte); ok {
					id = string(idBytes)
				}
			}
		}

		*t.items = append(*t.items, TOCItem{
			Level: heading.Level,
			ID:    id,
			Text:  text,
		})

		return ast.WalkContinue, nil
	})
	if err != nil {
		log.Println("error walking ast:", err)
	}
}

// extractText recursively extracts text content from a node
func extractText(n ast.Node, source []byte) string {
	var text string
	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		switch c := child.(type) {
		case *ast.Text:
			text += string(c.Segment.Value(source))
		case *ast.String:
			text += string(c.Value)
		default:
			// Recursively extract from child nodes
			text += extractText(c, source)
		}
	}
	return text
}

func GenerateTOCFiltered(items []TOCItem, minLevel, maxLevel int) string {
	filtered := []TOCItem{}
	for _, item := range items {
		if item.Level >= minLevel && item.Level <= maxLevel {
			filtered = append(filtered, item)
		}
	}
	return GenerateTOC(filtered)
}

// GenerateTOC creates a table of contents in html from heading items
func GenerateTOC(items []TOCItem) string {
	if len(items) == 0 {
		return ""
	}

	var buf bytes.Buffer
	buf.WriteString("<nav class=\"toc\">\n")
	buf.WriteString("<h2>Table of Contents</h2>\n")
	buf.WriteString("<ul>\n")

	currentLevel := 0
	for _, item := range items {
		for currentLevel > item.Level {
			buf.WriteString("</ul>\n</li>\n")
			currentLevel--
		}

		for currentLevel < item.Level {
			if currentLevel > 0 {
				buf.WriteString("<ul>\n")
			}
			currentLevel++
		}

		buf.WriteString(fmt.Sprintf("<li><a href=\"#%s\">%s</a>", item.ID, template.HTMLEscapeString(item.Text)))

		isLast := false
		if len(items) > 0 {
			// Peek at next item to see if we should close this <li>
			nextIdx := -1
			for i, it := range items {
				if it == item {
					nextIdx = i + 1
					break
				}
			}
			if nextIdx >= len(items) || items[nextIdx].Level <= item.Level {
				buf.WriteString("</li>\n")
				isLast = true
			}
		}

		if !isLast {
			buf.WriteString("\n")
		}
	}

	for currentLevel > 0 {
		buf.WriteString("</ul>\n</li>\n")
		currentLevel--
	}

	buf.WriteString("</ul>\n")
	buf.WriteString("</nav>\n")
	return buf.String()
}
