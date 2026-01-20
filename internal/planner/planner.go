// Package planner 提供 AI 规划功能
package planner

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/browser-automation/internal/browser"
)

// Planner AI 规划器接口
type Planner interface {
	ParseTask(ctx context.Context, req *PlanRequest) (*TaskPlan, error)
	RefineStep(ctx context.Context, step *ActionStep, snapshot *browser.PageSnapshot) (*ActionStep, error)
	GenerateStepDescription(ctx context.Context, step *ActionStep, result *StepResult) (string, error)
}

// PlanRequest 规划请求
type PlanRequest struct {
	UserInput    string                 `json:"user_input"`
	TargetURL    string                 `json:"target_url"`
	PageSnapshot *browser.PageSnapshot  `json:"page_snapshot"`
}

// TaskPlan 任务计划
type TaskPlan struct {
	TaskID      string       `json:"task_id"`
	Description string       `json:"description"`
	Steps       []ActionStep `json:"steps"`
}

// ActionStep 操作步骤
type ActionStep struct {
	Order       int                `json:"order"`
	Action      browser.ActionType `json:"action"`
	Target      string             `json:"target"`
	Value       string             `json:"value,omitempty"`
	WaitFor     string             `json:"wait_for,omitempty"`
	Screenshot  bool               `json:"screenshot"`
	Description string             `json:"description"`
}

// StepResult 步骤执行结果
type StepResult struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	Screenshot []byte `json:"screenshot,omitempty"`
}

// AIPlanner AI 规划器实现
type AIPlanner struct {
	llmClient LLMClient
}

// NewAIPlanner 创建 AI 规划器
func NewAIPlanner(llmClient LLMClient) *AIPlanner {
	return &AIPlanner{llmClient: llmClient}
}

// ParseTask 解析任务生成执行计划
func (p *AIPlanner) ParseTask(ctx context.Context, req *PlanRequest) (*TaskPlan, error) {
	prompt := p.buildTaskParsePrompt(req)
	
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}
	
	resp, err := p.llmClient.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm chat: %w", err)
	}
	
	// 解析 JSON 响应
	var plan TaskPlan
	if err := json.Unmarshal([]byte(resp.Content), &plan); err != nil {
		// 尝试提取 JSON
		jsonStr := extractJSON(resp.Content)
		if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
			return nil, fmt.Errorf("parse plan: %w", err)
		}
	}
	
	return &plan, nil
}

// RefineStep 根据页面状态优化步骤
func (p *AIPlanner) RefineStep(ctx context.Context, step *ActionStep, snapshot *browser.PageSnapshot) (*ActionStep, error) {
	prompt := fmt.Sprintf(`当前步骤执行失败，请根据页面状态优化选择器。

原步骤:
- 操作: %s
- 目标: %s
- 描述: %s

当前页面 URL: %s
页面标题: %s

可交互元素:
%s

请输出优化后的步骤 JSON。`, 
		step.Action, step.Target, step.Description,
		snapshot.URL, snapshot.Title,
		formatElements(snapshot.Elements))
	
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}
	
	resp, err := p.llmClient.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm chat: %w", err)
	}
	
	var refined ActionStep
	jsonStr := extractJSON(resp.Content)
	if err := json.Unmarshal([]byte(jsonStr), &refined); err != nil {
		return nil, fmt.Errorf("parse refined step: %w", err)
	}
	
	return &refined, nil
}

// GenerateStepDescription 生成步骤描述
func (p *AIPlanner) GenerateStepDescription(ctx context.Context, step *ActionStep, result *StepResult) (string, error) {
	prompt := fmt.Sprintf(`请为以下操作步骤生成用户友好的描述（用于帮助文档）：

操作: %s
目标: %s
值: %s
执行结果: %v

要求：
1. 使用简洁明了的语言
2. 面向普通用户，不要使用技术术语
3. 描述应该是指导性的，告诉用户如何操作

直接输出描述文本，不要包含其他内容。`,
		step.Action, step.Target, step.Value, result.Success)
	
	messages := []Message{
		{Role: "user", Content: prompt},
	}
	
	resp, err := p.llmClient.Chat(ctx, messages)
	if err != nil {
		return step.Description, nil // 降级使用原描述
	}
	
	return resp.Content, nil
}

func (p *AIPlanner) buildTaskParsePrompt(req *PlanRequest) string {
	pageInfo := ""
	if req.PageSnapshot != nil {
		pageInfo = fmt.Sprintf(`
当前页面 URL: %s
页面标题: %s

可交互元素:
%s`,
			req.PageSnapshot.URL,
			req.PageSnapshot.Title,
			formatElements(req.PageSnapshot.Elements))
	}
	
	return fmt.Sprintf(`## 用户任务
%s

## 目标网站
%s
%s

## 输出要求
生成 JSON 格式的操作步骤列表，格式如下：
{
  "task_id": "uuid",
  "description": "任务总体描述",
  "steps": [
    {
      "order": 1,
      "action": "navigate|click|fill|hover|screenshot|wait",
      "target": "CSS选择器或URL",
      "value": "输入值（如适用）",
      "wait_for": "等待条件（如适用）",
      "screenshot": true,
      "description": "用户友好的步骤说明"
    }
  ]
}

## 注意事项
1. 优先使用稳定的选择器（id > name > class > xpath）
2. 每个关键操作后添加截图（screenshot: true）
3. 步骤说明要清晰易懂，面向普通用户
4. 包含必要的等待步骤，确保页面加载完成

请输出 JSON：`, req.UserInput, req.TargetURL, pageInfo)
}

const systemPrompt = `你是一个浏览器自动化专家，负责将用户的自然语言任务描述转换为可执行的浏览器操作步骤。

你的输出必须是有效的 JSON 格式，包含以下字段：
- task_id: 任务唯一标识
- description: 任务描述
- steps: 操作步骤数组

每个步骤包含：
- order: 步骤序号
- action: 操作类型（navigate/click/fill/hover/screenshot/wait）
- target: 目标（URL 或 CSS 选择器）
- value: 输入值（可选）
- wait_for: 等待条件（可选）
- screenshot: 是否截图
- description: 步骤描述（用户友好）

确保生成的选择器是稳定可靠的，优先使用 id、name 属性。`

func formatElements(elements []browser.Element) string {
	if len(elements) == 0 {
		return "（无可交互元素）"
	}
	
	result := ""
	for i, el := range elements {
		if i >= 20 { // 限制数量
			result += fmt.Sprintf("... 还有 %d 个元素\n", len(elements)-20)
			break
		}
		result += fmt.Sprintf("- <%s> %s\n", el.TagName, el.Text)
	}
	return result
}

func extractJSON(content string) string {
	// 尝试提取 JSON 块
	start := -1
	end := -1
	depth := 0
	
	for i, c := range content {
		if c == '{' {
			if start == -1 {
				start = i
			}
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				end = i + 1
				break
			}
		}
	}
	
	if start != -1 && end != -1 {
		return content[start:end]
	}
	return content
}
