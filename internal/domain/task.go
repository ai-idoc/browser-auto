// Package domain 定义核心业务模型
package domain

import "time"

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusRunning    TaskStatus = "running"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Task 任务实体
type Task struct {
	ID           string        `json:"id"`
	Description  string        `json:"description"`   // 自然语言任务描述
	TargetURL    string        `json:"target_url"`    // 目标网站 URL
	Status       TaskStatus    `json:"status"`
	Auth         *AuthConfig   `json:"auth,omitempty"`
	LLM          *LLMConfig    `json:"llm"`
	Output       *OutputConfig `json:"output"`
	Result       *TaskResult   `json:"result,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	CompletedAt  *time.Time    `json:"completed_at,omitempty"`
}

// TaskResult 任务执行结果
type TaskResult struct {
	Steps       []StepResult   `json:"steps"`
	Screenshots []Screenshot   `json:"screenshots"`
	Documents   []DocumentInfo `json:"documents"`
	Duration    time.Duration  `json:"duration"`
}

// StepResult 步骤执行结果
type StepResult struct {
	Order        int        `json:"order"`
	Action       string     `json:"action"`
	Description  string     `json:"description"`
	Success      bool       `json:"success"`
	Error        string     `json:"error,omitempty"`
	Screenshot   *Screenshot `json:"screenshot,omitempty"`
	ExecutedAt   time.Time  `json:"executed_at"`
}

// Screenshot 截图信息
type Screenshot struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`       // 存储地址
	StepOrder int       `json:"step_order"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	CreatedAt time.Time `json:"created_at"`
}

// DocumentInfo 生成的文档信息
type DocumentInfo struct {
	ID        string     `json:"id"`
	Format    DocFormat  `json:"format"`
	URL       string     `json:"url,omitempty"`
	Content   string     `json:"content,omitempty"`
	Size      int64      `json:"size"`
	CreatedAt time.Time  `json:"created_at"`
}
