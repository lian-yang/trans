package cmd

import (
	"fmt"
	"os"

	"github.com/lian-yang/trans/internal/config"
	"github.com/lian-yang/trans/internal/input"
	"github.com/lian-yang/trans/internal/llm"
	"github.com/lian-yang/trans/internal/output"
	"github.com/spf13/cobra"
)

var (
	flagTo    string
	flagModel string
	flagVerbose bool
)

var rootCmd = &cobra.Command{
	Use:   "trans [text]",
	Short: "Terminal AI translation tool powered by OpenAI",
	Long: `Translate text using OpenAI-compatible APIs.

Supports pipe and argument input:
  echo "hello world" | trans
  trans "hello world"
  cat README.md | trans -t ja`,
	Args: cobra.MinimumNArgs(0),
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&flagTo, "to", "t", "", "target language (default: zh)")
	rootCmd.Flags().StringVarP(&flagModel, "model", "m", "", "model to use (default: gpt-4o-mini)")
	rootCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "show source language annotation")
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
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

	// 5. Call OpenAI.
	client := llm.NewClient(cfg.APIKey, cfg.BaseURL, cfg.Model)

	if output.IsTerminal() {
		// Stream to terminal.
		err = client.TranslateStream(text, cfg.TargetLang, func(chunk string) {
			fmt.Print(chunk)
		})
		if err != nil {
			output.WriteErr("%v", err)
			return err
		}
		fmt.Println() // trailing newline
	} else {
		// Batch for pipe.
		result, err := client.Translate(text, cfg.TargetLang)
		if err != nil {
			output.WriteErr("%v", err)
			return err
		}
		w := output.NewWriter(cfg.Verbose)
		w.Write(result)
	}

	return nil
}
