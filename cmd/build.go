package cmd

import (
	"context"
	"log"

	"github.com/msolo/cmdflag"
	"github.com/msolo/yugo/internal/build"
)

var cmdBuild = &cmdflag.Command{
	Name:      "build",
	Run:       runBuild,
	UsageLine: "yugo build [flags]",
	UsageLong: "Build a yugo site into the public output directory.",
	Flags: []cmdflag.Flag{
		{Name: "tidy-html", FlagType: cmdflag.FlagTypeBool, DefaultValue: true, Usage: "Tidy HTML - normalize and pretty-print"},
		{Name: "site", FlagType: cmdflag.FlagTypeString, DefaultValue: ".", Usage: "Path to site directory (default: current directory)", Predictor: cmdflag.PredictDirs("*")},
		{Name: "outdir", FlagType: cmdflag.FlagTypeString, DefaultValue: "", Usage: "Path to out directory (default: ./public)", Predictor: cmdflag.PredictDirs("*")},
	},
	// We have no positional args
	Args: cmdflag.PredictNothing,
}

func runBuild(ctx context.Context, cmd *cmdflag.Command, args []string) {
	opts, ropts := build.NewOptions()

	// FIXME: It would be interesting to do this with reflection, much like
	// the json module.
	fs := cmd.BindFlagSet(map[string]any{
		"tidy-html": &ropts.TidyHTML,
		"site":      &ropts.SiteDir,
		"outdir":    &ropts.OutDir,
	})
	_ = fs.Parse(args)
	if err := opts.MergeConfig(); err != nil {
		log.Fatal(err)
	}

	build.Run(opts)
}
