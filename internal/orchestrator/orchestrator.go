// Package orchestrator 提供任务编排功能
package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/browser-automation/internal/auth"
	"github.com/browser-automation/internal/browser"
	"github.com/browser-automation/internal/docgen"
	"github.com/browser-automation/internal/domain"
	"github.com/browser-automation/internal/planner"
	"github.com/browser-automation/internal/storage"
	"github.com/google/uuid"
)

// Orchestrator 任务编排器
type Orchestrator struct {
	browserCtrl browser.Controller
	authService *auth.Service
	planner     planner.Planner
	docGen      docgen.Generator
	taskStore   storage.TaskStore
	llmFactory  *planner.LLMClientFactory
}

// NewOrchestrator 创建任务编排器
func NewOrchestrator(
	browserCtrl browser.Controller,
	taskStore storage.TaskStore,
	llmFactory *planner.LLMClientFactory,
) *Orchestrator {
	return &Orchestrator{
		browserCtrl: browserCtrl,
		authService: auth.NewService(browserCtrl),
		taskStore:   taskStore,
		llmFactory:  llmFactory,
	}
}

// ExecuteTask 执行任务
func (o *Orchestrator) ExecuteTask(ctx context.Context, task *domain.Task) error {
	log.Printf("[Task %s] Starting execution", task.ID)

	// 更新任务状态为运行中
	task.Status = domain.TaskStatusRunning
	task.UpdatedAt = time.Now()
	if err := o.taskStore.Update(ctx, task); err != nil {
		return fmt.Errorf("update task status: %w", err)
	}

	startTime := time.Now()

	// 创建 LLM 客户端
	log.Printf("[Task %s] Creating LLM client: provider=%s, model=%s", task.ID, task.LLM.Provider, task.LLM.Model)
	llmClient, err := o.llmFactory.NewClient(task.LLM)
	if err != nil {
		return o.failTask(ctx, task, fmt.Errorf("create llm client: %w", err))
	}

	// 创建 AI 规划器
	aiPlanner := planner.NewAIPlanner(llmClient)

	// 连接浏览器
	log.Printf("[Task %s] Connecting browser", task.ID)
	if err := o.browserCtrl.Connect(ctx); err != nil {
		return o.failTask(ctx, task, fmt.Errorf("connect browser: %w", err))
	}
	defer o.browserCtrl.Close(ctx)

	// 处理认证
	if task.Auth != nil && task.Auth.Type != domain.AuthTypeNone {
		log.Printf("[Task %s] Processing authentication: type=%s", task.ID, task.Auth.Type)
		// 先导航到目标页面
		if err := o.browserCtrl.Navigate(ctx, task.TargetURL); err != nil {
			return o.failTask(ctx, task, fmt.Errorf("navigate for auth: %w", err))
		}

		// 执行认证
		session, err := o.authService.Authenticate(ctx, task.Auth)
		if err != nil {
			return o.failTask(ctx, task, fmt.Errorf("authenticate: %w", err))
		}

		// 注入会话
		if len(session.Cookies) > 0 {
			if err := o.browserCtrl.SetCookies(ctx, session.Cookies); err != nil {
				return o.failTask(ctx, task, fmt.Errorf("set cookies: %w", err))
			}
		}

		// 刷新页面应用认证
		if err := o.browserCtrl.Navigate(ctx, task.TargetURL); err != nil {
			return o.failTask(ctx, task, fmt.Errorf("navigate after auth: %w", err))
		}
	} else {
		// 直接导航到目标页面
		log.Printf("[Task %s] Navigating to: %s", task.ID, task.TargetURL)
		if err := o.browserCtrl.Navigate(ctx, task.TargetURL); err != nil {
			return o.failTask(ctx, task, fmt.Errorf("navigate: %w", err))
		}
	}

	// 等待页面加载
	log.Printf("[Task %s] Waiting for page load (2s)", task.ID)
	time.Sleep(2 * time.Second)

	// 获取页面快照
	log.Printf("[Task %s] Taking page snapshot", task.ID)
	snapshot, err := o.browserCtrl.TakeSnapshot(ctx)
	if err != nil {
		return o.failTask(ctx, task, fmt.Errorf("take snapshot: %w", err))
	}
	log.Printf("[Task %s] Snapshot: URL=%s, Title=%s, Elements=%d", task.ID, snapshot.URL, snapshot.Title, len(snapshot.Elements))

	// AI 解析任务生成计划
	log.Printf("[Task %s] Calling LLM to parse task...", task.ID)
	plan, err := aiPlanner.ParseTask(ctx, &planner.PlanRequest{
		UserInput:    task.Description,
		TargetURL:    task.TargetURL,
		PageSnapshot: snapshot,
	})
	if err != nil {
		log.Printf("[Task %s] LLM parse failed: %v", task.ID, err)
		return o.failTask(ctx, task, fmt.Errorf("parse task: %w", err))
	}
	log.Printf("[Task %s] LLM returned %d steps", task.ID, len(plan.Steps))

	// 执行步骤
	var stepResults []planner.StepResult
	var screenshots []domain.Screenshot

	for i, step := range plan.Steps {
		log.Printf("[Task %s] Executing step %d/%d: %s", task.ID, i+1, len(plan.Steps), step.Description)
		result, screenshot, err := o.executeStep(ctx, step)
		if err != nil {
			log.Printf("[Task %s] Step %d failed: %v, attempting refine...", task.ID, i+1, err)
			// 尝试重新规划
			refined, refineErr := aiPlanner.RefineStep(ctx, &step, snapshot)
			if refineErr != nil {
				log.Printf("[Task %s] Refine failed: %v", task.ID, refineErr)
				stepResults = append(stepResults, planner.StepResult{
					Success: false,
					Error:   err.Error(),
				})
				continue
			}
			log.Printf("[Task %s] Refined step: %s -> %s", task.ID, step.Target, refined.Target)
			// 重新执行
			result, screenshot, _ = o.executeStep(ctx, *refined)
		}

		stepResults = append(stepResults, *result)
		if screenshot != nil {
			screenshots = append(screenshots, *screenshot)
		}

		// 更新快照
		snapshot, _ = o.browserCtrl.TakeSnapshot(ctx)
	}

	// 生成文档
	docs, err := o.generateDocuments(ctx, task, plan, stepResults)
	if err != nil {
		return o.failTask(ctx, task, fmt.Errorf("generate docs: %w", err))
	}

	// 更新任务结果
	task.Status = domain.TaskStatusCompleted
	task.UpdatedAt = time.Now()
	completedAt := time.Now()
	task.CompletedAt = &completedAt
	task.Result = &domain.TaskResult{
		Steps:       convertStepResults(stepResults),
		Screenshots: screenshots,
		Documents:   docs,
		Duration:    time.Since(startTime),
	}

	if err := o.taskStore.Update(ctx, task); err != nil {
		return fmt.Errorf("update task result: %w", err)
	}

	return nil
}

func (o *Orchestrator) executeStep(ctx context.Context, step planner.ActionStep) (*planner.StepResult, *domain.Screenshot, error) {
	var err error

	log.Printf("[Step] Executing action=%s, target=%s, value=%s", step.Action, step.Target, step.Value)

	switch step.Action {
	case browser.ActionNavigate:
		log.Printf("[Step] Navigate to: %s", step.Target)
		err = o.browserCtrl.Navigate(ctx, step.Target)
	case browser.ActionClick:
		log.Printf("[Step] Click on: %s", step.Target)
		err = o.browserCtrl.Click(ctx, step.Target)
	case browser.ActionFill:
		log.Printf("[Step] Fill %s with: %s", step.Target, step.Value)
		err = o.browserCtrl.Fill(ctx, step.Target, step.Value)
	case browser.ActionHover:
		log.Printf("[Step] Hover on: %s", step.Target)
		err = o.browserCtrl.Hover(ctx, step.Target)
	case browser.ActionSelect:
		log.Printf("[Step] Select %s in: %s", step.Value, step.Target)
		err = o.browserCtrl.Select(ctx, step.Target, step.Value)
	case browser.ActionWait:
		if step.WaitFor != "" {
			err = o.browserCtrl.WaitForSelector(ctx, step.WaitFor, 10*time.Second)
		} else {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		return &planner.StepResult{Success: false, Error: err.Error()}, nil, err
	}

	// 等待动作完成
	time.Sleep(500 * time.Millisecond)

	// 截图
	var screenshot *domain.Screenshot
	if step.Screenshot {
		imgData, err := o.browserCtrl.TakeScreenshot(ctx, browser.ScreenshotOptions{
			Quality: 90,
			Type:    "png",
		})
		if err == nil {
			screenshot = &domain.Screenshot{
				ID:        uuid.New().String(),
				StepOrder: step.Order,
				CreatedAt: time.Now(),
			}
			// TODO: 保存截图到存储
			_ = imgData
		}
	}

	return &planner.StepResult{Success: true}, screenshot, nil
}

func (o *Orchestrator) generateDocuments(ctx context.Context, task *domain.Task, plan *planner.TaskPlan, results []planner.StepResult) ([]domain.DocumentInfo, error) {
	var docs []domain.DocumentInfo

	for _, format := range task.Output.Formats {
		var gen docgen.Generator
		switch format {
		case domain.DocFormatMarkdown:
			gen = docgen.NewMarkdownGenerator()
		case domain.DocFormatHTML:
			gen = docgen.NewHTMLGenerator()
		default:
			continue // 暂不支持的格式
		}

		doc, err := gen.Generate(ctx, task, plan, results)
		if err != nil {
			continue
		}

		// 保存文档内容
		docs = append(docs, domain.DocumentInfo{
			ID:        uuid.New().String(),
			Format:    format,
			Content:   doc.Content,
			Size:      int64(len(doc.Content)),
			CreatedAt: time.Now(),
		})
	}

	return docs, nil
}

func (o *Orchestrator) failTask(ctx context.Context, task *domain.Task, err error) error {
	task.Status = domain.TaskStatusFailed
	task.ErrorMessage = err.Error()
	task.UpdatedAt = time.Now()
	o.taskStore.Update(ctx, task)
	return err
}

func convertStepResults(results []planner.StepResult) []domain.StepResult {
	var domainResults []domain.StepResult
	for i, r := range results {
		domainResults = append(domainResults, domain.StepResult{
			Order:      i + 1,
			Success:    r.Success,
			Error:      r.Error,
			ExecutedAt: time.Now(),
		})
	}
	return domainResults
}
