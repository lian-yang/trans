package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client wraps OpenAI API calls.
type Client struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewClient creates an OpenAI client.
func NewClient(apiKey, baseURL, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// chatRequest is the request body for /chat/completions.
type chatRequest struct {
	Model    string        `json:"model"`
	Stream   bool          `json:"stream"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatResponse is the non-stream response.
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// streamChunk is a single SSE chunk.
type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// BuildSystemPrompt constructs the translation system prompt.
func BuildSystemPrompt(targetLang string) string {
	return fmt.Sprintf(`You are a translation engine. Translate the following text to %s.
Rules:
- Output ONLY the translated text, nothing else.
- If the text is already in %s, return it unchanged.
- Preserve the original formatting (markdown, code blocks, newlines).
- Detect the source language automatically.`, targetLang, targetLang)
}

// Translate sends a non-stream request and returns the full translation.
func (c *Client) Translate(text, targetLang string) (string, error) {
	reqBody := chatRequest{
		Model:  c.model,
		Stream: false,
		Messages: []chatMessage{
			{Role: "system", Content: BuildSystemPrompt(targetLang)},
			{Role: "user", Content: text},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.doRequest(body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp chatResponse
		if json.Unmarshal(data, &errResp) == nil && errResp.Error != nil {
			return "", fmt.Errorf("API error %s: %s", errResp.Error.Code, errResp.Error.Message)
		}
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(data))
	}

	var result chatResponse
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return result.Choices[0].Message.Content, nil
}

// TranslateStream sends a stream request and calls onChunk for each token.
func (c *Client) TranslateStream(text, targetLang string, onChunk func(string)) error {
	reqBody := chatRequest{
		Model:  c.model,
		Stream: true,
		Messages: []chatMessage{
			{Role: "system", Content: BuildSystemPrompt(targetLang)},
			{Role: "user", Content: text},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Streaming needs longer timeout — disable client-level timeout.
	c.client.Timeout = 0
	resp, err := c.doRequest(body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(resp.Body)
		var errResp chatResponse
		if json.Unmarshal(data, &errResp) == nil && errResp.Error != nil {
			return fmt.Errorf("API error %s: %s", errResp.Error.Code, errResp.Error.Message)
		}
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(data))
	}

	return parseSSE(resp.Body, onChunk)
}

// doRequest creates and sends an HTTP request.
func (c *Client) doRequest(body []byte) (*http.Response, error) {
	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return resp, nil
}

// parseSSE reads Server-Sent Events from the response body.
func parseSSE(r io.Reader, onChunk func(string)) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// SSE lines start with "data: ".
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Stream end marker.
		if data == "[DONE]" {
			return nil
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // skip malformed chunks
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				onChunk(content)
			}
		}
	}

	return scanner.Err()
}
