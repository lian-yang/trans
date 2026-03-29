# trans

[中文文档](README_CN.md)

Terminal AI translation tool powered by OpenAI-compatible APIs.

Pipe-friendly, zero-friction translation at your fingertips.

## Install

```bash
go build -o trans .
# or
go install github.com/lian-yang/trans@latest
```

## Usage

```bash
# Translate to Chinese (default)
trans "hello world"
# → 你好，世界

# Pipe support
echo "The quick brown fox" | trans
cat README.md | trans

# Specify target language
trans -t ja "good morning"
# → おはようございます

trans -t ko "I love programming"
# → 나는 프로그래밍을 좋아합니다.

# Verbose mode (show source→target annotation)
trans -v "hello world"
# → [en→zh] 你好，世界

# Use a different model
trans -m gpt-4o "hello world"

# Print version
trans -V
# → v1.0.0
```

### Stream control

By default, output mode is auto-detected: **streaming** in terminal, **batch** in pipe.

```bash
# Force streaming (e.g. real-time output in pipe)
echo "hello" | trans -s

# Force batch (e.g. wait for full result in terminal)
trans --no-stream "hello world"
```

## Configuration

Priority: **CLI flags > environment variables > config file > defaults**

### Config file `~/.trans.json`

```json
{
  "api_key": "sk-xxx",
  "base_url": "https://api.deepseek.com/v1",
  "model": "deepseek-chat",
  "target_lang": "zh"
}
```

### Environment variables

| Variable | Purpose |
|----------|---------|
| `OPENAI_API_KEY` | API key |
| `OPENAI_BASE_URL` | Custom base URL (for OpenAI-compatible services) |
| `TRANS_MODEL` | Model name |
| `TRANS_TARGET_LANG` | Default target language |

### CLI flags

```
-t, --to string      target language (default: zh)
-m, --model string   model to use (default: gpt-4o-mini)
-s, --stream         force streaming output
    --no-stream      force batch output (disable streaming)
-v, --verbose        show source→target language annotation
-V, --version        print version and exit
```

## Compatible Providers

Any OpenAI-compatible API works. Just change `base_url` and `model`:

| Provider | base_url | model |
|----------|----------|-------|
| OpenAI | `https://api.openai.com/v1` | `gpt-4o-mini` |
| DeepSeek | `https://api.deepseek.com/v1` | `deepseek-chat` |
| Groq | `https://api.groq.com/openai/v1` | `llama-3.3-70b-versatile` |
| OpenRouter | `https://openrouter.ai/api/v1` | `openai/gpt-4o-mini` |

## Design Principles

- **Pipe-first**: stdout is clean text, errors go to stderr
- **Auto streaming**: streams to terminal, batches for pipes (`isatty` detection)
- **Zero SDK**: raw `net/http` + SSE parsing, no OpenAI SDK dependency
- **Minimal deps**: only [cobra](https://github.com/spf13/cobra) + stdlib

## License

MIT
