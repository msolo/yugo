package build

import (
	"bytes"
	"encoding/json"
	"html/template"
)

// jsonify encodes a Go value into readable, pretty JSON.
// if encoding fails, it panics since the build should fail.
// Output is marked html-safe to prevent template escaping.
func jsonify(v any) template.HTML {
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(v); err != nil {
		panic(err)
	}

	// json.Encoder.Encode adds one trailing newline — remove it
	out := bytes.TrimRight(buf.Bytes(), "\n")

	return template.HTML(out)
}
