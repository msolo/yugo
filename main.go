package main

import (
	"context"

	"github.com/msolo/cmdflag"
	"github.com/msolo/yugo/cmd"
)

func main() {
	root, subcommands := cmd.Commands()

	ctx := context.Background()
	command, args := cmdflag.Parse(root, subcommands)

	command.Run(ctx, command, args)
}
