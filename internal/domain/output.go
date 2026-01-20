// Package domain 定义核心业务模型
package domain

// DocFormat 文档格式
type DocFormat string

const (
	DocFormatMarkdown DocFormat = "markdown"
	DocFormatHTML     DocFormat = "html"
	DocFormatPDF      DocFormat = "pdf"
	DocFormatDOCX     DocFormat = "docx"
)

// OutputConfig 文档输出配置
type OutputConfig struct {
	Formats          []DocFormat     `json:"formats"`           // 输出格式（支持多选）
	Language         string          `json:"language"`          // 文档语言: zh, en
	Title            string          `json:"title"`             // 文档标题
	ScreenshotConfig *ScreenshotConf `json:"screenshot_config"` // 截图配置
	StyleConfig      *StyleConfig    `json:"style_config"`      // 样式配置
	ContentConfig    *ContentConfig  `json:"content_config"`    // 内容配置
}

// ScreenshotConf 截图配置
type ScreenshotConf struct {
	Quality    int  `json:"quality"`     // 截图质量 1-100
	Annotate   bool `json:"annotate"`    // 是否标注操作位置
	FullPage   bool `json:"full_page"`   // 是否全页截图
	HighlightColor string `json:"highlight_color"` // 标注颜色
}

// StyleConfig 样式配置
type StyleConfig struct {
	Template   string `json:"template"`    // 模板: simple, professional, custom
	LogoURL    string `json:"logo_url"`    // Logo 地址
	ThemeColor string `json:"theme_color"` // 主题色
}

// ContentConfig 内容配置
type ContentConfig struct {
	IncludeTOC     bool   `json:"include_toc"`     // 是否包含目录
	IncludeCover   bool   `json:"include_cover"`   // 是否包含封面
	StepNumbering  string `json:"step_numbering"`  // 步骤编号: number, letter, none
	IncludeTips    bool   `json:"include_tips"`    // 是否包含提示信息
}

// DefaultOutputConfig 默认输出配置
func DefaultOutputConfig() *OutputConfig {
	return &OutputConfig{
		Formats:  []DocFormat{DocFormatMarkdown},
		Language: "zh",
		Title:    "",
		ScreenshotConfig: &ScreenshotConf{
			Quality:        90,
			Annotate:       true,
			FullPage:       false,
			HighlightColor: "#FF0000",
		},
		StyleConfig: &StyleConfig{
			Template:   "simple",
			ThemeColor: "#3B82F6",
		},
		ContentConfig: &ContentConfig{
			IncludeTOC:    true,
			IncludeCover:  false,
			StepNumbering: "number",
			IncludeTips:   true,
		},
	}
}

// GetSupportedFormats 获取支持的输出格式
func GetSupportedFormats() []FormatInfo {
	return []FormatInfo{
		{
			Format:      DocFormatMarkdown,
			Name:        "Markdown",
			Description: "轻量级标记语言，适合技术文档和 Git 仓库",
			Extension:   ".md",
			Icon:        "file-text",
		},
		{
			Format:      DocFormatHTML,
			Name:        "HTML",
			Description: "网页格式，可直接在浏览器中查看",
			Extension:   ".html",
			Icon:        "globe",
		},
		{
			Format:      DocFormatPDF,
			Name:        "PDF",
			Description: "便携文档格式，适合打印和正式分发",
			Extension:   ".pdf",
			Icon:        "file",
		},
		{
			Format:      DocFormatDOCX,
			Name:        "Word (DOCX)",
			Description: "Microsoft Word 格式，便于二次编辑",
			Extension:   ".docx",
			Icon:        "file-word",
		},
	}
}

// FormatInfo 格式信息
type FormatInfo struct {
	Format      DocFormat `json:"format"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Extension   string    `json:"extension"`
	Icon        string    `json:"icon"`
}
