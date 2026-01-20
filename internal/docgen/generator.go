// Package docgen 提供文档生成功能
package docgen

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/browser-automation/internal/domain"
	"github.com/browser-automation/internal/planner"
)

// Generator 文档生成器接口
type Generator interface {
	Generate(ctx context.Context, task *domain.Task, plan *planner.TaskPlan, results []planner.StepResult) (*Document, error)
}

// Document 生成的文档
type Document struct {
	Title     string
	Content   string
	Format    domain.DocFormat
	CreatedAt time.Time
}

// MarkdownGenerator Markdown 文档生成器
type MarkdownGenerator struct{}

// NewMarkdownGenerator 创建 Markdown 生成器
func NewMarkdownGenerator() *MarkdownGenerator {
	return &MarkdownGenerator{}
}

// Generate 生成 Markdown 文档
func (g *MarkdownGenerator) Generate(ctx context.Context, task *domain.Task, plan *planner.TaskPlan, results []planner.StepResult) (*Document, error) {
	var buf bytes.Buffer
	
	// 标题
	title := task.Output.Title
	if title == "" {
		title = plan.Description
	}
	
	buf.WriteString(fmt.Sprintf("# %s\n\n", title))
	
	// 目录（如果启用）
	if task.Output.ContentConfig != nil && task.Output.ContentConfig.IncludeTOC {
		buf.WriteString("## 目录\n\n")
		for i, step := range plan.Steps {
			buf.WriteString(fmt.Sprintf("%d. [%s](#步骤-%d)\n", i+1, step.Description, i+1))
		}
		buf.WriteString("\n---\n\n")
	}
	
	// 概述
	buf.WriteString("## 概述\n\n")
	buf.WriteString(fmt.Sprintf("本指南将演示如何在 [%s](%s) 上完成以下操作：\n\n", task.TargetURL, task.TargetURL))
	buf.WriteString(fmt.Sprintf("> %s\n\n", task.Description))
	
	// 步骤
	buf.WriteString("## 操作步骤\n\n")
	
	for i, step := range plan.Steps {
		result := getStepResult(results, i)
		
		// 步骤标题
		stepNum := formatStepNumber(i+1, task.Output.ContentConfig)
		buf.WriteString(fmt.Sprintf("### 步骤 %s：%s\n\n", stepNum, step.Description))
		
		// 步骤详情
		buf.WriteString(g.formatStepContent(step, result))
		
		// 截图占位符
		if step.Screenshot && result != nil && result.Success {
			buf.WriteString(fmt.Sprintf("\n![步骤 %s 截图](screenshots/step_%d.png)\n\n", stepNum, i+1))
		}
		
		// 提示（如果启用）
		if task.Output.ContentConfig != nil && task.Output.ContentConfig.IncludeTips {
			tips := g.generateTips(step)
			if len(tips) > 0 {
				buf.WriteString("\n> **提示**：")
				buf.WriteString(strings.Join(tips, " "))
				buf.WriteString("\n\n")
			}
		}
	}
	
	// 总结
	buf.WriteString("## 总结\n\n")
	buf.WriteString(fmt.Sprintf("通过以上 %d 个步骤，您已成功完成了「%s」操作。\n\n", len(plan.Steps), task.Description))
	
	// 生成时间
	buf.WriteString("---\n\n")
	buf.WriteString(fmt.Sprintf("*文档生成时间：%s*\n", time.Now().Format("2006-01-02 15:04:05")))
	
	return &Document{
		Title:     title,
		Content:   buf.String(),
		Format:    domain.DocFormatMarkdown,
		CreatedAt: time.Now(),
	}, nil
}

func (g *MarkdownGenerator) formatStepContent(step planner.ActionStep, result *planner.StepResult) string {
	var buf bytes.Buffer
	
	switch step.Action {
	case "navigate":
		buf.WriteString(fmt.Sprintf("打开网址：`%s`\n", step.Target))
	case "click":
		buf.WriteString(fmt.Sprintf("点击「%s」按钮/链接。\n", step.Description))
	case "fill":
		buf.WriteString(fmt.Sprintf("在输入框中填写：`%s`\n", step.Value))
	case "hover":
		buf.WriteString(fmt.Sprintf("将鼠标悬停在「%s」上。\n", step.Description))
	case "select":
		buf.WriteString(fmt.Sprintf("从下拉列表中选择「%s」。\n", step.Value))
	case "wait":
		buf.WriteString("等待页面加载完成。\n")
	default:
		buf.WriteString(step.Description + "\n")
	}
	
	return buf.String()
}

func (g *MarkdownGenerator) generateTips(step planner.ActionStep) []string {
	var tips []string
	
	switch step.Action {
	case "fill":
		tips = append(tips, "请确保输入的信息准确无误。")
	case "click":
		if strings.Contains(strings.ToLower(step.Description), "提交") ||
			strings.Contains(strings.ToLower(step.Description), "确认") {
			tips = append(tips, "提交前请仔细检查填写的内容。")
		}
	}
	
	return tips
}

// HTMLGenerator HTML 文档生成器
type HTMLGenerator struct {
	template *template.Template
}

// NewHTMLGenerator 创建 HTML 生成器
func NewHTMLGenerator() *HTMLGenerator {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}
	tmpl := template.Must(template.New("doc").Funcs(funcMap).Parse(htmlTemplate))
	return &HTMLGenerator{template: tmpl}
}

// Generate 生成 HTML 文档
func (g *HTMLGenerator) Generate(ctx context.Context, task *domain.Task, plan *planner.TaskPlan, results []planner.StepResult) (*Document, error) {
	title := task.Output.Title
	if title == "" {
		title = plan.Description
	}
	
	// 默认主题色
	themeColor := "#3B82F6"
	if task.Output.StyleConfig != nil && task.Output.StyleConfig.ThemeColor != "" {
		themeColor = task.Output.StyleConfig.ThemeColor
	}
	
	data := map[string]interface{}{
		"Title":       title,
		"Description": task.Description,
		"TargetURL":   task.TargetURL,
		"Steps":       plan.Steps,
		"Results":     results,
		"ThemeColor":  themeColor,
		"GeneratedAt": time.Now().Format("2006-01-02 15:04:05"),
	}
	
	var buf bytes.Buffer
	if err := g.template.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}
	
	return &Document{
		Title:     title,
		Content:   buf.String(),
		Format:    domain.DocFormatHTML,
		CreatedAt: time.Now(),
	}, nil
}

func getStepResult(results []planner.StepResult, index int) *planner.StepResult {
	if index < len(results) {
		return &results[index]
	}
	return nil
}

func formatStepNumber(num int, config *domain.ContentConfig) string {
	if config == nil {
		return fmt.Sprintf("%d", num)
	}
	
	switch config.StepNumbering {
	case "letter":
		return string(rune('A' + num - 1))
	case "none":
		return ""
	default:
		return fmt.Sprintf("%d", num)
	}
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #1e293b;
            background: #f8fafc;
            padding: 2rem;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            border-radius: 16px;
            box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1);
            padding: 2rem;
        }
        h1 {
            color: {{.ThemeColor}};
            margin-bottom: 1rem;
            font-size: 2rem;
        }
        .description {
            background: #f1f5f9;
            padding: 1rem;
            border-radius: 8px;
            margin-bottom: 2rem;
        }
        .step {
            border-left: 3px solid {{.ThemeColor}};
            padding-left: 1.5rem;
            margin-bottom: 1.5rem;
        }
        .step-number {
            display: inline-block;
            width: 28px;
            height: 28px;
            background: {{.ThemeColor}};
            color: white;
            border-radius: 50%;
            text-align: center;
            line-height: 28px;
            font-weight: bold;
            margin-right: 0.5rem;
        }
        .step h3 {
            display: inline;
            font-size: 1.1rem;
        }
        .step p {
            margin-top: 0.5rem;
            color: #64748b;
        }
        .step img {
            max-width: 100%;
            border-radius: 8px;
            margin-top: 1rem;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .footer {
            margin-top: 2rem;
            padding-top: 1rem;
            border-top: 1px solid #e2e8f0;
            color: #94a3b8;
            font-size: 0.875rem;
            text-align: center;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>{{.Title}}</h1>
        <div class="description">
            <p>{{.Description}}</p>
            <p><small>目标网站：<a href="{{.TargetURL}}">{{.TargetURL}}</a></small></p>
        </div>
        
        <h2>操作步骤</h2>
        {{range $i, $step := .Steps}}
        <div class="step">
            <span class="step-number">{{add $i 1}}</span>
            <h3>{{$step.Description}}</h3>
            {{if $step.Screenshot}}
            <img src="screenshots/step_{{add $i 1}}.png" alt="步骤 {{add $i 1}} 截图">
            {{end}}
        </div>
        {{end}}
        
        <div class="footer">
            文档生成时间：{{.GeneratedAt}}
        </div>
    </div>
</body>
</html>`

func init() {
	// 模板函数已在 NewHTMLGenerator 中注册
}
