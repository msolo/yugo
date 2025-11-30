package build

import (
	"bytes"
	"errors"

	"github.com/msolo/jsonr"
)

type Page struct {
	Params map[string]any
	Body   []byte // content without frontmatter
}

// Extracts JSONR frontmatter iff the file begins with '---'
func ParsePage(src []byte) (Page, error) {
	trimmed := bytes.TrimSpace(src)
	const delim = "---\n"
	if !bytes.HasPrefix(trimmed, []byte(delim)) {
		// No frontmatter
		return Page{Params: map[string]any{}, Body: src}, nil
	}
	trimmed = trimmed[len(delim):]

	// Find matching closing brace for the JSON object
	depth := 0
	end := -1
loop:
	for i, b := range trimmed {
		switch b {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i
				break loop
			}
		}
	}
	if end == -1 {
		return Page{}, errors.New("unclosed JSONR frontmatter")
	}

	fmText := trimmed[:end+1]
	rest := trimmed[end+1:]

	// Look for a frontmatter separator --- on its own line
	if !bytes.HasPrefix(bytes.TrimSpace(rest), []byte(delim)) {
		return Page{}, errors.New("missing --- after JSONR frontmatter")
	}

	rest = rest[len(delim):]

	params := map[string]any{}
	if err := jsonr.Unmarshal(fmText, &params); err != nil {
		return Page{}, err
	}

	return Page{
		Params: params,
		Body:   []byte(rest),
	}, nil
}
