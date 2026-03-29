package input

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Read detects input source: args first, then stdin pipe.
// When args are provided, they take priority (avoids blocking on stdin).
// Returns the text to translate, or an error if no input.
func Read(args []string) (string, error) {
	// Args take priority — avoids stdin blocking in non-interactive shells.
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	// No args — check stdin.
	info, err := os.Stdin.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat stdin: %w", err)
	}

	// ModeCharDevice means stdin is a terminal (no pipe).
	isPiped := info.Mode()&os.ModeCharDevice == 0
	if !isPiped {
		return "", fmt.Errorf("no input: pipe text or provide arguments (e.g., trans \"hello world\")")
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", fmt.Errorf("failed to read stdin: %w", err)
	}

	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", fmt.Errorf("stdin is empty")
	}
	return text, nil
}
