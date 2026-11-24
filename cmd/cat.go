package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"gosd/completions"
	"gosd/sd"
)

var cat = cli.Command{
	Name:          "cat",
	Usage:         "Display a script",
	ArgsUsage:     "<script_path>",
	Action:        catAction,
	ShellComplete: completions.Cat,
}

func catAction(ctx context.Context, command *cli.Command) error {
	runner := sd.Runner{
		Root:   os.Getenv("SD_ROOT"),
		Stdin:  command.Reader,
		Stdout: command.Writer,
		Stderr: command.ErrWriter,
	}
	target, err := runner.Resolve(command.Args().Slice())
	if err != nil {
		return err
	}
	if len(target.Remaining) > 0 {
		return fmt.Errorf("unexpected or unresolved target arguments: %s", strings.Join(target.Remaining, " "))
	}
	if !target.Info.Mode().IsRegular() {
		return fmt.Errorf("script path %q is not a regular file", target.Path)
	}

	if err := runner.View(ctx, target.Path); err != nil {
		return fmt.Errorf("display script %q: %w", target.Path, err)
	}
	return nil
}
