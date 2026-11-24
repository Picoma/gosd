package sd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func (r Runner) View(ctx context.Context, path string) error {
	viewer := []string{"cat"}
	if configured := os.Getenv("SD_CAT"); configured != "" {
		var err error
		viewer, err = SplitCommand(configured)
		if err != nil {
			return fmt.Errorf("parse SD_CAT: %w", err)
		}
	}

	command := exec.CommandContext(ctx, viewer[0], append(viewer[1:], path)...)
	command.Stdin = r.stdin()
	command.Stdout = r.stdout()
	command.Stderr = r.stderr()
	return command.Run()
}
