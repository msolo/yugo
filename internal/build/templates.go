package build

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type TemplateLoader struct {
	TemplateDir string
}

func (tl *TemplateLoader) Load() (*template.Template, error) {
	tmpl := template.New("").
		Funcs(template.FuncMap{
			"now": time.Now, // expose time.Now()
			// Go html template forbids naked comments for reasons that are probably theorically sound, but practically annoying so we need an escape hatch.
			"htmlComment": func(s template.HTML) template.HTML { return template.HTML("<!--\n" + s + "\n-->") },
			"jsonify":     jsonify,
		})

	// Walk TemplateDir to find *.html
	err := filepath.Walk(tl.TemplateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".html") {
			return nil
		}

		// Compute template name relative to TemplateDir
		rel, err := filepath.Rel(tl.TemplateDir, path)
		if err != nil {
			return err
		}
		// Normalize to forward slashes
		rel = filepath.ToSlash(rel)

		// Parse file
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Add/extend template under its relative name
		_, err = tmpl.New(rel).Parse(string(b))
		return err
	})

	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
