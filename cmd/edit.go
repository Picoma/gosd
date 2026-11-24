package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/urfave/cli/v3"

	"gosd/completions"
	"gosd/sd"
)

var edit = cli.Command{
	Name:          "edit",
	Usage:         "Edit an existing script",
	ArgsUsage:     "<script_name>",
	Action:        editAction,
	ShellComplete: completions.Edit,
}

func editAction(ctx context.Context, command *cli.Command) error {
	runner := sd.Runner{Root: os.Getenv("SD_ROOT")}
	target, err := runner.Resolve(command.Args().Slice())
	if err != nil {
		return err
	}
	if !target.Info.Mode().IsRegular() {
		return fmt.Errorf("script path %q is not an existing script file", target.Path)
	}
	if len(target.Remaining) > 0 {
		return fmt.Errorf("unexpected arguments after script path: %s", strings.Join(target.Remaining, " "))
	}

	editor, err := sd.SelectEditor()
	if err != nil {
		return err
	}
	args := append(append([]string(nil), editor[1:]...), target.Path)
	process := exec.CommandContext(ctx, editor[0], args...)
	process.Stdin = command.Reader
	process.Stdout = command.Writer
	process.Stderr = command.ErrWriter
	if err := process.Run(); err != nil {
		return fmt.Errorf("run editor %q for script %q: %w", editor[0], target.Path, err)
	}
	return nil
}
