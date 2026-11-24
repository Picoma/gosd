package sd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var helpComment = regexp.MustCompile(`^#\s*([^!].*)$`)

func (r Runner) Help(args []string) error {
	return r.HelpContext(context.Background(), args)
}

func (r Runner) HelpContext(ctx context.Context, args []string) error {
	target, err := r.Resolve(args)
	if err != nil {
		return err
	}
	if len(target.Remaining) > 0 {
		return fmt.Errorf("unexpected or unresolved target arguments: %s", strings.Join(target.Remaining, " "))
	}

	helpPath := target.Path + ".help"
	if target.Info.IsDir() {
		helpPath = filepath.Join(target.Path, "help")
	}
	info, err := os.Stat(helpPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect help file %q: %w", helpPath, err)
	}
	if err == nil && info.Mode().IsRegular() {
		if err := r.View(ctx, helpPath); err != nil {
			return fmt.Errorf("display help file %q: %w", helpPath, err)
		}
		return nil
	}
	if target.Info.IsDir() {
		_, err = fmt.Fprintln(r.stdout(), filepath.Base(target.Path)+" commands")
		return err
	}

	content, err := os.ReadFile(target.Path)
	if err != nil {
		return fmt.Errorf("read script %q: %w", target.Path, err)
	}
	var comments []string
	for _, line := range strings.Split(string(content), "\n") {
		if match := helpComment.FindStringSubmatch(line); match != nil {
			comments = append(comments, match[1])
		}
	}
	_, err = fmt.Fprintln(r.stdout(), strings.Join(comments, "\n"))
	return err
}
