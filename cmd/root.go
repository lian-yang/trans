package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/lian-yang/trans/internal/config"
	"github.com/lian-yang/trans/internal/input"
	"github.com/lian-yang/trans/internal/llm"
	"github.com/lian-yang/trans/internal/output"
	"github.com/spf13/cobra"
)

var (
	flagTo       string
	flagModel    string
	flagVerbose  bool
	flagStream   bool
	flagNoStream bool
	flagContrast bool
	flagVersion  bool
)

// version is resolved via three-tier fallback at init time:
//  1. -ldflags injection (highest priority)
//  2. runtime/debug build info (vcs.tag from go install)
//  3. "unknown" (lowest priority)
var version = ""

func init() {
	resolveVersion()

	rootCmd.Flags().BoolVarP(&flagVersion, "version", "V", false, "print version and exit")
	rootCmd.Flags().StringVarP(&flagTo, "to", "t", "", "target language (default: zh)")
	rootCmd.Flags().StringVarP(&flagModel, "model", "m", "", "model to use (default: gpt-4o-mini)")
	rootCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "show source language annotation")
	rootCmd.Flags().BoolVarP(&flagStream, "stream", "s", false, "force streaming output (default: auto-detect by TTY)")
	rootCmd.Flags().BoolVar(&flagNoStream, "no-stream", false, "force batch output (disable streaming)")
	rootCmd.Flags().BoolVarP(&flagContrast, "contrast", "c", false, "contrast mode: show original and translated lines alternately")
}

func resolveVersion() {
	if version != "" {
		return // already set via ldflags
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		for _, s := range bi.Settings {
			if s.Key == "vcs.tag" && s.Value != "" {
				version = s.Value
				return
			}
		}
		if bi.Main.Version != "" {
			version = bi.Main.Version
			return
		}
	}
	version = "unknown"
}

var rootCmd = &cobra.Command{
	Use:   "trans [text]",
	Short: "Terminal AI translation tool powered by OpenAI",
	Long: `Translate text using OpenAI-compatible APIs.

Supports pipe and argument input:
  echo "hello world" | trans
  trans "hello world"
  cat README.md | trans -t ja`,
	Args:              cobra.MinimumNArgs(0),
	RunE:              run,
	DisableAutoGenTag: true,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	// Handle --version before any I/O or config loading.
	if flagVersion {
		fmt.Println(version)
		return nil
	}

	// 1. Load config: file → env → defaults.
	cfg, err := config.Load()
	if err != nil {
		output.WriteErr("failed to load config: %v", err)
		return err
	}

	// 2. Read input first (fail fast if no input, before API key check).
	text, err := input.Read(args)
	if err != nil {
		output.WriteErr("%v", err)
		return err
	}

	// 3. Apply CLI flag overrides.
	cfg.SetModel(flagModel)
	cfg.SetTargetLang(flagTo)
	cfg.SetVerbose(flagVerbose)

	// 4. Validate config (after confirming there's input to process).
	if err := cfg.Validate(); err != nil {
		output.WriteErr("%v", err)
		return err
	}

	// 5. Detect source language when verbose.
	var srcLang string
	client := llm.NewClient(cfg.APIKey, cfg.BaseURL, cfg.Model)
	if cfg.Verbose {
		srcLang, _ = client.DetectLanguage(text)
	}

	// 6. Determine output mode: --no-stream wins, then -s, then auto-detect.
	useStream := !flagNoStream && (flagStream || output.IsTerminal())

	// 7. Call OpenAI.
	if useStream {
		// Stream to terminal.
		if cfg.Verbose && srcLang != "" {
			fmt.Fprintf(os.Stdout, "[%s→%s] ", srcLang, cfg.TargetLang)
		}
		err = client.TranslateStream(text, cfg.TargetLang, flagContrast, func(chunk string) {
			fmt.Print(chunk)
		})
		if err != nil {
			output.WriteErr("%v", err)
			return err
		}
		fmt.Println() // trailing newline
	} else {
		// Batch for pipe.
		result, err := client.Translate(text, cfg.TargetLang, flagContrast)
		if err != nil {
			output.WriteErr("%v", err)
			return err
		}
		w := output.NewWriter(cfg.Verbose)
		w.WriteVerbose(srcLang, cfg.TargetLang, result)
	}

	return nil
}
