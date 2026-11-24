package completions

import (
	"context"
	"os"

	"github.com/urfave/cli/v3"

	"gosd/sd"
)

func Help(_ context.Context, command *cli.Command) {
	runner := sd.Runner{Root: os.Getenv("SD_ROOT")}
	candidates, err := runner.Complete(command.Args().Slice())
	if err != nil {
		reportError(command, "complete script path", err)
		return
	}
	writeCandidates(command, candidates)
}
