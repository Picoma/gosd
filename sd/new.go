package sd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func (r Runner) CreateRoot() error {
	root, err := r.scriptRoot()
	if err != nil {
		return err
	}
	if _, err := os.Stat(root); err == nil {
		return fmt.Errorf("script target is required")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect script root %q: %w", root, err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("create script root %q: %w", root, err)
	}
	return nil
}

func (r Runner) Create(ctx context.Context, name string, content []string, edit, fish bool) error {
	if name == "" || !filepath.IsLocal(name) || name == "." || strings.HasSuffix(name, string(filepath.Separator)) {
		return fmt.Errorf("script path %q must be a relative file path", name)
	}

	root, err := r.scriptRoot()
	if err != nil {
		return err
	}
	target := filepath.Join(root, name)
	if _, err := os.Lstat(target); err == nil {
		return fmt.Errorf("script %q already exists", name)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect script %q: %w", target, err)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create script directory for %q: %w", name, err)
	}
	file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("script %q already exists", name)
		}
		return fmt.Errorf("create script %q: %w", name, err)
	}

	script := "#!/usr/bin/env bash\n\n"
	if fish {
		script = "#!/usr/bin/env fish\n\n"
	}
	if len(content) > 0 {
		script += strings.Join(content, " ")
		if !strings.HasSuffix(script, "\n") {
			script += "\n"
		}
	}
	_, writeErr := file.WriteString(script)
	closeErr := file.Close()
	if writeErr != nil {
		_ = os.Remove(target)
		return fmt.Errorf("write script %q: %w", name, writeErr)
	}
	if closeErr != nil {
		_ = os.Remove(target)
		return fmt.Errorf("close script %q: %w", name, closeErr)
	}
	if err := os.Chmod(target, 0o755); err != nil {
		return fmt.Errorf("make script %q executable: %w", name, err)
	}

	if len(content) == 0 || edit {
		return r.edit(ctx, target)
	}
	return nil
}

func (r Runner) CompleteNew(args []string) ([]Completion, error) {
	name := filepath.Join(args...)
	if name != "" && !filepath.IsLocal(name) {
		return nil, nil
	}

	directory, prefix := filepath.Split(name)
	directory = filepath.Clean(directory)
	if name == "" {
		directory = "."
	}
	var parent []string
	if directory != "." {
		parent = strings.Split(directory, string(filepath.Separator))
	}

	entries, err := r.Complete(parent)
	if err != nil {
		return nil, err
	}
	var result []Completion
	for _, entry := range entries {
		if entry.Directory && strings.HasPrefix(entry.Name, prefix) {
			entry.Name = filepath.Join(directory, entry.Name)
			entry.Name += string(filepath.Separator)
			result = append(result, entry)
		}
	}
	return result, nil
}

func (r Runner) edit(ctx context.Context, target string) error {
	parts, err := SelectEditor()
	if err != nil {
		return err
	}
	command := exec.CommandContext(ctx, parts[0], append(parts[1:], target)...)
	command.Stdin = r.stdin()
	command.Stdout = r.stdout()
	command.Stderr = r.stderr()
	if err := command.Run(); err != nil {
		return fmt.Errorf("run editor %q: %w", parts[0], err)
	}
	return nil
}
