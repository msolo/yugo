package htmltidy

import (
	"slices"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

const (
	spaceChars = " \n\t"
)

// NormalizeHTML parses an entire HTML document and returns a tidied version.
//
// // Fragment behavior isn't defined.
//
// It pretty-prints the DOM:
//   - text nodes have their whitespace normalized, runs of adjacent spaces are collapsed to a single space.
//   - whitespace-only text nodes are pruned
//   - inline elements are not indented
//   - <pre>, <textarea>, <script>, <style> preserve all contents
func NormalizeHTML(htmlIn string) (string, error) {
	root, err := html.Parse(strings.NewReader(htmlIn))
	if err != nil {
		return "", err
	}

	normalizeWhitespace(ModeFree, root)

	out := &strings.Builder{}
	p := &printer{w: out}
	p.printNode(ModeFree, 0, root)

	return out.String(), nil
}

// NOTE: it's possible that some text nodes will have trailing whitespace.
func normalizeWhitespace(mode int, nl ...*html.Node) {
	defer func() {
		for _, n := range nl {
			if n.Type == html.TextNode && n.Data == "" {
				n.Parent.RemoveChild(n)
			}
		}
	}()

	for _, n := range nl {
		// for i, n := range nl {
		// 	fmt.Printf("norm %d mode: %v type: %v data: %q\n", i, mode, n.Type, n.Data)
		switch n.Type {
		case html.ElementNode:
			childMode := mode
			switch mode {
			case ModePreserve:
				// preserve all children verbatim
				continue
			case ModeNormalize:
				if isPreserve(n) {
					childMode = ModePreserve
				} else if !isInline(n) {
					mode = ModeFree
					childMode = ModeFree
				}
			case ModeFree:
				if isPreserve(n) {
					childMode = ModePreserve
				} else if isInline(n) {
					childMode = ModeNormalize
					mode = ModeNormalize
				}
			}

			normalizeWhitespace(childMode, slices.Collect(n.ChildNodes())...)
		case html.TextNode:
			switch mode {
			case ModePreserve:
				// preserve all children verbatim
				continue
			case ModeNormalize:
				n.Data = collapseWhitespace(n.Data)
				if isAllWhitespace(n.Data) && !isInline(n.PrevSibling) && !isInline(n.NextSibling) {
					n.Data = ""
				}
			case ModeFree:
				n.Data = collapseWhitespace(n.Data)
				if isAllWhitespace(n.Data) {
					n.Data = ""
				} else {
					// Keep the prefix clear for indent below block nodes.
					if n.PrevSibling == nil && !isInline(n.Parent) {
						n.Data = strings.TrimLeft(n.Data, spaceChars)
					}
					if n.NextSibling == nil && !isInline(n.Parent) {
						n.Data = strings.TrimRight(n.Data, spaceChars)
					}
					// Once we have seen real text, we need to normalize.
					mode = ModeNormalize
				}
			}
		default:
			normalizeWhitespace(mode, slices.Collect(n.ChildNodes())...)
		}
	}
}

// isPreserve reports whether n is within a tag that preserves
// raw text content (<pre>, <script>, <style>, <textarea>).
func isPreserve(n *html.Node) bool {
	if n == nil {
		return false
	}
	if n.Type != html.ElementNode {
		return false
	}

	switch strings.ToLower(n.Data) {
	case "pre", "script", "style", "textarea":
		return true
	}
	return false
}

// isVoidElement determines whether a tag is a void element.
// Void elements never have closing tags.
func isVoidElement(n *html.Node) bool {
	if n == nil {
		return false
	}
	if n.Type != html.ElementNode {
		return false
	}

	switch strings.ToLower(n.Data) {
	case "area", "base", "br", "col", "embed", "hr", "img",
		"input", "link", "meta", "param", "source", "track", "wbr":
		return true
	}
	return false
}

// isInline reports whether the element is considered inline
// for formatting purposes.
func isInline(n *html.Node) bool {
	if n == nil {
		return false
	}
	if n.Type != html.ElementNode {
		return false
	}

	switch strings.ToLower(n.Data) {
	case "a", "abbr", "b", "bdi", "bdo", "cite",
		"code", "data", "del", "dfn", "em", "i", "ins", "kbd", "label",
		"mark", "q", "rp", "rt", "ruby", "s", "samp",
		"small", "span", "strong", "sub", "sup", "time",
		"u", "var", "wbr":
		return true
	}
	return false
}

func isText(n *html.Node) bool {
	return n != nil && n.Type == html.TextNode
}

func isAllWhitespace(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// CollapseWhitespace replaces every run of one or more whitespace
// characters with a single space. Leading and trailing whitespace
// are also reduced to at most one space.
func collapseWhitespace(s string) string {
	out := make([]rune, 0, len(s))
	inSpace := false

	for _, r := range s {
		if unicode.IsSpace(r) {
			if !inSpace {
				out = append(out, ' ')
				inSpace = true
			}
		} else {
			out = append(out, r)
			inSpace = false
		}
	}

	return string(out)
}
