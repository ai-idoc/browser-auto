// Package browser 提供浏览器控制功能
package browser

import (
	"context"
	"time"

	"github.com/browser-automation/internal/domain"
)

// Controller 浏览器控制器接口
type Controller interface {
	// 生命周期
	Connect(ctx context.Context) error
	Close(ctx context.Context) error

	// 导航
	Navigate(ctx context.Context, url string) error
	GetCurrentURL(ctx context.Context) (string, error)
	WaitForNavigation(ctx context.Context, timeout time.Duration) error
	WaitForURL(ctx context.Context, urlPattern string, timeout time.Duration) error

	// 元素操作
	Click(ctx context.Context, selector string) error
	Fill(ctx context.Context, selector string, value string) error
	Hover(ctx context.Context, selector string) error
	Select(ctx context.Context, selector string, value string) error

	// 等待
	WaitForSelector(ctx context.Context, selector string, timeout time.Duration) error
	WaitForText(ctx context.Context, text string, timeout time.Duration) error

	// 页面分析
	TakeSnapshot(ctx context.Context) (*PageSnapshot, error)
	TakeScreenshot(ctx context.Context, opts ScreenshotOptions) ([]byte, error)
	GetPageTitle(ctx context.Context) (string, error)

	// Cookie 管理
	GetCookies(ctx context.Context) ([]domain.Cookie, error)
	SetCookies(ctx context.Context, cookies []domain.Cookie) error
	ClearCookies(ctx context.Context) error
}

// PageSnapshot 页面快照
type PageSnapshot struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	HTML      string    `json:"html"`
	A11yTree  string    `json:"a11y_tree"`  // 无障碍树（供 AI 分析）
	Elements  []Element `json:"elements"`   // 可交互元素
	Timestamp time.Time `json:"timestamp"`
}

// Element 页面元素
type Element struct {
	TagName    string            `json:"tag_name"`
	Selector   string            `json:"selector"`
	Text       string            `json:"text"`
	Attributes map[string]string `json:"attributes"`
	Rect       *Rect             `json:"rect"`
	Visible    bool              `json:"visible"`
	Clickable  bool              `json:"clickable"`
}

// Rect 元素位置
type Rect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ScreenshotOptions 截图选项
type ScreenshotOptions struct {
	FullPage bool   `json:"full_page"`
	Quality  int    `json:"quality"` // 1-100
	Type     string `json:"type"`    // png, jpeg
	Clip     *Rect  `json:"clip,omitempty"`
}

// ActionType 操作类型
type ActionType string

const (
	ActionNavigate   ActionType = "navigate"
	ActionClick      ActionType = "click"
	ActionFill       ActionType = "fill"
	ActionHover      ActionType = "hover"
	ActionSelect     ActionType = "select"
	ActionScreenshot ActionType = "screenshot"
	ActionWait       ActionType = "wait"
	ActionScroll     ActionType = "scroll"
)

// Action 浏览器操作
type Action struct {
	Type        ActionType `json:"type"`
	Target      string     `json:"target"`      // URL 或选择器
	Value       string     `json:"value"`       // 输入值
	Description string     `json:"description"` // 操作描述
	Screenshot  bool       `json:"screenshot"`  // 是否截图
	WaitAfter   int        `json:"wait_after"`  // 操作后等待毫秒
}
