// Package handler 提供 HTTP 请求处理
package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/browser-automation/internal/domain"
	"github.com/browser-automation/internal/orchestrator"
	"github.com/browser-automation/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskHandler 任务处理器
type TaskHandler struct {
	taskStore    storage.TaskStore
	orchestrator *orchestrator.Orchestrator
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskStore storage.TaskStore, orch *orchestrator.Orchestrator) *TaskHandler {
	return &TaskHandler{
		taskStore:    taskStore,
		orchestrator: orch,
	}
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Description string               `json:"description" binding:"required"`
	TargetURL   string               `json:"target_url" binding:"required,url"`
	Auth        *AuthConfigRequest   `json:"auth,omitempty"`
	LLM         *LLMConfigRequest    `json:"llm" binding:"required"`
	Output      *OutputConfigRequest `json:"output,omitempty"`
}

// AuthConfigRequest 认证配置请求
type AuthConfigRequest struct {
	Type        string          `json:"type" binding:"required,oneof=none form sso manual cookie token"`
	Username    string          `json:"username,omitempty"`
	Password    string          `json:"password,omitempty"`
	SSOProvider string          `json:"sso_provider,omitempty"`
	SSOLoginURL string          `json:"sso_login_url,omitempty"`
	Token       string          `json:"token,omitempty"`
	SessionID   string          `json:"session_id,omitempty"`
	Cookies     []CookieRequest `json:"cookies,omitempty"`
}

// CookieRequest Cookie 请求
type CookieRequest struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Domain   string `json:"domain"`
	Path     string `json:"path"`
	Secure   bool   `json:"secure"`
	HTTPOnly bool   `json:"http_only"`
}

// LLMConfigRequest LLM 配置请求
type LLMConfigRequest struct {
	Provider    string  `json:"provider" binding:"required"`
	Model       string  `json:"model" binding:"required"`
	Endpoint    string  `json:"endpoint"`
	APIKey      string  `json:"api_key,omitempty"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`
}

// OutputConfigRequest 输出配置请求
type OutputConfigRequest struct {
	Formats          []string `json:"formats" binding:"required,min=1"`
	Language         string   `json:"language"`
	Title            string   `json:"title"`
	ScreenshotQuality int     `json:"screenshot_quality"`
	Annotate         bool     `json:"annotate"`
	IncludeTOC       bool     `json:"include_toc"`
	IncludeCover     bool     `json:"include_cover"`
	Template         string   `json:"template"`
	LogoURL          string   `json:"logo_url"`
	ThemeColor       string   `json:"theme_color"`
}

// CreateTask 创建任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := &domain.Task{
		ID:          uuid.New().String(),
		Description: req.Description,
		TargetURL:   req.TargetURL,
		Status:      domain.TaskStatusPending,
		Auth:        h.convertAuthConfig(req.Auth),
		LLM:         h.convertLLMConfig(req.LLM),
		Output:      h.convertOutputConfig(req.Output),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.taskStore.Create(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create task"})
		return
	}

	// 异步执行任务
	go func() {
		ctx := context.Background()
		if err := h.orchestrator.ExecuteTask(ctx, task); err != nil {
			log.Printf("Task execution failed: %v", err)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"task_id": task.ID,
		"status":  task.Status,
		"message": "任务已创建，正在处理中",
	})
}

// GetTask 获取任务详情
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")
	
	task, err := h.taskStore.Get(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// ListTasks 获取任务列表
func (h *TaskHandler) ListTasks(c *gin.Context) {
	tasks, err := h.taskStore.List(c.Request.Context(), 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"total": len(tasks),
	})
}

// CancelTask 取消任务
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")
	
	if err := h.taskStore.UpdateStatus(c.Request.Context(), taskID, domain.TaskStatusCancelled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已取消"})
}

func (h *TaskHandler) convertAuthConfig(req *AuthConfigRequest) *domain.AuthConfig {
	if req == nil {
		return nil
	}

	// 转换 Cookies
	var cookies []domain.Cookie
	for _, c := range req.Cookies {
		cookies = append(cookies, domain.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
		})
	}

	return &domain.AuthConfig{
		Type: domain.AuthType(req.Type),
		Credentials: &domain.Credentials{
			Username: req.Username,
			Password: req.Password,
			Token:    req.Token,
		},
		SSOConfig: &domain.SSOConfig{
			Provider: domain.SSOProvider(req.SSOProvider),
			LoginURL: req.SSOLoginURL,
		},
		SessionID: req.SessionID,
		Cookies:   cookies,
	}
}

func (h *TaskHandler) convertLLMConfig(req *LLMConfigRequest) *domain.LLMConfig {
	if req == nil {
		return nil
	}
	return &domain.LLMConfig{
		Provider: domain.LLMProvider(req.Provider),
		Model:    req.Model,
		Endpoint: req.Endpoint,
		APIKey:   req.APIKey,
		Options: &domain.LLMOptions{
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
		},
	}
}

func (h *TaskHandler) convertOutputConfig(req *OutputConfigRequest) *domain.OutputConfig {
	if req == nil {
		return domain.DefaultOutputConfig()
	}
	
	formats := make([]domain.DocFormat, len(req.Formats))
	for i, f := range req.Formats {
		formats[i] = domain.DocFormat(f)
	}
	
	return &domain.OutputConfig{
		Formats:  formats,
		Language: req.Language,
		Title:    req.Title,
		ScreenshotConfig: &domain.ScreenshotConf{
			Quality:  req.ScreenshotQuality,
			Annotate: req.Annotate,
		},
		StyleConfig: &domain.StyleConfig{
			Template:   req.Template,
			LogoURL:    req.LogoURL,
			ThemeColor: req.ThemeColor,
		},
		ContentConfig: &domain.ContentConfig{
			IncludeTOC:   req.IncludeTOC,
			IncludeCover: req.IncludeCover,
		},
	}
}
