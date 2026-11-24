package completions

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"gosd/sd"
)

// New provides path completion for the new command.
func New(_ context.Context, command *cli.Command) {
	args := command.Args().Slice()
	var target []string
	for _, arg := range args {
		if arg == "--" {
			return
		}
		if arg == "-e" || arg == "--edit" || arg == "--fish" {
			continue
		}
		target = append(target, arg)
	}

	runner := sd.Runner{Root: os.Getenv("SD_ROOT")}
	candidates, err := runner.CompleteNew(target)
	if err != nil {
		reportError(command, "complete new script path", err)
		return
	}
	writeCandidates(command, candidates)
}
