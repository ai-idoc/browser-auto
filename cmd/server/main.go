// Package main 浏览器自动化服务入口
package main

import (
	"log"

	"github.com/browser-automation/internal/api"
	"github.com/browser-automation/internal/browser"
	"github.com/browser-automation/internal/orchestrator"
	"github.com/browser-automation/internal/planner"
	"github.com/browser-automation/internal/storage"
)

func main() {
	// 初始化存储
	taskStore := storage.NewMemoryTaskStore()

	// 初始化 LLM 工厂
	llmFactory := planner.NewLLMClientFactory()

	// 初始化浏览器控制器（非 headless 模式方便观察）
	browserCtrl := browser.NewPlaywrightController(browser.PlaywrightOptions{
		Headless: false, // 设为 false 可以看到浏览器操作
	})

	// 初始化编排器
	orch := orchestrator.NewOrchestrator(browserCtrl, taskStore, llmFactory)

	// 设置路由
	r := api.SetupRouter(taskStore, llmFactory, orch)

	// 启动服务
	log.Println("Server starting on port 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
