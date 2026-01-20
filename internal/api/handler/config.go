// Package handler 提供 HTTP 请求处理
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/browser-automation/internal/domain"
	"github.com/browser-automation/internal/planner"
	"github.com/gin-gonic/gin"
)

// ConfigHandler 配置处理器
type ConfigHandler struct {
	llmFactory *planner.LLMClientFactory
}

// NewConfigHandler 创建配置处理器
func NewConfigHandler(llmFactory *planner.LLMClientFactory) *ConfigHandler {
	return &ConfigHandler{llmFactory: llmFactory}
}

// GetLLMPresets 获取 LLM 预设列表
func (h *ConfigHandler) GetLLMPresets(c *gin.Context) {
	presets := domain.GetLLMPresets()
	c.JSON(http.StatusOK, gin.H{
		"presets": presets,
	})
}

// ValidateLLMRequest LLM 验证请求
type ValidateLLMRequest struct {
	Provider    string  `json:"provider" binding:"required"`
	Model       string  `json:"model" binding:"required"`
	Endpoint    string  `json:"endpoint"`
	APIKey      string  `json:"api_key"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

// ValidateLLM 验证 LLM 配置
func (h *ConfigHandler) ValidateLLM(c *gin.Context) {
	var req ValidateLLMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := &domain.LLMConfig{
		Provider: domain.LLMProvider(req.Provider),
		Model:    req.Model,
		Endpoint: req.Endpoint,
		APIKey:   req.APIKey,
		Options: &domain.LLMOptions{
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
		},
	}

	client, err := h.llmFactory.NewClient(config)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":   false,
			"error":   err.Error(),
			"message": "无效的 LLM 配置",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := client.Validate(ctx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":   false,
			"error":   err.Error(),
			"message": "无法连接到 LLM 服务，请检查配置",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "LLM 配置验证成功",
	})
}

// GetOutputFormats 获取支持的输出格式
func (h *ConfigHandler) GetOutputFormats(c *gin.Context) {
	formats := domain.GetSupportedFormats()
	c.JSON(http.StatusOK, gin.H{
		"formats": formats,
	})
}

// GetAuthTypes 获取支持的认证类型
func (h *ConfigHandler) GetAuthTypes(c *gin.Context) {
	authTypes := []map[string]interface{}{
		{
			"type":        "none",
			"name":        "无需认证",
			"description": "目标网站无需登录即可访问",
		},
		{
			"type":        "form",
			"name":        "表单登录",
			"description": "使用用户名密码通过登录表单认证",
		},
		{
			"type":        "sso",
			"name":        "SSO 单点登录",
			"description": "通过企业 SSO 系统认证（OAuth2/SAML/OIDC）",
		},
		{
			"type":        "manual",
			"name":        "手动登录",
			"description": "打开浏览器窗口，手动完成登录（适合复杂认证如扫码、MFA）",
		},
		{
			"type":        "cookie",
			"name":        "Cookie 注入",
			"description": "直接注入已有的 Cookie 信息",
		},
		{
			"type":        "token",
			"name":        "Token 注入",
			"description": "使用 Bearer Token 或 API Key 认证",
		},
	}
	c.JSON(http.StatusOK, gin.H{
		"auth_types": authTypes,
	})
}
