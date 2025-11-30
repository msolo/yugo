package cmd

import (
	"github.com/msolo/cmdflag"
)

var cmdMain = &cmdflag.Command{
	Name:      "yugo",
	UsageLine: "yugo [command]",
	UsageLong: `yugo - a simple static site generator.

Commands:
  init    Initialize a new site
  build   Build a site
  serve   Serve a site with live reload`,
}

var subcommands = []*cmdflag.Command{
	cmdInit,
	cmdBuild,
	cmdServe,
}

// Commands returns the root command and all subcommands for use by main.
func Commands() (*cmdflag.Command, []*cmdflag.Command) {
	return cmdMain, subcommands
}
