package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"gosd/sd"
)

const helpText = `Usage: gosd <subcommand> <script_name>
SD_ROOT: ~/.local/bin/scripts

Subcommands:

  run          Run a script
  help         Display help information
  new          Create new script
  edit         Edit existing script
  which        Display script path
  cat          Display script content
  cmd          Execute shell command from target directory
`

var Root = cli.Command{
	Name:                          "gosd",
	UsageText:                     "gosd <subcommand> <script_name>",
	CustomRootCommandHelpTemplate: helpText,
	EnableShellCompletion:         true,
	Commands: []*cli.Command{
		&run,
		&help,
		&newScript,
		&edit,
		&which,
		&cat,
		&cmdCommand,
	},
	Action:    rootAction,
	Reader:    os.Stdin,
	Writer:    os.Stdout,
	ErrWriter: os.Stderr,
}

func rootAction(ctx context.Context, command *cli.Command) error {
	if command.Args().Len() != 0 {
		return nil
	}

	runner := sd.Runner{
		Root:   os.Getenv("SD_ROOT"),
		Stdin:  command.Reader,
		Stdout: command.Writer,
		Stderr: command.ErrWriter,
	}
	root, err := runner.ScriptRoot()
	if err != nil {
		return err
	}
	runner.Root = root
	text := strings.Replace(helpText, "SD_ROOT: ~/.local/bin/scripts", "SD_ROOT: "+root, 1)
	if _, err := fmt.Fprint(command.Writer, text); err != nil {
		return fmt.Errorf("write root display: %w", err)
	}

	customDir := filepath.Join(root, "scriptura")
	info, err := os.Stat(customDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect user subcommands %q: %w", customDir, err)
	}
	if err == nil && info.IsDir() {
		entries, err := runner.Complete([]string{"scriptura"})
		if err != nil {
			return err
		}
		var custom []sd.Completion
		for _, entry := range entries {
			if entry.Directory {
				info, err := os.Stat(filepath.Join(customDir, entry.Name))
				if err != nil {
					return fmt.Errorf("inspect user subcommand %q: %w", entry.Name, err)
				}
				if info.Mode().Perm()&0o111 == 0 {
					continue
				}
			}
			custom = append(custom, entry)
		}
		if len(custom) != 0 {
			if _, err := fmt.Fprint(command.Writer, "\nUser Subcommands:\n\n"); err != nil {
				return fmt.Errorf("write user subcommands: %w", err)
			}
			for _, entry := range custom {
				if _, err := fmt.Fprintf(command.Writer, "  %-12s %s\n", entry.Name, entry.Description); err != nil {
					return fmt.Errorf("write user subcommand %q: %w", entry.Name, err)
				}
			}
		}
	}
	if _, err := fmt.Fprint(command.Writer, "\nScripts:\n\n"); err != nil {
		return fmt.Errorf("write root display: %w", err)
	}
	return runner.Run(ctx, nil)
}

func showHelp(_ context.Context, command *cli.Command) error {
	_, err := fmt.Fprint(command.Writer, helpText)
	return err
}
