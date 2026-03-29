# trans

[English](README.md)

基于 OpenAI 兼容 API 的终端 AI 翻译工具。

管道友好，即装即用。

## 安装

```bash
go build -o trans .
# 或
go install github.com/lian-yang/trans@latest
```

## 使用

```bash
# 翻译为中文（默认）
trans "hello world"
# → 你好，世界

# 管道输入
echo "The quick brown fox" | trans
cat README.md | trans

# 指定目标语言
trans -t ja "good morning"
# → おはようございます

trans -t ko "I love programming"
# → 나는 프로그래밍을 좋아합니다.

# 详细模式（显示语言标注）
trans -v "hello world"
# → [en→zh] 你好，世界

# 指定模型
trans -m gpt-4o "hello world"
```

## 配置

优先级：**命令行参数 > 环境变量 > 配置文件 > 默认值**

### 配置文件 `~/.trans.json`

```json
{
  "api_key": "sk-xxx",
  "base_url": "https://api.deepseek.com/v1",
  "model": "deepseek-chat",
  "target_lang": "zh"
}
```

### 环境变量

| 变量 | 用途 |
|------|------|
| `OPENAI_API_KEY` | API 密钥 |
| `OPENAI_BASE_URL` | 自定义 API 地址（兼容服务） |
| `TRANS_MODEL` | 模型名称 |
| `TRANS_TARGET_LANG` | 默认目标语言 |

### 命令行参数

```
-t, --to string      目标语言（默认：zh）
-m, --model string   模型（默认：gpt-4o-mini）
-v, --verbose        显示源语言标注
```

## 兼容服务商

只需修改 `base_url` 和 `model` 即可接入任意 OpenAI 兼容 API：

| 服务商 | base_url | model |
|--------|----------|-------|
| OpenAI | `https://api.openai.com/v1` | `gpt-4o-mini` |
| DeepSeek | `https://api.deepseek.com/v1` | `deepseek-chat` |
| Groq | `https://api.groq.com/openai/v1` | `llama-3.3-70b-versatile` |
| OpenRouter | `https://openrouter.ai/api/v1` | `openai/gpt-4o-mini` |

## 设计原则

- **管道优先**：stdout 只输出翻译文本，错误走 stderr，不污染管道链
- **自动流式**：终端模式流式输出，管道模式批量返回（`isatty` 检测）
- **零 SDK**：纯 `net/http` + SSE 解析，不依赖 OpenAI SDK
- **极简依赖**：仅 [cobra](https://github.com/spf13/cobra) + 标准库

## 许可证

MIT
