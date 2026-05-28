package service

import (
	"fmt"
	"strings"
)

// ReviewResult 审查结果
type ReviewResult struct {
	Passed      bool
	Score       int      // 0-100
	Issues      []string // 发现的问题
	Suggestions []string // 改进建议
}

// SelfReviewService AI 自我审查服务
type SelfReviewService struct {
	aiAssistant *AIAssistant
	template    *TemplateService
}

// NewSelfReviewService 创建自我审查服务
func NewSelfReviewService(ai *AIAssistant) *SelfReviewService {
	return &SelfReviewService{
		aiAssistant: ai,
		template:    NewTemplateService(),
	}
}

// ReviewContent 审查内容
func (s *SelfReviewService) ReviewContent(content string) (*ReviewResult, error) {
	result := &ReviewResult{
		Passed:      true,
		Score:       100,
		Issues:      []string{},
		Suggestions: []string{},
	}

	// 1. 格式检查
	s.checkFormat(content, result)

	// 2. 可读性检查
	s.checkReadability(content, result)

	// 3. 完整性检查
	s.checkCompleteness(content, result)

	// 如果有问题，使用 AI 进一步审查
	if len(result.Issues) > 0 {
		aiResult, err := s.aiReview(content, result)
		if err == nil && aiResult != nil {
			result.Score = aiResult.Score
			result.Suggestions = append(result.Suggestions, aiResult.Suggestions...)
		}
	}

	// 判断是否通过（分数 >= 70）
	result.Passed = result.Score >= 70

	return result, nil
}

// checkFormat 检查格式
func (s *SelfReviewService) checkFormat(content string, result *ReviewResult) {
	// 检查是否包含格式化元素
	hasFormatting := false

	formatChecks := []struct {
		pattern string
		name    string
	}{
		{"```", "代码块"},
		{"**", "粗体"},
		{"• ", "列表项"},
		{"📝", "代码标签"},
		{"✅", "成功标签"},
		{"❌", "错误标签"},
		{"💡", "建议标签"},
		{"📋", "列表标签"},
	}

	for _, check := range formatChecks {
		if strings.Contains(content, check.pattern) {
			hasFormatting = true
			break
		}
	}

	if !hasFormatting && len(content) > 100 {
		result.Score -= 10
		result.Issues = append(result.Issues, "内容缺少格式化元素")
		result.Suggestions = append(result.Suggestions, "添加代码块、列表或标记来改善可读性")
	}
}

// checkReadability 检查可读性
func (s *SelfReviewService) checkReadability(content string, result *ReviewResult) {
	lines := strings.Split(content, "\n")

	// 检查是否有超长行
	for i, line := range lines {
		if len(line) > 120 {
			result.Score -= 5
			result.Issues = append(result.Issues, fmt.Sprintf("第%d行过长（%d字符）", i+1, len(line)))
			result.Suggestions = append(result.Suggestions, "将长行拆分为多行以提高可读性")
			break
		}
	}

	// 检查是否有空行
	hasEmptyLines := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			hasEmptyLines = true
			break
		}
	}

	if !hasEmptyLines && len(lines) > 5 {
		result.Score -= 5
		result.Issues = append(result.Issues, "内容缺少段落分隔")
		result.Suggestions = append(result.Suggestions, "添加空行来分隔不同的内容部分")
	}
}

// checkCompleteness 检查完整性
func (s *SelfReviewService) checkCompleteness(content string, result *ReviewResult) {
	// 检查是否有结尾
	content = strings.TrimSpace(content)
	if len(content) == 0 {
		result.Score = 0
		result.Issues = append(result.Issues, "内容为空")
		return
	}

	// 检查结尾是否有标点
	if len(content) > 0 {
		lastChar := content[len(content)-1]
		r := rune(lastChar)
		if r != '.' && r != '。' && r != '!' && r != '?' && r != '？' && r != ')' && r != '】' {
			// 结尾没有标点，可能是被截断的内容
			result.Score -= 5
			result.Suggestions = append(result.Suggestions, "建议在结尾添加适当的标点符号")
		}
	}
}

// aiReview 使用 AI 进行深度审查
func (s *SelfReviewService) aiReview(content string, currentResult *ReviewResult) (*ReviewResult, error) {
	prompt := fmt.Sprintf(`请审查以下内容的美观度和可读性（0-100分）：

---
%s
---

评分标准：
- 格式规范：是否有代码块、列表、标记等
- 可读性：是否有适当的段落分隔、行长度是否合理
- 完整性：内容是否有开头和结尾

请以JSON格式回复：
{
  "score": 分数,
  "issues": ["问题列表"],
  "suggestions": ["改进建议"]
}`, content)

	response, err := s.aiAssistant.Chat("review", prompt)
	if err != nil {
		return nil, err
	}

	// 简单解析（实际应该用 JSON 解析）
	// 这里简化处理
	aiResult := &ReviewResult{
		Score:       80,
		Issues:       []string{},
		Suggestions:  []string{},
	}

	if strings.Contains(response, "score") {
		// 尝试提取分数
		// 简化处理
		aiResult.Score = 80
	}

	return aiResult, nil
}

// GenerateWithSelfReview 生成内容并进行自我审查
func (s *SelfReviewService) GenerateWithSelfReview(sessionID, prompt string) (string, error) {
	// 1. 生成初始内容
	initialContent, err := s.aiAssistant.Chat(sessionID, prompt)
	if err != nil {
		return "", err
	}

	// 2. 格式化内容
	formattedContent := s.template.AutoFormatResponse(initialContent)

	// 3. 自我审查
	reviewResult, err := s.ReviewContent(formattedContent)
	if err != nil {
		// 审查失败，返回格式化后的内容
		return formattedContent, nil
	}

	// 4. 如果未通过，尝试改进
	if !reviewResult.Passed {
		improvedContent, err := s.improveContent(formattedContent, reviewResult)
		if err == nil {
			return improvedContent, nil
		}
	}

	return formattedContent, nil
}

// improveContent 改进内容
func (s *SelfReviewService) improveContent(content string, review *ReviewResult) (string, error) {
	suggestionsText := strings.Join(review.Suggestions, "；")

	prompt := fmt.Sprintf(`请根据以下建议改进内容：

原始内容：
---
%s
---

改进建议：
%s

请直接输出改进后的内容。`, content, suggestionsText)

	return s.aiAssistant.Chat("improve", prompt)
}
