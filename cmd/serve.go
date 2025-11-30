package cmd

import (
	"context"
	"log"

	"github.com/msolo/cmdflag"
	"github.com/msolo/yugo/internal/build"
	"github.com/msolo/yugo/internal/serve"
)

var cmdServe = &cmdflag.Command{
	Name:      "serve",
	Run:       runServe,
	UsageLine: "yugo serve [flags]",
	UsageLong: "Serve a yugo site with live reload and optional HTML tidying.",
	Flags: append([]cmdflag.Flag{
		{Name: "host", FlagType: cmdflag.FlagTypeString, DefaultValue: "127.0.0.1", Usage: "Host to bind HTTP server"},
		{Name: "port", FlagType: cmdflag.FlagTypeInt, DefaultValue: 8817, Usage: "Port for HTTP server"},
		{Name: "live-reload", FlagType: cmdflag.FlagTypeBool, DefaultValue: true, Usage: "Control live reload (default: enabled)"},
	}, cmdBuild.Flags...),
	// We have no positional args
	Args: cmdflag.PredictNothing,
}

func runServe(ctx context.Context, cmd *cmdflag.Command, args []string) {
	opts, ropts := build.NewOptions()

	fs := cmd.BindFlagSet(map[string]any{
		"host":        &ropts.Host,
		"port":        &ropts.Port,
		"live-reload": &ropts.LiveReload,
		"tidy-html":   &ropts.TidyHTML,
		"site":        &ropts.SiteDir,
		"outdir":      &ropts.OutDir,
	})

	_ = fs.Parse(args)
	if err := opts.MergeConfig(); err != nil {
		log.Fatal(err)
	}

	serve.Run(opts)
}
