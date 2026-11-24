package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"gosd/completions"
	"gosd/sd"
)

var which = cli.Command{
	Name:          "which",
	Usage:         "Display script path",
	ArgsUsage:     "<script_name>",
	Action:        whichAction,
	ShellComplete: completions.Which,
}

func whichAction(_ context.Context, command *cli.Command) error {
	args := command.Args().Slice()
	if len(args) == 0 {
		return fmt.Errorf("script target is required")
	}

	target, err := (sd.Runner{Root: os.Getenv("SD_ROOT")}).Resolve(args)
	if err != nil {
		return err
	}
	if len(target.Remaining) > 0 {
		if target.Info.IsDir() {
			return fmt.Errorf("script target %q not found under %q", strings.Join(target.Remaining, string(filepath.Separator)), target.Path)
		}
		return fmt.Errorf("unexpected arguments after script target: %s", strings.Join(target.Remaining, " "))
	}
	if !target.Info.Mode().IsRegular() {
		return fmt.Errorf("script target %q is not a regular file", target.Path)
	}
	if target.Info.Mode().Perm()&0o111 == 0 {
		return fmt.Errorf("script target %q is not executable", target.Path)
	}

	path, err := filepath.Abs(target.Path)
	if err != nil {
		return fmt.Errorf("resolve script path %q: %w", target.Path, err)
	}
	if _, err := fmt.Fprintln(command.Writer, path); err != nil {
		return fmt.Errorf("write script path: %w", err)
	}
	return nil
}
