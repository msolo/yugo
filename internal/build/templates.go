package build

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"
)

type TemplateLoader struct {
	TemplateDir string
	StaticDir   string
}

func (tl *TemplateLoader) Load() (*template.Template, error) {
	tmpl := template.New("").
		Funcs(template.FuncMap{
			"now": time.Now, // expose time.Now()
			// Go html template forbids naked comments for reasons that are probably theorically sound, but practically annoying so we need an escape hatch.
			"htmlComment": func(s template.HTML) template.HTML { return template.HTML("<!--\n" + s + "\n-->") },
			"jsonify":     jsonify,
		})

	maybeAddTemplate := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(info.Name())
		switch ext {
		case ".html", ".md", ".tmpl":
		default:
			return nil
		}

		// Compute template name relative to TemplateDir
		rel, err := filepath.Rel(tl.TemplateDir, path)
		if err != nil {
			return err
		}
		// Normalize to forward slashes
		rel = filepath.ToSlash(rel)

		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed reading %s: %w", path, err)
		}

		// Add/extend template under its relative name
		_, err = tmpl.New(rel).Parse(string(b))
		if err != nil {
			return fmt.Errorf("failed parsing %s: %w", rel, err)
		}

		return err
	}

	maybeAddInclude := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(info.Name())
		switch ext {
		case ".html", ".md":
		case ".css", ".js", ".txt":
			// Allow these as "templates" to be low-budget includes.
		default:
			return nil
		}

		// Compute template name relative to parent of static dir so
		// the include is obvious.
		rel, err := filepath.Rel(tl.StaticDir, path)
		if err != nil {
			return err
		}
		// Normalize to forward slashes
		rel = filepath.ToSlash(rel)

		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed reading %s: %w", path, err)
		}

		// Add/extend template under implicit prefix "static" so that the include
		// is obvious.
		_, err = tmpl.New(filepath.Join("static", rel)).Parse(string(b))
		if err != nil {
			return fmt.Errorf("failed parsing %s: %w", rel, err)
		}
		return err
	}

	// In theory, static should be a distinct namespace, but just
	// in case, we do these first so true templates take precedence.
	// static content is optional, so only load it if it exists.
	if _, err := os.Stat(tl.StaticDir); err == nil {
		err := filepath.Walk(tl.StaticDir, maybeAddInclude)
		if err != nil {
			return nil, err
		}
	}

	err := filepath.Walk(tl.TemplateDir, maybeAddTemplate)
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}
