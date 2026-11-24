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

var newScript = cli.Command{
	Name:            "new",
	Usage:           "Create new script",
	ArgsUsage:       "[--fish] <script_name> [-- script_content]",
	SkipFlagParsing: true,
	Action:          newScriptAction,
	ShellComplete:   completions.New,
}

func newScriptAction(ctx context.Context, command *cli.Command) error {
	var names []string
	var content []string
	var edit bool
	var fish bool
	afterSeparator := false
	for _, arg := range command.Args().Slice() {
		if afterSeparator {
			content = append(content, arg)
			continue
		}
		if arg == "--" {
			afterSeparator = true
			continue
		}
		if arg == "--edit" || arg == "-e" {
			edit = true
			continue
		}
		if arg == "--fish" {
			fish = true
			continue
		}
		names = append(names, arg)
	}
	if len(content) > 0 && strings.Join(content, "") == "" {
		content = nil
	}

	runner := sd.Runner{
		Root:   os.Getenv("SD_ROOT"),
		Stdin:  command.Reader,
		Stdout: command.Writer,
		Stderr: command.ErrWriter,
	}
	if len(names) == 0 {
		if len(content) != 0 || edit {
			return fmt.Errorf("script target is required")
		}
		return runner.CreateRoot()
	}
	return runner.Create(ctx, filepath.Join(names...), content, edit, fish)
}
