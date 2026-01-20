// Package planner 提供 AI 规划功能
package planner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/browser-automation/internal/domain"
)

// LLMClient LLM 客户端接口
type LLMClient interface {
	Chat(ctx context.Context, messages []Message) (*Response, error)
	Validate(ctx context.Context) error
}

// Message 消息
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response 响应
type Response struct {
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
	Usage        *Usage `json:"usage"`
}

// Usage 使用量
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// LLMClientFactory LLM 客户端工厂
type LLMClientFactory struct {
	httpClient *http.Client
}

// NewLLMClientFactory 创建 LLM 客户端工厂
func NewLLMClientFactory() *LLMClientFactory {
	return &LLMClientFactory{
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // 增加超时时间
		},
	}
}

// NewClient 根据配置创建客户端
func (f *LLMClientFactory) NewClient(config *domain.LLMConfig) (LLMClient, error) {
	switch config.Provider {
	case domain.LLMProviderAnthropic:
		return NewAnthropicClient(config, f.httpClient), nil
	default:
		// OpenAI 兼容接口（包括 OpenAI、DeepSeek、Ollama、本地代理等）
		return NewOpenAICompatibleClient(config, f.httpClient), nil
	}
}

// OpenAICompatibleClient OpenAI 兼容客户端
type OpenAICompatibleClient struct {
	config     *domain.LLMConfig
	httpClient *http.Client
}

// NewOpenAICompatibleClient 创建 OpenAI 兼容客户端
func NewOpenAICompatibleClient(config *domain.LLMConfig, httpClient *http.Client) *OpenAICompatibleClient {
	// 设置默认端点
	if config.Endpoint == "" {
		switch config.Provider {
		case domain.LLMProviderOpenAI:
			config.Endpoint = "https://api.openai.com/v1"
		case domain.LLMProviderDeepSeek:
			config.Endpoint = "https://api.deepseek.com/v1"
		case domain.LLMProviderOllama:
			config.Endpoint = "http://localhost:11434/v1"
		case domain.LLMProviderMoonshot:
			config.Endpoint = "https://api.moonshot.cn/v1"
		case domain.LLMProviderQwen:
			config.Endpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1"
		}
	}
	return &OpenAICompatibleClient{config: config, httpClient: httpClient}
}

// Chat 发送对话请求
func (c *OpenAICompatibleClient) Chat(ctx context.Context, messages []Message) (*Response, error) {
	log.Printf("[LLM] Chat request: model=%s, endpoint=%s", c.config.Model, c.config.Endpoint)
	
	reqBody := map[string]interface{}{
		"model":    c.config.Model,
		"messages": messages,
	}
	
	if c.config.Options != nil {
		if c.config.Options.Temperature > 0 {
			reqBody["temperature"] = c.config.Options.Temperature
		}
		if c.config.Options.MaxTokens > 0 {
			reqBody["max_tokens"] = c.config.Options.MaxTokens
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	
	log.Printf("[LLM] Request body size: %d bytes", len(body))

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.config.Endpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.config.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	}

	log.Printf("[LLM] Sending request...")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[LLM] Request failed: %v", err)
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("[LLM] Response status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("[LLM] Error response: %s", string(respBody))
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	var result OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	log.Printf("[LLM] Response received, content length: %d", len(result.Choices[0].Message.Content))

	return &Response{
		Content:      result.Choices[0].Message.Content,
		FinishReason: result.Choices[0].FinishReason,
		Usage: &Usage{
			PromptTokens:     result.Usage.PromptTokens,
			CompletionTokens: result.Usage.CompletionTokens,
			TotalTokens:      result.Usage.TotalTokens,
		},
	}, nil
}

// Validate 验证配置
func (c *OpenAICompatibleClient) Validate(ctx context.Context) error {
	_, err := c.Chat(ctx, []Message{
		{Role: "user", Content: "hi"},
	})
	return err
}

// OpenAIResponse OpenAI API 响应
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// AnthropicClient Anthropic 客户端
type AnthropicClient struct {
	config     *domain.LLMConfig
	httpClient *http.Client
}

// NewAnthropicClient 创建 Anthropic 客户端
func NewAnthropicClient(config *domain.LLMConfig, httpClient *http.Client) *AnthropicClient {
	if config.Endpoint == "" {
		config.Endpoint = "https://api.anthropic.com/v1"
	}
	return &AnthropicClient{config: config, httpClient: httpClient}
}

// Chat 发送对话请求
func (c *AnthropicClient) Chat(ctx context.Context, messages []Message) (*Response, error) {
	reqBody := map[string]interface{}{
		"model":      c.config.Model,
		"messages":   messages,
		"max_tokens": 4096,
	}

	if c.config.Options != nil {
		if c.config.Options.Temperature > 0 {
			reqBody["temperature"] = c.config.Options.Temperature
		}
		if c.config.Options.MaxTokens > 0 {
			reqBody["max_tokens"] = c.config.Options.MaxTokens
		}
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		c.config.Endpoint+"/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	var result AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	return &Response{
		Content:      result.Content[0].Text,
		FinishReason: result.StopReason,
		Usage: &Usage{
			PromptTokens:     result.Usage.InputTokens,
			CompletionTokens: result.Usage.OutputTokens,
			TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
		},
	}, nil
}

// Validate 验证配置
func (c *AnthropicClient) Validate(ctx context.Context) error {
	_, err := c.Chat(ctx, []Message{
		{Role: "user", Content: "hi"},
	})
	return err
}

// AnthropicResponse Anthropic API 响应
type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}
