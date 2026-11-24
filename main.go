package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"gosd/cmd"
)

func main() {
	if err := cmd.Root.Run(context.Background(), os.Args); err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			os.Exit(exitErr.ExitCode())
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
