package htmltidy

import (
	"io"
	"slices"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

const (
	// Preserve all whitespace - for tags like pre.
	ModePreserve = 1
	// Follow HTML space normalization rules.
	ModeNormalize = 2
	// Transitions from some whitespace to zero whitespace are allowed.
	ModeFree = 3
)

const (
	indentStr = "    "
)

type printer struct {
	w io.StringWriter
}

func (p *printer) indent(level int) {
	p.write(strings.Repeat(indentStr, level))
}

func (p *printer) write(s string) {
	// stringbuilder never errors.
	_, _ = p.w.WriteString(s)
}

// printNode prints the node n at the given indentation level.
// inlineContext indicates if an ancestor is an inline element.
// preserveContext indicate if an ancestor is a preserve element.
func (p *printer) printNode(mode int, indent int, nl ...*html.Node) {
	for _, n := range nl {
		// for i, n := range nl {
		// 	fmt.Printf("print %d mode: %v type: %v indent: %d data: %q\n", i, mode, n.Type, indent, n.Data)
		switch n.Type {
		case html.DocumentNode:
			p.printNode(mode, indent, slices.Collect(n.ChildNodes())...)

		case html.DoctypeNode:
			p.write("<!DOCTYPE " + n.Data + ">\n")

		case html.TextNode:
			switch mode {
			case ModePreserve:
				p.write(n.Data)
			case ModeNormalize:
				if (n.NextSibling == nil || !isInline(n.NextSibling)) && (hasSpaceSuffix(n.Data) || !isInline(n.Parent)) {
					n.Data = strings.TrimRight(n.Data, spaceChars) + "\n"
				}
				p.write(n.Data)
			case ModeFree:
				p.indent(indent)
				text := strings.TrimLeft(n.Data, spaceChars)
				if n.NextSibling == nil {
					text = strings.TrimRight(n.Data, spaceChars) + "\n"
				}
				p.write(text)
				mode = ModeNormalize
			}

		case html.ElementNode:
			name := n.Data
			inline := isInline(n)
			preserve := isPreserve(n)
			childMode := mode
			nextMode := mode // change mode after processing this node.

			switch mode {
			case ModePreserve:
			case ModeNormalize:
				if preserve {
					childMode = ModePreserve
					mode = ModeFree // change mode immediately
				} else if !inline {
					childMode = ModeFree
					mode = ModeFree
				}
			case ModeFree:
				if preserve {
					childMode = ModePreserve
				} else if inline {
					childMode = ModePreserve
					nextMode = ModeNormalize
				}
			}

			if mode == ModeFree {
				p.indent(indent)
			}

			// Opening tag
			p.write("<" + name)
			for _, a := range n.Attr {
				p.write(` ` + a.Key + `="` + a.Val + `"`)
			}
			p.write(">")

			if isVoidElement(n) {
				if mode == ModeFree {
					p.write("\n")
				}
				mode = nextMode
				continue
			}

			if preserve {
				// This is unusual behavior. If this a preserve node,
				// the browser always ignores the first newline, so normalize
				// so that there is always exactly one for aesthetics.
				if isText(n.FirstChild) && !strings.HasPrefix(n.FirstChild.Data, "\n") {
					p.write("\n")
				}
			} else if !inline && mode == ModeFree {
				p.write("\n")
			}

			p.printNode(childMode, indent+1, slices.Collect(n.ChildNodes())...)

			if mode == ModeFree && !(preserve || inline) {
				p.indent(indent)
			}

			// Closing tag
			p.write("</" + name + ">")

			switch mode {
			case ModeNormalize:
				if n.NextSibling != nil &&
					!(isInline(n.NextSibling) || isText(n.NextSibling)) {
					p.write("\n")
				} else if n.NextSibling == nil && !isInline(n.Parent) {
					p.write("\n")
				}
			case ModeFree:
				if !inline || isPreserve(n.NextSibling) || (!isInline(n.NextSibling) && !isText(n.NextSibling)) {
					p.write("\n")
				}
			}
			mode = nextMode

		case html.CommentNode:
			if mode == ModeFree {
				p.indent(indent)
			}
			p.write("<!--" + n.Data + "-->")
			if mode == ModeFree {
				p.write("\n")
			}
		}
	}
}

func hasSpaceSuffix(s string) bool {
	rs := []rune(s)
	return len(rs) > 0 && unicode.IsSpace(rs[len(rs)-1])
}
