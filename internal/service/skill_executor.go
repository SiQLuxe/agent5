package service

import (
	"fmt"
	"strings"
)

type SkillExecutor struct {
	registry    *SkillRegistry
	aiAssistant *AIAssistant
}

func NewSkillExecutor(registry *SkillRegistry, ai *AIAssistant) *SkillExecutor {
	return &SkillExecutor{registry: registry, aiAssistant: ai}
}

func (e *SkillExecutor) Execute(cmd *ParsedCommand) (string, error) {
	skill, ok := e.registry.Get(cmd.Name)
	if !ok {
		return "", fmt.Errorf("skill not found: %s", cmd.Name)
	}

	switch skill.Type {
	case SkillHandler:
		fn, ok := e.registry.GetHandler(skill.Name)
		if !ok {
			return "", fmt.Errorf("no handler registered for: %s", skill.Name)
		}
		return fn(SkillContext{Input: cmd.Args}), nil

	case SkillPrompt:
		prompt := skill.Prompt
		if cmd.Args != "" && strings.Contains(prompt, "%s") {
			prompt = strings.ReplaceAll(prompt, "%s", cmd.Args)
		}
		if e.aiAssistant != nil {
			return e.aiAssistant.Chat("skill-"+skill.Name, prompt)
		}
		return prompt, nil
	}

	return "", fmt.Errorf("unknown skill type")
}
