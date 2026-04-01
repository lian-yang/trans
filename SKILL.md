---
name: trans
description: >
  Terminal AI translation CLI tool (Go). Use when working with this repository:
  modifying translation logic, adding CLI flags/subcommands, changing config handling,
  updating the self-update mechanism, modifying CI/release workflows, or adding new
  OpenAI-compatible provider support. Triggers on: trans, translation CLI, openai CLI tool,
  cobra CLI, self-update, homebrew tap, SSE streaming.
---

# trans — Terminal AI Translation Tool

Go CLI that translates text via OpenAI-compatible chat/completions API. Pipe-friendly, zero-SDK.

## Architecture

```
main.go              → entry point, calls cmd.Execute()
cmd/
  root.go            → cobra root command, flag parsing, run pipeline
  update.go          → cobra update subcommand
internal/
  config/config.go   → three-tier config: file (~/.trans.json) → env → defaults
  input/input.go     → dual input: args join || stdin pipe (non-blocking)
  output/output.go   → TTY detection, verbose annotation, stderr errors
  llm/openai.go      → raw HTTP client, SSE stream parser, batch translate, language detect
  updater/updater.go  → GitHub Release check, tar.gz download, atomic binary replace
```

### Execution Flow (root command)

1. `--version` check → exit early
2. `config.Load()` → file → env → defaults
3. `input.Read(args)` → args || stdin pipe
4. Apply CLI flag overrides (model, target lang, verbose)
5. `cfg.Validate()` → API key required
6. Optional: `client.DetectLanguage()` (verbose mode)
7. Stream or batch output based on: `--no-stream` > `-s` > `isatty` auto-detect

## Key Design Decisions

- **Zero OpenAI SDK**: raw `net/http` + manual SSE line parsing (`parseSSE`)
- **Pipe-first**: stdout = clean translation only; errors → stderr via `output.WriteErr`
- **Three-tier version**: ldflags injection > `runtime/debug.ReadBuildInfo().vcs.tag` > "unknown"
- **Atomic self-update**: write `.new` → backup `.old` → rename → cleanup; rollback on failure
- **Homebrew detection**: path contains "homebrew" or "Cellar" → redirect to `brew upgrade`

## Build & Run

```bash
# Dev build
go build -o trans .

# Build with version injection
go build -ldflags "-X github.com/lian-yang/trans/cmd.version=v1.0.5" -o trans .

# Run tests
go test ./...

# Vet
go vet ./...
```

## Configuration Hierarchy

CLI flags > env vars (`OPENAI_API_KEY`, `OPENAI_BASE_URL`, `TRANS_MODEL`, `TRANS_TARGET_LANG`) > `~/.trans.json` > defaults (`gpt-4o-mini`, `zh`, `https://api.openai.com/v1`)

## Release & Distribution

- **Trigger**: push tag `v*` → `.github/workflows/release.yml`
- **Platforms**: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- **Artifacts**: per-platform tar.gz + checksums.txt → GitHub Release
- **Homebrew**: auto-updates `lian-yang/homebrew-tap` Formula/trans.rb via `HOMEBREW_TAP_TOKEN`
- **CI**: vet + build + test on push/PR to main

## Common Development Patterns

### Adding a new CLI flag

1. Add var in `cmd/root.go` init()
2. Register with `rootCmd.Flags()`
3. Apply override in `run()` function
4. Update README.md and README_CN.md flags table

### Adding a new subcommand

1. Create `cmd/<name>.go` with `cobra.Command`
2. Register in `init()` via `rootCmd.AddCommand()`
3. Keep logic in `internal/` — cmd files are thin wrappers

### Adding LLM features

1. Add method to `internal/llm/openai.go` Client struct
2. Reuse `doRequest()` for HTTP, `parseSSE()` for streaming
3. System prompt construction stays in `BuildSystemPrompt()` or new dedicated builder

### Modifying config

1. Add field to `Config` struct in `internal/config/config.go`
2. Add JSON tag for file persistence
3. Add env overlay in `applyEnv()`
4. Add default in `applyDefaults()`
5. Add setter for CLI flag override
