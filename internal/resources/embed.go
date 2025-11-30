package resources

import (
	"embed"
	"io/fs"
	"log"
)

// Embed all of ./root in the binary to be copied into public on build.
// Anything here will eventually override user content, so we rely on
// using namespaces to keep things from conflicting as much as possible.

//go:embed all:root all:example
var EmbeddedResources embed.FS

// Turn it into a sub-FS rooted at `root/`
var RootFS fs.FS

// Turn it into a sub-FS rooted at `example/`
var ExampleFS fs.FS

func init() {
	sub, err := fs.Sub(EmbeddedResources, "root")
	if err != nil {
		log.Fatalf("unable to initialize embedded resources: %v", err)
	}
	RootFS = sub

	sub, err = fs.Sub(EmbeddedResources, "example")
	if err != nil {
		log.Fatalf("unable to initialize embedded resources: %v", err)
	}
	ExampleFS = sub
}
