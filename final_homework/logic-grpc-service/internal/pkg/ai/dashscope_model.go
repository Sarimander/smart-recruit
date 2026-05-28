package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

type DashScopeConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

type dashScopeChatModel struct {
	cfg    DashScopeConfig
	client *http.Client
}

func NewDashScopeChatModel(cfg DashScopeConfig) model.ChatModel {
	return &dashScopeChatModel{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (m *dashScopeChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	msgs := make([]chatMessage, 0, len(input))
	for _, msg := range input {
		msgs = append(msgs, chatMessage{Role: string(msg.Role), Content: msg.Content})
	}
	body, err := json.Marshal(chatRequest{Model: m.cfg.Model, Messages: msgs})
	if err != nil {
		return nil, err
	}

	url := m.cfg.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.cfg.APIKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dashscope request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result chatResponse
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("dashscope error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("empty response from dashscope")
	}
	return schema.AssistantMessage(result.Choices[0].Message.Content, nil), nil
}

func (m *dashScopeChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, err := m.Generate(ctx, input, opts...)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}

func (m *dashScopeChatModel) BindTools(tools []*schema.ToolInfo) error {
	return nil
}

func (m *dashScopeChatModel) GetType() string {
	return "DashScopeChatModel"
}

func (m *dashScopeChatModel) IsCallbacksEnabled() bool {
	return true
}
