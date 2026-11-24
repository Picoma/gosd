package completions

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"gosd/sd"
)

// Which provides dynamic completion for the which command.
func Which(_ context.Context, command *cli.Command) {
	runner := sd.Runner{Root: os.Getenv("SD_ROOT")}
	completions, err := runner.Complete(command.Args().Slice())
	if err != nil {
		reportError(command, "complete script path", err)
		return
	}
	writeCandidates(command, completions)
}
