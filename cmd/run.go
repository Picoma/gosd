package cmd

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"gosd/completions"
	"gosd/sd"
)

var run = cli.Command{
	Name:            "run",
	Usage:           "Run a script",
	ArgsUsage:       "<script_name> [arguments...]",
	SkipFlagParsing: true,
	Action:          runAction,
	ShellComplete:   completions.Run,
}

func runAction(ctx context.Context, command *cli.Command) error {
	runner := sd.Runner{
		Root:   os.Getenv("SD_ROOT"),
		Stdin:  command.Reader,
		Stdout: command.Writer,
		Stderr: command.ErrWriter,
	}
	return runner.Run(ctx, command.Args().Slice())
}
