// Gerege Template AI v1.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

// Package aiclient нь Anthropic Messages API-ийн streaming дуудлагыг
// нимгэн боож HTTP клиент болгоно (pkg/verify-тэй ижил загвар). SDK
// хамаарал нэмэхгүйгээр зөвхөн стандарт сангаар SSE урсгалыг уншина.
//
// Аюулгүй байдал: API түлхүүр зөвхөн энэ процессын санах ойд байна —
// browser/клиент рүү хэзээ ч дамжихгүй (BFF → backend → Anthropic).
package aiclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// DefaultBaseURL нь Anthropic API-ийн production endpoint юм.
const DefaultBaseURL = "https://api.anthropic.com"

// apiVersion нь Anthropic API-ийн хувилбарын толгой. Огноо хэлбэртэй
// боловч энэ нь API гэрээний хувилбар — модель хувилбар биш.
const apiVersion = "2023-06-01"

// Message нь харилцан ярианы нэг мөр (Anthropic-ийн user/assistant role).
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage нь нэг дуудлагын токен зарцуулалт — кост хяналт, метерингд
// зориулж дуудагч руу буцаана.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// Streamer нь usecase давхаргын харьцах хил (boundary) — тестэд бодит
// HTTP алхалгүй мокчлох боломж олгоно (pkg/verify.Sender-тэй ижил санаа).
type Streamer interface {
	// StreamMessage нь system prompt + メッセж түүхээр Claude-г дуудаж,
	// текст хэсэг бүрийг onDelta callback-аар дамжуулна. onDelta алдаа
	// буцаавал урсгал зогсоно (клиент салсан гэх мэт). Бүрэн хариу
	// болон токен зарцуулалтыг буцаана.
	StreamMessage(ctx context.Context, system string, messages []Message, onDelta func(delta string) error) (full string, usage Usage, err error)
}

// Client нь Anthropic Messages API-ийн HTTP клиент юм.
type Client struct {
	baseURL   string
	apiKey    string
	model     string
	maxTokens int
	http      *http.Client
}

// NewClient нь шинэ Client үүсгэнэ. apiKey хоосон үед клиент бүтэх боловч
// дуудлага бүр "api key is not configured" алдаа буцаана — операторт
// чимээгүй буруу тохиргоо үлдээхгүй (verify.NewClient-тэй ижил зарчим).
// timeout нь streaming дуудлагын нийт дээд хугацаа.
func NewClient(apiKey, model string, maxTokens int, timeout time.Duration) *Client {
	if model == "" {
		model = "claude-sonnet-4-6"
	}
	if maxTokens <= 0 {
		maxTokens = 1024
	}
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &Client{
		baseURL:   DefaultBaseURL,
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		http:      &http.Client{Timeout: timeout},
	}
}

// Model нь тохируулагдсан модель нэрийг буцаана (usage метерингд бичигдэнэ).
func (c *Client) Model() string { return c.model }

// Configured нь API түлхүүр тохируулагдсан эсэхийг буцаана.
func (c *Client) Configured() bool { return c.apiKey != "" }

type messagesReq struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
}

// SSE event-ийн бидэнд хэрэгтэй хэсгүүд. Anthropic-ийн streaming формат:
// message_start (input usage) → content_block_delta* (текст) →
// message_delta (output usage) → message_stop.
type sseEvent struct {
	Type    string `json:"type"`
	Message *struct {
		Usage struct {
			InputTokens int `json:"input_tokens"`
		} `json:"usage"`
	} `json:"message,omitempty"`
	Delta *struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta,omitempty"`
	Usage *struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// StreamMessage нь Streamer интерфейсийг хэрэгжүүлнэ.
func (c *Client) StreamMessage(ctx context.Context, system string, messages []Message, onDelta func(delta string) error) (string, Usage, error) {
	var usage Usage
	if c.apiKey == "" {
		return "", usage, errors.New("aiclient: api key is not configured")
	}

	payload, err := json.Marshal(messagesReq{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    system,
		Messages:  messages,
		Stream:    true,
	})
	if err != nil {
		return "", usage, fmt.Errorf("aiclient: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return "", usage, fmt.Errorf("aiclient: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", usage, fmt.Errorf("aiclient: send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Алдааны body нь жижиг JSON — хязгаартай уншаад мессежийг ил болго.
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		var er struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(raw, &er)
		msg := strings.TrimSpace(er.Error.Message)
		if msg == "" {
			msg = strings.TrimSpace(string(raw))
		}
		return "", usage, fmt.Errorf("aiclient: messages returned %d: %s", resp.StatusCode, msg)
	}

	// SSE урсгалыг мөр мөрөөр уншина. "data: {json}" мөрүүд л бидэнд
	// хэрэгтэй; "event:" мөрүүдийн нэр нь data доторх type-тэй давхардна.
	var sb strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	// Нэг event 64KB-аас том байж болзошгүй (урт текст delta) тул буферийг өсгөнө.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		var ev sseEvent
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			// Танигдахгүй event-ийг алгасна — API ирээдүйд шинэ төрөл нэмж болно.
			continue
		}
		switch ev.Type {
		case "message_start":
			if ev.Message != nil {
				usage.InputTokens = ev.Message.Usage.InputTokens
			}
		case "content_block_delta":
			if ev.Delta != nil && ev.Delta.Type == "text_delta" && ev.Delta.Text != "" {
				sb.WriteString(ev.Delta.Text)
				if onDelta != nil {
					if cbErr := onDelta(ev.Delta.Text); cbErr != nil {
						return sb.String(), usage, fmt.Errorf("aiclient: delta callback: %w", cbErr)
					}
				}
			}
		case "message_delta":
			if ev.Usage != nil {
				usage.OutputTokens = ev.Usage.OutputTokens
			}
		case "error":
			if ev.Error != nil {
				return sb.String(), usage, fmt.Errorf("aiclient: stream error: %s", ev.Error.Message)
			}
			return sb.String(), usage, errors.New("aiclient: unknown stream error")
		}
	}
	if err := scanner.Err(); err != nil {
		return sb.String(), usage, fmt.Errorf("aiclient: read stream: %w", err)
	}
	return sb.String(), usage, nil
}
