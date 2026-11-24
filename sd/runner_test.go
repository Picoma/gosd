package sd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunNestedScriptPassesArgumentsAndEnvironment(t *testing.T) {
	root := t.TempDir()
	scriptDir := filepath.Join(root, "demo")
	if err := os.Mkdir(scriptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(scriptDir, "args")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf '%s\\n' \"$SD\" \"$SD_ROOT\" \"$@\"\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	runner := Runner{Root: root, Stdout: &stdout, Stderr: &stderr}
	if err := runner.Run(context.Background(), []string{"demo", "args", "one", "two words"}); err != nil {
		t.Fatal(err)
	}

	want := scriptDir + "\n" + root + "\none\ntwo words\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}
}

func TestRunReturnsScriptExitError(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, "fail")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 7\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	err := (Runner{Root: root}).Run(context.Background(), []string{"fail"})
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("error = %v, want *exec.ExitError", err)
	}
	if got := exitErr.ExitCode(); got != 7 {
		t.Fatalf("exit code = %d, want 7", got)
	}
}

func TestRunDisplaysNonExecutableFile(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, "not-executable")
	const content = "#!/bin/sh\necho contents\n"
	if err := os.WriteFile(script, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	runner := Runner{Root: root, Stdout: &stdout, Stderr: &stderr}
	if err := runner.Run(context.Background(), []string{"not-executable"}); err != nil {
		t.Fatal(err)
	}
	if got := stdout.String(); got != content {
		t.Fatalf("stdout = %q, want %q", got, content)
	}
	if got := stderr.String(); got == "" {
		t.Fatal("expected a not-executable diagnostic")
	}
}

func TestRunListsRootWhenNoScriptIsGiven(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "hello"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "hidden"), []byte("not executable\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	runner := Runner{Root: root, Stdout: &stdout}
	if err := runner.Run(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("%-12s -- %-64s\n%-12s -- %-64s\n", "demo ...", "demo commands", "hello", "")
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunDirectoryListingUsesHelpAndScriptComments(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "demo")
	if err := os.Mkdir(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "help"), []byte("Demo commands\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(root, "hello")
	if err := os.WriteFile(script, []byte("#!/bin/sh\n# Say hello.\necho hello\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if err := (Runner{Root: root, Stdout: &stdout}).Run(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("%-12s -- %-64s\n%-12s -- %-64s\n", "demo ...", "Demo commands", "hello", "Say hello.")
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}
