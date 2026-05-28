package service

import (
	"fmt"
	"strings"
)

// MessageTemplate 消息模板
type MessageTemplate struct {
	Name        string
	Description string
	Format      string
}

// 预定义的消息模板
var Templates = []MessageTemplate{
	{
		Name:        "code_response",
		Description: "代码回复模板",
		Format: "**代码**:\n```\n{{code}}\n```\n\n**解释**: {{explanation}}\n\n**使用建议**: {{suggestion}}",
	},
	{
		Name:        "simple_response",
		Description: "简单回复模板",
		Format: "**回答**: {{answer}}\n\n**建议**: {{suggestion}}\n\n**后续步骤**: {{next_steps}}",
	},
	{
		Name:        "error_response",
		Description: "错误响应模板",
		Format: "**错误**: {{error}}\n\n**原因**: {{reason}}\n\n**解决方案**: {{solution}}",
	},
	{
		Name:        "list_response",
		Description: "列表响应模板",
		Format: "**列表**:\n\n{{items}}\n\n**总结**: {{summary}}",
	},
}

// TemplateService 模板服务
type TemplateService struct{}

// NewTemplateService 创建模板服务
func NewTemplateService() *TemplateService {
	return &TemplateService{}
}

// ApplyTemplate 应用模板
func (ts *TemplateService) ApplyTemplate(templateName string, data map[string]string) (string, error) {
	var template *MessageTemplate
	for i := range Templates {
		if Templates[i].Name == templateName {
			template = &Templates[i]
			break
		}
	}

	if template == nil {
		return "", fmt.Errorf("template '%s' not found", templateName)
	}

	result := template.Format
	for key, value := range data {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

// ListTemplates 列出所有模板
func (ts *TemplateService) ListTemplates() []MessageTemplate {
	return Templates
}

// AutoFormatResponse 自动格式化响应
func (ts *TemplateService) AutoFormatResponse(content string) string {
	// 如果内容已经包含格式符号，保持原样
	if strings.Contains(content, "```") ||
		strings.Contains(content, "**") ||
		strings.Contains(content, "📝") ||
		strings.Contains(content, "✅") ||
		strings.Contains(content, "❌") {
		return content
	}

	// 如果是多行代码，添加代码块
	if strings.Contains(content, "\n") && !strings.Contains(content, "。") {
		return fmt.Sprintf("📝 **代码**:\n```\n%s\n```", content)
	}

	// 如果是列表，添加列表格式
	if strings.Count(content, "\n") > 2 {
		lines := strings.Split(content, "\n")
		formatted := "📋 **列表**:\n\n"
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				formatted += "• " + strings.TrimSpace(line) + "\n"
			}
		}
		return formatted
	}

	// 默认返回带格式的简单回复
	return fmt.Sprintf("✅ **回答**: %s", content)
}
