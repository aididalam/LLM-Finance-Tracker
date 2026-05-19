package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type anthropicProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewAnthropicProvider(apiKey, model string) Provider {
	return &anthropicProvider{apiKey: apiKey, model: model, client: &http.Client{}}
}

func (p *anthropicProvider) Name() string  { return "anthropic" }
func (p *anthropicProvider) Model() string { return p.model }

type anthropicReq struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResp struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *anthropicProvider) ParseExpense(ctx context.Context, message string, categories []string, defaultCurrency string) (*ParsedExpense, *Usage, error) {
	body, _ := json.Marshal(anthropicReq{
		Model:     p.model,
		MaxTokens: 512,
		System:    buildSystemPrompt(categories, defaultCurrency),
		Messages:  []anthropicMessage{{Role: "user", Content: message}},
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic: request failed: %w", err)
	}
	defer resp.Body.Close()

	var ar anthropicResp
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, nil, fmt.Errorf("anthropic: decode failed: %w", err)
	}
	if ar.Error != nil {
		return nil, nil, fmt.Errorf("anthropic: %s", ar.Error.Message)
	}
	if len(ar.Content) == 0 {
		return nil, nil, fmt.Errorf("anthropic: empty response")
	}

	var parsed ParsedExpense
	if err := json.Unmarshal([]byte(stripJSON(ar.Content[0].Text)), &parsed); err != nil {
		return nil, nil, fmt.Errorf("anthropic: parse json: %w", err)
	}
	parsed.Normalize(defaultCurrency)

	return &parsed, &Usage{
		PromptTokens: ar.Usage.InputTokens,
		OutputTokens: ar.Usage.OutputTokens,
	}, nil
}

// ParseReceipt sends a receipt image to the Anthropic vision API and returns parsed expense details.
func (p *anthropicProvider) ParseReceipt(ctx context.Context, imageData []byte, mediaType string, categories []string, defaultCurrency string) (*ParsedExpense, *Usage, error) {
	// Resize large images and enforce PDF size limits before spending tokens.
	var err error
	imageData, mediaType, err = PreprocessReceipt(imageData, mediaType)
	if err != nil {
		return nil, nil, err
	}

	// Build a multimodal message: image/document block + text instruction
	var docBlock map[string]any
	if mediaType == "application/pdf" {
		docBlock = map[string]any{
			"type": "document",
			"source": map[string]string{
				"type":       "base64",
				"media_type": "application/pdf",
				"data":       base64.StdEncoding.EncodeToString(imageData),
			},
		}
	} else {
		docBlock = map[string]any{
			"type": "image",
			"source": map[string]string{
				"type":       "base64",
				"media_type": mediaType,
				"data":       base64.StdEncoding.EncodeToString(imageData),
			},
		}
	}

	content := []map[string]any{
		docBlock,
		{
			"type": "text",
			"text": "This is a receipt or bill. Extract the expense details and respond with ONLY the JSON object described in the system prompt. If you cannot read the receipt clearly, set is_expense=false and explain in not_expense_reply.",
		},
	}

	contentJSON, _ := json.Marshal(content)

	type visionMessage struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}

	type visionReq struct {
		Model     string          `json:"model"`
		MaxTokens int             `json:"max_tokens"`
		System    string          `json:"system"`
		Messages  []visionMessage `json:"messages"`
	}

	body, _ := json.Marshal(visionReq{
		Model:     p.model,
		MaxTokens: 1024,
		System:    buildReceiptSystemPrompt(categories, defaultCurrency),
		Messages:  []visionMessage{{Role: "user", Content: json.RawMessage(contentJSON)}},
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic receipt: request failed: %w", err)
	}
	defer resp.Body.Close()

	var ar anthropicResp
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, nil, fmt.Errorf("anthropic receipt: decode failed: %w", err)
	}
	if ar.Error != nil {
		return nil, nil, fmt.Errorf("anthropic receipt: %s", ar.Error.Message)
	}
	if len(ar.Content) == 0 {
		return nil, nil, fmt.Errorf("anthropic receipt: empty response")
	}

	var parsed ParsedExpense
	if err := json.Unmarshal([]byte(stripJSON(ar.Content[0].Text)), &parsed); err != nil {
		return nil, nil, fmt.Errorf("anthropic receipt: parse json: %w", err)
	}
	parsed.Normalize(defaultCurrency)

	return &parsed, &Usage{
		PromptTokens: ar.Usage.InputTokens,
		OutputTokens: ar.Usage.OutputTokens,
	}, nil
}

// Chat sends a multi-turn conversation to Anthropic and returns a parsed response.
func (p *anthropicProvider) Chat(ctx context.Context, messages []ChatMessage, categories []string, defaultCurrency string) (*ParsedExpense, *Usage, error) {
	msgs := make([]anthropicMessage, len(messages))
	for i, m := range messages {
		msgs[i] = anthropicMessage{Role: m.Role, Content: m.Content}
	}

	body, _ := json.Marshal(anthropicReq{
		Model:     p.model,
		MaxTokens: 512,
		System:    buildChatSystemPrompt(categories, defaultCurrency),
		Messages:  msgs,
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("anthropic chat: request failed: %w", err)
	}
	defer resp.Body.Close()

	var ar anthropicResp
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, nil, fmt.Errorf("anthropic chat: decode failed: %w", err)
	}
	if ar.Error != nil {
		return nil, nil, fmt.Errorf("anthropic chat: %s", ar.Error.Message)
	}
	if len(ar.Content) == 0 {
		return nil, nil, fmt.Errorf("anthropic chat: empty response")
	}

	var parsed ParsedExpense
	if err := json.Unmarshal([]byte(stripJSON(ar.Content[0].Text)), &parsed); err != nil {
		return nil, nil, fmt.Errorf("anthropic chat: parse json: %w", err)
	}
	parsed.Normalize(defaultCurrency)

	return &parsed, &Usage{
		PromptTokens: ar.Usage.InputTokens,
		OutputTokens: ar.Usage.OutputTokens,
	}, nil
}
