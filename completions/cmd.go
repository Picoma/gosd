package completions

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"gosd/sd"
)

// Cmd provides completion for the cmd command's target path.
func Cmd(_ context.Context, command *cli.Command) {
	runner := sd.Runner{Root: os.Getenv("SD_ROOT")}
	candidates, err := runner.Complete(command.Args().Slice())
	if err != nil {
		reportError(command, "complete target path", err)
		return
	}
	writeCandidates(command, candidates)
}
