# trans - Terminal AI Translation Tool Design

## Overview

A pipe-friendly terminal translation tool powered by OpenAI-compatible APIs. Zero external dependencies beyond Cobra.

## Input Modes

Priority: **stdin pipe > CLI args > error**

- `echo "hello" | trans` — read from pipe
- `trans "hello world"` — read from args
- `trans hello world` — multiple args joined by space
- No input → error + usage hint

## Translation Direction

- Default: auto-detect source → Chinese (`zh`)
- Override: `trans -t ja "hello"` specify target language
- Configurable via `TRANS_TARGET_LANG` env or config file

## Output Format

- Default: **pure text** (pipe-friendly)
- Verbose (`-v`): `[en→zh] 你好世界` with source language tag
- Auto-detect terminal: `isatty()` — stream to terminal, batch to pipe

## Configuration

Priority chain: **CLI flags > env vars > config file > defaults**

### `~/.trans.json`

```json
{
  "api_key": "sk-xxx",
  "base_url": "https://api.openai.com/v1",
  "model": "gpt-4o-mini",
  "target_lang": "zh"
}
```

### Environment Variables

| Variable | Purpose |
|----------|---------|
| `OPENAI_API_KEY` | API Key |
| `OPENAI_BASE_URL` | Custom Base URL |
| `TRANS_MODEL` | Model selection |
| `TRANS_TARGET_LANG` | Default target language |

### CLI Flags

```
-t, --to string    target language (default "zh")
-m, --model string  model (default "gpt-4o-mini")
-v, --verbose       show source language annotation
```

Note: No `--api-key` flag to avoid `ps aux` leaking secrets.

## API Call

- Raw `net/http` to `/chat/completions`, no SDK
- Stream mode: SSE parsing with stdlib
- Batch mode: full response wait
- 30s timeout
- All errors to stderr, never pollute stdout

## Architecture

```
trans/
├── cmd/
│   └── root.go          # Cobra entry, arg parsing, input reading
├── internal/
│   ├── config/
│   │   └── config.go    # Config loading chain
│   ├── input/
│   │   └── input.go     # stdin/args unified input
│   ├── llm/
│   │   └── openai.go    # OpenAI HTTP call + SSE parsing
│   └── output/
│       └── output.go    # isatty detection + formatting
├── main.go
├── go.mod
└── go.sum
```

~410 lines total, 6 files.

## Prompt

```
System: You are a translation engine. Translate the following text to {target_lang}.
Rules:
- Output ONLY the translated text, nothing else.
- If the text is already in {target_lang}, return it unchanged.
- Preserve the original formatting (markdown, code blocks, newlines).
- Detect the source language automatically.

User: {input text}
```
