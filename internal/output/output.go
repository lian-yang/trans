package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// IsTerminal returns true if stdout is connected to a terminal.
func IsTerminal() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// Writer wraps output formatting logic.
type Writer struct {
	verbose bool
	dst     io.Writer
}

// NewWriter creates an output writer.
// verbose: show [src→tgt] prefix.
func NewWriter(verbose bool) *Writer {
	return &Writer{verbose: verbose, dst: os.Stdout}
}

// Write outputs translated text.
// If verbose, prepends the language annotation.
func (w *Writer) Write(text string) {
	text = strings.TrimSpace(text)
	if w.verbose {
		// Verbose format: raw text, user sees the translation directly.
		// Language annotation is handled at a higher level if needed.
		fmt.Fprintln(w.dst, text)
		return
	}
	fmt.Fprintln(w.dst, text)
}

// WriteVerbose outputs with language direction annotation.
func (w *Writer) WriteVerbose(srcLang, tgtLang, text string) {
	text = strings.TrimSpace(text)
	if w.verbose && srcLang != "" {
		fmt.Fprintf(w.dst, "[%s→%s] %s\n", srcLang, tgtLang, text)
		return
	}
	fmt.Fprintln(w.dst, text)
}

// WriteErr writes error message to stderr.
func WriteErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
}
