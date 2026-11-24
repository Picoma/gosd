package sd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

func SelectEditor() ([]string, error) {
	for _, name := range []string{"SD_EDITOR", "VISUAL", "EDITOR"} {
		if value, ok := os.LookupEnv(name); ok && strings.TrimSpace(value) != "" {
			parts, err := SplitCommand(value)
			if err != nil {
				return nil, fmt.Errorf("parse %s: %w", name, err)
			}
			return parts, nil
		}
	}

	for _, name := range []string{"vim", "nano", "vi"} {
		if path, err := exec.LookPath(name); err == nil {
			return []string{path}, nil
		}
	}
	return nil, fmt.Errorf("no editor configured; set SD_EDITOR, VISUAL, or EDITOR, or install vim, nano, or vi")
}

func SplitCommand(value string) ([]string, error) {
	var parts []string
	var current strings.Builder
	var quote rune
	escaped, started := false, false
	flush := func() {
		if started {
			parts = append(parts, current.String())
			current.Reset()
			started = false
		}
	}

	for _, char := range value {
		switch {
		case escaped:
			current.WriteRune(char)
			escaped = false
			started = true
		case char == '\\' && quote != '\'':
			escaped = true
			started = true
		case quote != 0:
			if char == quote {
				quote = 0
			} else {
				current.WriteRune(char)
			}
			started = true
		case char == '\'' || char == '"':
			quote = char
			started = true
		case unicode.IsSpace(char):
			flush()
		default:
			current.WriteRune(char)
			started = true
		}
	}

	if escaped {
		return nil, fmt.Errorf("trailing escape")
	}
	if quote != 0 {
		return nil, fmt.Errorf("unclosed quote")
	}
	flush()
	if len(parts) == 0 || parts[0] == "" {
		return nil, fmt.Errorf("empty command")
	}
	return parts, nil
}
