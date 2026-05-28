package service

import (
	"testing"
)

func TestTemplateService_NewTemplateService(t *testing.T) {
	ts := NewTemplateService()
	if ts == nil {
		t.Error("NewTemplateService should return non-nil")
	}
}

func TestTemplateService_ListTemplates(t *testing.T) {
	ts := NewTemplateService()
	templates := ts.ListTemplates()

	if len(templates) == 0 {
		t.Error("ListTemplates should return at least one template")
	}

	// 验证模板有名称
	for _, tmpl := range templates {
		if tmpl.Name == "" {
			t.Error("Template should have a name")
		}
	}
}

func TestTemplateService_ApplyTemplate(t *testing.T) {
	ts := NewTemplateService()

	data := map[string]string{
		"answer":       "这是一个测试答案",
		"suggestion":   "建议内容",
		"next_steps":   "下一步",
	}

	result, err := ts.ApplyTemplate("simple_response", data)
	if err != nil {
		t.Errorf("ApplyTemplate failed: %v", err)
	}

	if result == "" {
		t.Error("ApplyTemplate should return non-empty string")
	}

	// 验证占位符被替换
	if result == ts.ListTemplates()[0].Format {
		t.Error("Placeholders should be replaced")
	}
}

func TestTemplateService_ApplyTemplate_NotFound(t *testing.T) {
	ts := NewTemplateService()

	_, err := ts.ApplyTemplate("nonexistent", map[string]string{})
	if err == nil {
		t.Error("ApplyTemplate should return error for unknown template")
	}
}

func TestTemplateService_AutoFormatResponse_Code(t *testing.T) {
	ts := NewTemplateService()

	content := "func main() {\n    println(\"hello\")\n}"
	result := ts.AutoFormatResponse(content)

	if result == content {
		t.Error("AutoFormatResponse should add formatting to code")
	}

	if result != content && len(result) > len(content) {
		// 成功添加了格式
	}
}

func TestTemplateService_AutoFormatResponse_List(t *testing.T) {
	ts := NewTemplateService()

	content := "item1\nitem2\nitem3\nitem4"
	result := ts.AutoFormatResponse(content)

	if result == content {
		t.Error("AutoFormatResponse should format lists")
	}
}

func TestTemplateService_AutoFormatResponse_AlreadyFormatted(t *testing.T) {
	ts := NewTemplateService()

	content := "✅ **回答**: 测试内容"
	result := ts.AutoFormatResponse(content)

	// 已格式化的内容应该保持不变
	if result != content {
		t.Error("Already formatted content should remain unchanged")
	}
}
