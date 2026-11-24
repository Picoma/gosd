package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"gosd/completions"
	"gosd/sd"
)

var cmdCommand = cli.Command{
	Name:            "cmd",
	Usage:           "Execute shell command from target directory",
	ArgsUsage:       "[target] -- <command> [arguments...]",
	SkipFlagParsing: true,
	Action:          cmdAction,
	ShellComplete:   completions.Cmd,
}

func cmdAction(ctx context.Context, command *cli.Command) error {
	args := command.Args().Slice()
	separator := -1
	for i, arg := range args {
		if arg == "--" {
			separator = i
			break
		}
	}
	if separator == -1 {
		return fmt.Errorf("cmd requires -- before the command")
	}
	commandArgs := args[separator+1:]
	if len(commandArgs) == 0 {
		return fmt.Errorf("cmd requires a command after --")
	}

	runner := sd.Runner{
		Root:   os.Getenv("SD_ROOT"),
		Stdin:  command.Reader,
		Stdout: command.Writer,
		Stderr: command.ErrWriter,
	}
	target, err := runner.Resolve(args[:separator])
	if err != nil {
		return err
	}
	if len(target.Remaining) > 0 {
		return fmt.Errorf("target path %q not found under %q", strings.Join(target.Remaining, string(filepath.Separator)), target.Path)
	}

	directory := target.Path
	if !target.Info.IsDir() {
		directory = target.Root
		commandArgs = append(commandArgs, target.Path)
	}

	external := exec.CommandContext(ctx, commandArgs[0], commandArgs[1:]...)
	external.Dir = directory
	external.Stdin = command.Reader
	external.Stdout = command.Writer
	external.Stderr = command.ErrWriter
	return external.Run()
}
