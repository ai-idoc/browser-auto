// Package api 提供 HTTP API 路由
package api

import (
	"github.com/browser-automation/internal/api/handler"
	"github.com/browser-automation/internal/orchestrator"
	"github.com/browser-automation/internal/planner"
	"github.com/browser-automation/internal/storage"
	"github.com/gin-gonic/gin"
)

// SetupRouter 设置路由
func SetupRouter(taskStore storage.TaskStore, llmFactory *planner.LLMClientFactory, orch *orchestrator.Orchestrator) *gin.Engine {
	r := gin.Default()

	// CORS 中间件
	r.Use(corsMiddleware())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// 任务相关
		taskHandler := handler.NewTaskHandler(taskStore, orch)
		tasks := v1.Group("/tasks")
		{
			tasks.POST("", taskHandler.CreateTask)
			tasks.GET("", taskHandler.ListTasks)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.POST("/:id/cancel", taskHandler.CancelTask)
		}

		// 配置相关
		configHandler := handler.NewConfigHandler(llmFactory)
		config := v1.Group("/config")
		{
			config.GET("/llm/presets", configHandler.GetLLMPresets)
			config.POST("/llm/validate", configHandler.ValidateLLM)
			config.GET("/output/formats", configHandler.GetOutputFormats)
			config.GET("/auth/types", configHandler.GetAuthTypes)
		}
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
