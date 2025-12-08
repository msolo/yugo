package build

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/msolo/jsonr"
	"github.com/msolo/yugo/internal/htmltidy"
	"github.com/msolo/yugo/internal/resources"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

// RawOptions can be read from the CLI or config, but should't be used by the
// rest of the application.
type RawOptions struct {
	Host         string `json:"-"`
	Port         int    `json:"-"`
	SiteDir      string `json:"-"`
	OutDir       string `json:"OutDir"`
	LiveReload   bool   `json:"-"`
	TidyHTML     bool   `json:"-"`
	BaseTemplate string `json:"-"`
}

// Allow certain options read from config to be merged with values from
// CLI flags.
func (o1 *RawOptions) mergeConfig(o2 *RawOptions) {
	if o1.OutDir == "" {
		o1.OutDir = o2.OutDir
	}
	if o1.BaseTemplate == "" {
		o1.BaseTemplate = o2.BaseTemplate
	}
}

type Options struct {
	rawOptions *RawOptions
}

func (o *Options) MergeConfig() error {
	o2, err := ReadConfig(o.rawOptions.SiteDir)
	if err != nil {
		return err
	}
	o.rawOptions.mergeConfig(o2)
	return nil
}

func (o Options) Host() string {
	return o.rawOptions.Host
}

func (o Options) Port() int {
	return o.rawOptions.Port
}

func (o Options) SiteDir() string {
	return o.rawOptions.SiteDir
}

func (o Options) OutDir() string {
	outDir := "public"
	if o.rawOptions.OutDir != "" {
		outDir = o.rawOptions.OutDir
	}
	return cleanJoin(o.rawOptions.SiteDir, outDir)
}

func (o Options) ContentDir() string {
	return cleanJoin(o.rawOptions.SiteDir, "content")
}

func (o Options) StaticDir() string {
	return cleanJoin(o.rawOptions.SiteDir, "static")
}

func (o Options) TemplatesDir() string {
	return cleanJoin(o.rawOptions.SiteDir, "templates")
}

func (o Options) LiveReload() bool {
	return o.rawOptions.LiveReload
}

func (o Options) TidyHTML() bool {
	return o.rawOptions.TidyHTML
}

func (o Options) BaseTemplate() string {
	baseTemplate := "base.html"
	if o.rawOptions.BaseTemplate != "" {
		baseTemplate = o.rawOptions.BaseTemplate
	}
	return baseTemplate
}

func cleanJoin(head, tail string) string {
	return filepath.Clean(filepath.Join(head, tail))
}

func NewOptions() (*Options, *RawOptions) {
	o := &Options{&RawOptions{}}
	return o, o.rawOptions
}

func shouldProcessFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".html":
		return true
	}
	return false
}

func ReadConfig(siteDir string) (*RawOptions, error) {
	configPath := filepath.Join(siteDir, "yugo.jsonr")
	configOpts := new(RawOptions)
	raw, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err := jsonr.Unmarshal(raw, configOpts); err != nil {
		return nil, err
	}
	return configOpts, nil
}

func Run(opts *Options, args []string) {
	if len(args) > 0 {
		tmpl, err := loadTemplates(opts)
		if err != nil {
			log.Fatalf("template load failed: %s\n", err)
		}
		rel, err := filepath.Rel(opts.ContentDir(), args[0])
		if err != nil {
			rel = filepath.Base(args[0])
		}

		out, err := renderFile(args[0], rel, tmpl, opts, nil)
		if err != nil {
			log.Fatalf("failed: %s\n", err)
		}
		outPath := "/dev/stdout"
		if len(args) == 2 {
			outPath = args[1]
		}
		if err := os.WriteFile(outPath, []byte(out), 0644); err != nil {
			log.Fatalf("unable to write %s: %s\n", outPath, err)
		}
		return
	}

	fmt.Println("Building site...")

	if err := renderContent(opts); err != nil {
		log.Fatal(err)
	}

	if err := copyTree(opts.StaticDir(), opts.OutDir(), false); err != nil {
		log.Fatal(err)
	}

	if err := copyContent(opts.ContentDir(), opts.OutDir()); err != nil {
		log.Fatal(err)
	}

	// Copy internal resources last so that our core functionality always works.
	if err := CopyEmbeddedResources(opts.OutDir(), resources.RootFS); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Build complete.")
}

func loadTemplates(opts *Options) (*template.Template, error) {
	tl := TemplateLoader{
		TemplateDir: opts.TemplatesDir(),
		StaticDir:   opts.StaticDir(),
	}

	tmpl, err := tl.Load()
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func renderContent(opts *Options) error {
	sitePath := filepath.Join(opts.SiteDir(), "site.jsonr")
	siteConfig := map[string]any{}
	raw, err := os.ReadFile(sitePath)
	if err != nil {
		return nil
	}
	if err := jsonr.Unmarshal(raw, &siteConfig); err != nil {
		return nil
	}

	// Remove the output directory entirely to ensure clean output every time.
	if err := os.RemoveAll(opts.OutDir()); err != nil {
		return nil
	}
	if err := os.MkdirAll(opts.OutDir(), 0755); err != nil {
		return nil
	}

	tmpl, err := loadTemplates(opts)
	if err != nil {
		return fmt.Errorf("template load failed: %w", err)
	}

	return filepath.WalkDir(opts.ContentDir(), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk dir failed: %w", err)
		}
		if d.IsDir() {
			return nil
		}

		if !shouldProcessFile(path) {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		rel, _ := filepath.Rel(opts.ContentDir(), path)
		outPath := filepath.Join(opts.OutDir(), rel)
		if ext == ".md" {
			outPath = outPath[:len(outPath)-len(ext)] + ".html"
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("dir create failed: %w", err)
		}

		out, err := renderFile(path, rel, tmpl, opts, siteConfig)
		if err != nil {
			return err
		}

		if err := os.WriteFile(outPath, []byte(out), 0644); err != nil {
			return fmt.Errorf("unable to write %s: %w", outPath, err)
		}
		fmt.Println("→", outPath)
		return nil
	})
}

func renderFile(path string, relPath string, tmpl *template.Template, opts *Options, siteConfig map[string]any) (string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	page, err := ParsePage(src)
	if err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(path))

	htmlStr := ""
	tocItems := []TOCItem{}

	if ext == ".md" {
		htmlBuf := &bytes.Buffer{}
		tocExt := &TOCExtension{Items: &tocItems}

		md := goldmark.New(
			goldmark.WithExtensions(
				extension.GFM,
				extension.Typographer,
				tocExt,
			),
			goldmark.WithRendererOptions(
				html.WithUnsafe(), // This option allows raw HTML rendering
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
				parser.WithASTTransformers(
					util.Prioritized(
						LinkRewriter{
							SiteDir:    opts.SiteDir(),
							ContentDir: opts.ContentDir(),
						},
						100,
					),
				),
			),
		)

		pc := parser.NewContext()
		pc.Set(SourceFileKey, relPath)

		if err := md.Convert(page.Body, htmlBuf, parser.WithContext(pc)); err != nil {
			return "", fmt.Errorf("failed rendering markdown: %w", err)
		}
		htmlStr = htmlBuf.String()
	} else {
		htmlStr = string(page.Body)
	}

	page.Params["TOCItems"] = tocItems
	page.Params["TOC"] = template.HTML(GenerateTOC(tocItems))

	debugMap := map[string]any{
		"Page":       page.Params,
		"Site":       siteConfig,
		"LiveReload": opts.LiveReload(),
	}

	tmplData := map[string]any{
		"Content":  template.HTML(htmlStr),
		"DebugMap": debugMap,
	}
	maps.Copy(tmplData, debugMap)

	// Apply templates to both HTML and Markdown
	var tmplBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&tmplBuf, opts.BaseTemplate(), tmplData)
	if err != nil {
		return "", fmt.Errorf("failed rendering template: %w", err)
	}

	out := tmplBuf.String()

	if opts.TidyHTML() {
		out, err = htmltidy.NormalizeHTML(out)
		if err != nil {
			return "", fmt.Errorf("failed normalizing html %s: %w\n", path, err)
		}
	}

	return out, nil
}
