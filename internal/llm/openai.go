package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type openaiProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIProvider(apiKey, model string) Provider {
	return &openaiProvider{apiKey: apiKey, model: model, client: &http.Client{}}
}

func (p *openaiProvider) Name() string  { return "openai" }
func (p *openaiProvider) Model() string { return p.model }

type openaiReq struct {
	Model          string            `json:"model"`
	ResponseFormat map[string]string `json:"response_format"`
	Messages       []openaiMessage   `json:"messages"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *openaiProvider) ParseExpense(ctx context.Context, message string, categories []string, defaultCurrency string) (*ParsedExpense, *Usage, error) {
	body, _ := json.Marshal(openaiReq{
		Model:          p.model,
		ResponseFormat: map[string]string{"type": "json_object"},
		Messages: []openaiMessage{
			{Role: "system", Content: buildSystemPrompt(categories, defaultCurrency)},
			{Role: "user", Content: message},
		},
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("openai: request failed: %w", err)
	}
	defer resp.Body.Close()

	var oResp openaiResp
	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil {
		return nil, nil, fmt.Errorf("openai: decode failed: %w", err)
	}
	if oResp.Error != nil {
		return nil, nil, fmt.Errorf("openai: %s", oResp.Error.Message)
	}
	if len(oResp.Choices) == 0 {
		return nil, nil, fmt.Errorf("openai: empty response")
	}

	var parsed ParsedExpense
	if err := json.Unmarshal([]byte(stripJSON(oResp.Choices[0].Message.Content)), &parsed); err != nil {
		return nil, nil, fmt.Errorf("openai: parse json: %w", err)
	}
	parsed.Normalize(defaultCurrency)

	return &parsed, &Usage{
		PromptTokens: oResp.Usage.PromptTokens,
		OutputTokens: oResp.Usage.CompletionTokens,
	}, nil
}

// ParseReceipt is not yet implemented for OpenAI — falls back to an error so the caller can surface it.
func (p *openaiProvider) ParseReceipt(_ context.Context, _ []byte, _ string, _ []string, _ string) (*ParsedExpense, *Usage, error) {
	return nil, nil, fmt.Errorf("receipt parsing is only supported with the Anthropic provider")
}

// Chat sends a multi-turn conversation to OpenAI and returns a parsed response.
func (p *openaiProvider) Chat(ctx context.Context, messages []ChatMessage, categories []string, defaultCurrency string) (*ParsedExpense, *Usage, error) {
	msgs := []openaiMessage{{Role: "system", Content: buildChatSystemPrompt(categories, defaultCurrency)}}
	for _, m := range messages {
		msgs = append(msgs, openaiMessage{Role: m.Role, Content: m.Content})
	}

	body, _ := json.Marshal(openaiReq{
		Model:          p.model,
		ResponseFormat: map[string]string{"type": "json_object"},
		Messages:       msgs,
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("openai chat: request failed: %w", err)
	}
	defer resp.Body.Close()

	var oResp openaiResp
	if err := json.NewDecoder(resp.Body).Decode(&oResp); err != nil {
		return nil, nil, fmt.Errorf("openai chat: decode failed: %w", err)
	}
	if oResp.Error != nil {
		return nil, nil, fmt.Errorf("openai chat: %s", oResp.Error.Message)
	}
	if len(oResp.Choices) == 0 {
		return nil, nil, fmt.Errorf("openai chat: empty response")
	}

	var parsed ParsedExpense
	if err := json.Unmarshal([]byte(stripJSON(oResp.Choices[0].Message.Content)), &parsed); err != nil {
		return nil, nil, fmt.Errorf("openai chat: parse json: %w", err)
	}
	parsed.Normalize(defaultCurrency)

	return &parsed, &Usage{
		PromptTokens: oResp.Usage.PromptTokens,
		OutputTokens: oResp.Usage.CompletionTokens,
	}, nil
}
