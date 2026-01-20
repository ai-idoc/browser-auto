// Package browser 提供浏览器控制功能
package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/browser-automation/internal/domain"
	"github.com/playwright-community/playwright-go"
)

// PlaywrightController Playwright 浏览器控制器
type PlaywrightController struct {
	pw       *playwright.Playwright
	browser  playwright.Browser
	page     playwright.Page
	headless bool
	wsURL    string
}

// PlaywrightOptions Playwright 选项
type PlaywrightOptions struct {
	Headless  bool
	WSEndpoint string
}

// NewPlaywrightController 创建 Playwright 控制器
func NewPlaywrightController(opts PlaywrightOptions) *PlaywrightController {
	return &PlaywrightController{
		headless: opts.Headless,
		wsURL:    opts.WSEndpoint,
	}
}

// Connect 连接浏览器
func (c *PlaywrightController) Connect(ctx context.Context) error {
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("start playwright: %w", err)
	}
	c.pw = pw

	var browser playwright.Browser
	if c.wsURL != "" {
		// 连接远程浏览器
		browser, err = pw.Chromium.Connect(c.wsURL)
	} else {
		// 启动本地浏览器
		browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(c.headless),
		})
	}
	if err != nil {
		return fmt.Errorf("launch browser: %w", err)
	}
	c.browser = browser

	page, err := browser.NewPage()
	if err != nil {
		return fmt.Errorf("new page: %w", err)
	}
	c.page = page

	return nil
}

// Close 关闭浏览器
func (c *PlaywrightController) Close(ctx context.Context) error {
	if c.page != nil {
		c.page.Close()
	}
	if c.browser != nil {
		c.browser.Close()
	}
	if c.pw != nil {
		c.pw.Stop()
	}
	return nil
}

// Navigate 导航到 URL
func (c *PlaywrightController) Navigate(ctx context.Context, url string) error {
	_, err := c.page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	return err
}

// GetCurrentURL 获取当前 URL
func (c *PlaywrightController) GetCurrentURL(ctx context.Context) (string, error) {
	return c.page.URL(), nil
}

// WaitForNavigation 等待导航完成
func (c *PlaywrightController) WaitForNavigation(ctx context.Context, timeout time.Duration) error {
	return c.page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
}

// WaitForURL 等待 URL 匹配
func (c *PlaywrightController) WaitForURL(ctx context.Context, urlPattern string, timeout time.Duration) error {
	return c.page.WaitForURL(urlPattern, playwright.PageWaitForURLOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
}

// Click 点击元素
func (c *PlaywrightController) Click(ctx context.Context, selector string) error {
	return c.page.Click(selector)
}

// Fill 填写输入框
func (c *PlaywrightController) Fill(ctx context.Context, selector string, value string) error {
	return c.page.Fill(selector, value)
}

// Hover 悬停元素
func (c *PlaywrightController) Hover(ctx context.Context, selector string) error {
	return c.page.Hover(selector)
}

// Select 选择下拉选项
func (c *PlaywrightController) Select(ctx context.Context, selector string, value string) error {
	_, err := c.page.SelectOption(selector, playwright.SelectOptionValues{
		Values: playwright.StringSlice(value),
	})
	return err
}

// WaitForSelector 等待选择器出现
func (c *PlaywrightController) WaitForSelector(ctx context.Context, selector string, timeout time.Duration) error {
	_, err := c.page.WaitForSelector(selector, playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	return err
}

// WaitForText 等待文本出现
func (c *PlaywrightController) WaitForText(ctx context.Context, text string, timeout time.Duration) error {
	_, err := c.page.WaitForSelector(fmt.Sprintf("text=%s", text), playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(float64(timeout.Milliseconds())),
	})
	return err
}

// TakeSnapshot 获取页面快照
func (c *PlaywrightController) TakeSnapshot(ctx context.Context) (*PageSnapshot, error) {
	url := c.page.URL()
	title, _ := c.page.Title()
	
	// 使用 JavaScript 直接获取页面信息，避免多次 IPC 调用
	result, err := c.page.Evaluate(`() => {
		const elements = [];
		const selectors = 'a, button, input, select, textarea, [role="button"], [onclick]';
		const els = document.querySelectorAll(selectors);
		const max = Math.min(els.length, 30);
		for (let i = 0; i < max; i++) {
			const el = els[i];
			if (!el.offsetParent) continue; // 跳过不可见元素
			const text = (el.innerText || el.value || el.placeholder || '').slice(0, 50);
			elements.push({
				tagName: el.tagName.toLowerCase(),
				text: text,
				id: el.id || '',
				name: el.name || '',
				type: el.type || ''
			});
		}
		return {
			elements: elements,
			elementCount: els.length
		};
	}`)
	if err != nil {
		result = map[string]interface{}{"elements": []interface{}{}, "elementCount": 0}
	}
	
	// 解析结果
	var elements []Element
	if resultMap, ok := result.(map[string]interface{}); ok {
		if elsRaw, ok := resultMap["elements"].([]interface{}); ok {
			for _, elRaw := range elsRaw {
				if el, ok := elRaw.(map[string]interface{}); ok {
					elements = append(elements, Element{
						TagName:   fmt.Sprintf("%v", el["tagName"]),
						Text:      fmt.Sprintf("%v", el["text"]),
						Visible:   true,
						Clickable: true,
					})
				}
			}
		}
	}

	return &PageSnapshot{
		URL:       url,
		Title:     title,
		Elements:  elements,
		Timestamp: time.Now(),
	}, nil
}

// TakeScreenshot 截图
func (c *PlaywrightController) TakeScreenshot(ctx context.Context, opts ScreenshotOptions) ([]byte, error) {
	screenshotOpts := playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(opts.FullPage),
	}
	
	if opts.Type == "jpeg" {
		screenshotOpts.Type = playwright.ScreenshotTypeJpeg
		if opts.Quality > 0 {
			screenshotOpts.Quality = playwright.Int(opts.Quality)
		}
	}

	return c.page.Screenshot(screenshotOpts)
}

// GetPageTitle 获取页面标题
func (c *PlaywrightController) GetPageTitle(ctx context.Context) (string, error) {
	return c.page.Title()
}

// GetCookies 获取 Cookies
func (c *PlaywrightController) GetCookies(ctx context.Context) ([]domain.Cookie, error) {
	cookies, err := c.page.Context().Cookies()
	if err != nil {
		return nil, err
	}

	result := make([]domain.Cookie, len(cookies))
	for i, c := range cookies {
		result[i] = domain.Cookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HttpOnly,
		}
	}
	return result, nil
}

// SetCookies 设置 Cookies
func (c *PlaywrightController) SetCookies(ctx context.Context, cookies []domain.Cookie) error {
	pwCookies := make([]playwright.OptionalCookie, len(cookies))
	for i, cookie := range cookies {
		pwCookies[i] = playwright.OptionalCookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Domain:   playwright.String(cookie.Domain),
			Path:     playwright.String(cookie.Path),
			Secure:   playwright.Bool(cookie.Secure),
			HttpOnly: playwright.Bool(cookie.HTTPOnly),
		}
	}
	return c.page.Context().AddCookies(pwCookies)
}

// ClearCookies 清除 Cookies
func (c *PlaywrightController) ClearCookies(ctx context.Context) error {
	return c.page.Context().ClearCookies()
}

// getAccessibilityTree 获取无障碍树
func (c *PlaywrightController) getAccessibilityTree(ctx context.Context) (string, error) {
	// 使用 JavaScript 获取页面结构信息
	result, err := c.page.Evaluate(`() => {
		function getTree(node, indent) {
			if (!node) return '';
			let result = '';
			const role = node.getAttribute && node.getAttribute('role') || node.tagName?.toLowerCase() || '';
			const text = node.innerText?.slice(0, 50) || '';
			if (role) {
				result += '  '.repeat(indent) + '[' + role + '] ' + text + '\\n';
			}
			for (const child of (node.children || [])) {
				result += getTree(child, indent + 1);
			}
			return result;
		}
		return getTree(document.body, 0);
	}`)
	if err != nil {
		return "", err
	}
	if result == nil {
		return "", nil
	}
	return fmt.Sprintf("%v", result), nil
}

// formatA11yNode 格式化无障碍节点（保留兼容）
func formatA11yNode(node interface{}, indent int) string {
	return ""
}

// getInteractiveElements 获取可交互元素
func (c *PlaywrightController) getInteractiveElements(ctx context.Context) ([]Element, error) {
	// 获取所有可交互元素
	selectors := "a, button, input, select, textarea, [role='button'], [onclick]"
	locators, err := c.page.Locator(selectors).All()
	if err != nil {
		return nil, err
	}

	// 限制元素数量，避免处理太长时间
	maxElements := 50
	elements := make([]Element, 0, maxElements)
	
	for i, loc := range locators {
		if i >= maxElements {
			break
		}
		
		tagName, _ := loc.Evaluate("el => el.tagName.toLowerCase()", nil)
		text, _ := loc.InnerText()
		visible, _ := loc.IsVisible()
		
		if !visible {
			continue
		}

		// 跳过获取 BoundingBox，太慢
		elements = append(elements, Element{
			TagName:   fmt.Sprintf("%v", tagName),
			Text:      text,
			Visible:   visible,
			Clickable: true,
		})
	}

	return elements, nil
}
