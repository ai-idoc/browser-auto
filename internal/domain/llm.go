// Package domain 定义核心业务模型
package domain

// LLMProvider LLM 提供商
type LLMProvider string

const (
	LLMProviderOpenAI     LLMProvider = "openai"
	LLMProviderAnthropic  LLMProvider = "anthropic"
	LLMProviderAzure      LLMProvider = "azure"
	LLMProviderGoogle     LLMProvider = "google"
	LLMProviderDeepSeek   LLMProvider = "deepseek"
	LLMProviderQwen       LLMProvider = "qwen"
	LLMProviderZhipu      LLMProvider = "zhipu"
	LLMProviderMoonshot   LLMProvider = "moonshot"
	LLMProviderOllama     LLMProvider = "ollama"
	LLMProviderLocalProxy LLMProvider = "local_proxy"
	LLMProviderCustom     LLMProvider = "custom"
)

// LLMConfig LLM 配置
type LLMConfig struct {
	Provider LLMProvider `json:"provider"`
	Model    string      `json:"model"`
	Endpoint string      `json:"endpoint"`
	APIKey   string      `json:"api_key,omitempty"`
	Options  *LLMOptions `json:"options,omitempty"`
}

// LLMOptions LLM 高级选项
type LLMOptions struct {
	Temperature      float64 `json:"temperature"`
	MaxTokens        int     `json:"max_tokens"`
	TopP             float64 `json:"top_p"`
	FrequencyPenalty float64 `json:"frequency_penalty"`
	PresencePenalty  float64 `json:"presence_penalty"`
	Timeout          int     `json:"timeout"`
	RetryCount       int     `json:"retry_count"`
}

// LLMPreset LLM 预设配置
type LLMPreset struct {
	Provider        LLMProvider `json:"provider"`
	Name            string      `json:"name"`
	DefaultModel    string      `json:"default_model"`
	AvailableModels []string    `json:"available_models"`
	DefaultEndpoint string      `json:"default_endpoint"`
	RequiresAPIKey  bool        `json:"requires_api_key"`
}

// GetLLMPresets 获取所有 LLM 预设
func GetLLMPresets() []LLMPreset {
	return []LLMPreset{
		{
			Provider:        LLMProviderOpenAI,
			Name:            "OpenAI",
			DefaultModel:    "gpt-4o",
			AvailableModels: []string{"gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-3.5-turbo"},
			DefaultEndpoint: "https://api.openai.com/v1",
			RequiresAPIKey:  true,
		},
		{
			Provider:        LLMProviderAnthropic,
			Name:            "Anthropic (Claude)",
			DefaultModel:    "claude-sonnet-4-20250514",
			AvailableModels: []string{"claude-sonnet-4-20250514", "claude-opus-4-20250514", "claude-3-5-sonnet-20241022"},
			DefaultEndpoint: "https://api.anthropic.com/v1",
			RequiresAPIKey:  true,
		},
		{
			Provider:        LLMProviderDeepSeek,
			Name:            "DeepSeek",
			DefaultModel:    "deepseek-chat",
			AvailableModels: []string{"deepseek-chat", "deepseek-coder"},
			DefaultEndpoint: "https://api.deepseek.com/v1",
			RequiresAPIKey:  true,
		},
		{
			Provider:        LLMProviderQwen,
			Name:            "通义千问",
			DefaultModel:    "qwen-max",
			AvailableModels: []string{"qwen-max", "qwen-plus", "qwen-turbo"},
			DefaultEndpoint: "https://dashscope.aliyuncs.com/compatible-mode/v1",
			RequiresAPIKey:  true,
		},
		{
			Provider:        LLMProviderMoonshot,
			Name:            "Moonshot (Kimi)",
			DefaultModel:    "moonshot-v1-8k",
			AvailableModels: []string{"moonshot-v1-8k", "moonshot-v1-32k", "moonshot-v1-128k"},
			DefaultEndpoint: "https://api.moonshot.cn/v1",
			RequiresAPIKey:  true,
		},
		{
			Provider:        LLMProviderOllama,
			Name:            "Ollama (本地)",
			DefaultModel:    "llama3.1",
			AvailableModels: []string{"llama3.1", "qwen2.5", "mistral", "codellama", "deepseek-coder"},
			DefaultEndpoint: "http://localhost:11434/v1",
			RequiresAPIKey:  false,
		},
		{
			Provider:        LLMProviderLocalProxy,
			Name:            "本地代理",
			DefaultModel:    "",
			AvailableModels: []string{},
			DefaultEndpoint: "http://localhost:8000/v1",
			RequiresAPIKey:  false,
		},
		{
			Provider:        LLMProviderCustom,
			Name:            "自定义 (OpenAI 兼容)",
			DefaultModel:    "",
			AvailableModels: []string{},
			DefaultEndpoint: "",
			RequiresAPIKey:  false,
		},
	}
}

// DefaultLLMOptions 默认 LLM 选项
func DefaultLLMOptions() *LLMOptions {
	return &LLMOptions{
		Temperature: 0.7,
		MaxTokens:   4096,
		TopP:        1.0,
		Timeout:     60,
		RetryCount:  3,
	}
}
