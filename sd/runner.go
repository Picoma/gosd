package sd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const defaultRoot = ".local/bin/scripts"

type Runner struct {
	Root   string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

type Completion struct {
	Name        string
	Description string
	Directory   bool
}

type Target struct {
	Root      string
	Path      string
	Info      os.FileInfo
	Remaining []string
}

func (r Runner) Run(ctx context.Context, args []string) error {
	root, err := r.scriptRoot()
	if err != nil {
		return err
	}

	rootInfo, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("script root %q: %w", root, err)
	}
	if !rootInfo.IsDir() {
		return fmt.Errorf("script root %q is not a directory", root)
	}

	target := root
	foundDirectory := false
	for i, arg := range args {
		if arg == "--" {
			return r.listDirectory(target)
		}

		if !filepath.IsLocal(arg) {
			return fmt.Errorf("script path %q must be relative to the script root", arg)
		}

		candidate := filepath.Join(target, arg)
		info, statErr := os.Stat(candidate)
		if statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) && foundDirectory {
				return r.listDirectory(target)
			}
			return fmt.Errorf("script %q: %w", candidate, statErr)
		}

		switch {
		case info.IsDir():
			target = candidate
			foundDirectory = true
		case info.Mode().IsRegular():
			scriptArgs := args[i+1:]
			if len(scriptArgs) > 0 && scriptArgs[0] == "--" {
				scriptArgs = scriptArgs[1:]
			}
			return r.runScript(ctx, root, candidate, scriptArgs)
		default:
			return fmt.Errorf("script %q is not a regular file or directory", candidate)
		}
	}

	if foundDirectory || len(args) == 0 {
		return r.listDirectory(target)
	}

	return fmt.Errorf("script %q not found under %q", args[0], root)
}

func (r Runner) ScriptRoot() (string, error) {
	root, err := r.scriptRoot()
	if err != nil {
		return "", err
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("script root %q: %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("script root %q is not a directory", root)
	}
	return root, nil
}

func (r Runner) Resolve(args []string) (Target, error) {
	root, err := r.ScriptRoot()
	if err != nil {
		return Target{}, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return Target{}, fmt.Errorf("inspect script root %q: %w", root, err)
	}
	target := Target{Root: root, Path: root, Info: info}
	foundDirectory := false
	for i, arg := range args {
		if arg == "--" {
			target.Remaining = args[i+1:]
			return target, nil
		}
		if !filepath.IsLocal(arg) {
			return Target{}, fmt.Errorf("script path %q must be relative to the script root", arg)
		}

		candidate := filepath.Join(target.Path, arg)
		info, statErr := os.Stat(candidate)
		if statErr != nil {
			if errors.Is(statErr, os.ErrNotExist) && foundDirectory {
				target.Remaining = args[i:]
				return target, nil
			}
			return Target{}, fmt.Errorf("script %q: %w", candidate, statErr)
		}
		if info.IsDir() {
			target.Path = candidate
			target.Info = info
			foundDirectory = true
			continue
		}
		if info.Mode().IsRegular() {
			target.Path = candidate
			target.Info = info
			target.Remaining = args[i+1:]
			return target, nil
		}
		return Target{}, fmt.Errorf("script %q is not a regular file or directory", candidate)
	}
	return target, nil
}

func (r Runner) Complete(args []string) ([]Completion, error) {
	root, err := r.scriptRoot()
	if err != nil {
		return nil, err
	}
	rootInfo, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("script root %q: %w", root, err)
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("script root %q is not a directory", root)
	}

	target := root
	for _, arg := range args {
		if arg == "--" || !filepath.IsLocal(arg) {
			return nil, nil
		}
		candidate := filepath.Join(target, arg)
		info, err := os.Stat(candidate)
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		if err != nil {
			return nil, fmt.Errorf("inspect completion path %q: %w", candidate, err)
		}
		if info.IsDir() {
			target = candidate
			continue
		}
		if info.Mode().IsRegular() {
			return nil, nil
		}
		return nil, nil
	}

	return scriptEntries(target)
}

func (r Runner) scriptRoot() (string, error) {
	root := r.Root
	if root == "" {
		root = os.Getenv("SD_ROOT")
	}
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("find home directory: %w", err)
		}
		root = filepath.Join(home, defaultRoot)
	} else if root == "~" || strings.HasPrefix(root, "~"+string(filepath.Separator)) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("expand script root: %w", err)
		}
		root = filepath.Join(home, strings.TrimPrefix(root, "~"+string(filepath.Separator)))
	}

	root, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve script root: %w", err)
	}
	return filepath.Clean(root), nil
}

func (r Runner) runScript(ctx context.Context, root, target string, args []string) error {
	info, err := os.Stat(target)
	if err != nil {
		return fmt.Errorf("inspect script %q: %w", target, err)
	}
	if info.Mode().Perm()&0o111 == 0 {
		if _, err := fmt.Fprintf(r.stderr(), "gosd: %s is not executable\n", target); err != nil {
			return fmt.Errorf("write script error: %w", err)
		}
		if err := r.View(ctx, target); err != nil {
			return fmt.Errorf("display script %q: %w", target, err)
		}
		return nil
	}

	command := exec.CommandContext(ctx, target, args...)
	command.Stdin = r.stdin()
	command.Stdout = r.stdout()
	command.Stderr = r.stderr()
	command.Env = setEnvironment(os.Environ(), "SD", filepath.Dir(target))
	command.Env = setEnvironment(command.Env, "SD_ROOT", root)
	if err := command.Run(); err != nil {
		return fmt.Errorf("run script %q: %w", target, err)
	}
	return nil
}

func (r Runner) listDirectory(directory string) error {
	entries, err := scriptEntries(directory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name
		if entry.Directory {
			name += " ..."
		}
		if _, err := fmt.Fprintf(r.stdout(), "%-12s -- %-64s\n", name, entry.Description); err != nil {
			return fmt.Errorf("write script listing: %w", err)
		}
	}
	return nil
}

func scriptEntries(directory string) ([]Completion, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, fmt.Errorf("list scripts in %q: %w", directory, err)
	}

	var scripts []Completion
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "scriptura" || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		description, err := scriptDescription(filepath.Join(directory, entry.Name()), true)
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, Completion{Name: entry.Name(), Description: description, Directory: true})
	}

	for _, entry := range entries {
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("inspect script %q: %w", entry.Name(), err)
		}
		if info.Mode().IsRegular() && info.Mode().Perm()&0o111 != 0 {
			description, err := scriptDescription(filepath.Join(directory, entry.Name()), false)
			if err != nil {
				return nil, err
			}
			scripts = append(scripts, Completion{Name: entry.Name(), Description: description})
		}
	}
	return scripts, nil
}

func scriptDescription(path string, isDirectory bool) (string, error) {
	helpPath := path + ".help"
	if isDirectory {
		helpPath = filepath.Join(path, "help")
	}
	if content, err := os.ReadFile(helpPath); err == nil {
		return strings.TrimRight(string(content), "\r\n"), nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("read help file %q: %w", helpPath, err)
	}
	if isDirectory {
		return filepath.Base(path) + " commands", nil
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read script description %q: %w", path, err)
	}
	var comments []string
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}
		comment := strings.TrimSpace(strings.TrimPrefix(line, "#"))
		if comment != "" && !strings.HasPrefix(comment, "!") {
			comments = append(comments, comment)
		}
	}
	return strings.Join(comments, "\n"), nil
}

func (r Runner) stdin() io.Reader {
	if r.Stdin != nil {
		return r.Stdin
	}
	return os.Stdin
}

func (r Runner) stdout() io.Writer {
	if r.Stdout != nil {
		return r.Stdout
	}
	return os.Stdout
}

func (r Runner) stderr() io.Writer {
	if r.Stderr != nil {
		return r.Stderr
	}
	return os.Stderr
}

func setEnvironment(environment []string, key, value string) []string {
	prefix := key + "="
	result := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		if !strings.HasPrefix(entry, prefix) {
			result = append(result, entry)
		}
	}
	return append(result, prefix+value)
}
