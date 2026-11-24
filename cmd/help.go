package cmd

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"gosd/completions"
	"gosd/sd"
)

var help = cli.Command{
	Name:            "help",
	Usage:           "Display help information",
	ArgsUsage:       "<script_name>",
	SkipFlagParsing: true,
	Action:          helpAction,
	ShellComplete:   completions.Help,
}

func helpAction(ctx context.Context, command *cli.Command) error {
	args := command.Args().Slice()
	if len(args) == 0 {
		return showHelp(ctx, command)
	}

	runner := sd.Runner{
		Root:   os.Getenv("SD_ROOT"),
		Stdin:  command.Reader,
		Stdout: command.Writer,
		Stderr: command.ErrWriter,
	}
	return runner.HelpContext(ctx, args)
}
