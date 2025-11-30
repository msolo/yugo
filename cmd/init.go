package cmd

import (
	"context"
	"flag"

	"github.com/msolo/cmdflag"
	"github.com/msolo/yugo/internal/initcmd"
)

var cmdInit = &cmdflag.Command{
	Name:      "init",
	Run:       runInit,
	UsageLine: "yugo init [directory]",
	UsageLong: "Initialize a new yugo site in the specified directory (default: current directory).",
	Args:      cmdflag.PredictDirs("*"),
}

func runInit(ctx context.Context, _ *cmdflag.Command, args []string) {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	_ = fs.Parse(args)

	opts := initcmd.Options{
		InitDir: fs.Arg(0),
	}

	initcmd.Run(opts)
}
