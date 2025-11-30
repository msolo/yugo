package initcmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/msolo/yugo/internal/build"
	"github.com/msolo/yugo/internal/resources"
)

type Options struct {
	InitDir string
}

func Run(opts Options) {
	if opts.InitDir == "" {
		log.Fatal("no init dir specified")
	}

	if err := os.MkdirAll(opts.InitDir, 0775); err != nil {
		log.Fatalf("error: %s", err)
	}

	if _, err := os.Stat(filepath.Join(opts.InitDir, "yugo.jsonr")); err == nil {
		log.Fatalf("error: yugo.jsonr config already found in %s", opts.InitDir)
	}

	if err := build.CopyEmbeddedResources(opts.InitDir, resources.ExampleFS); err != nil {
		log.Fatal(err)
	}
}
