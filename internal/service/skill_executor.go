package service

import (
	"fmt"
	"strings"
)

// LLMChatter provides a simple chat completion (no tools needed for skills)
type LLMChatter interface {
	Chat(sessionID, message string) (string, error)
}

type SkillExecutor struct {
	registry *SkillRegistry
	chater   LLMChatter
}

func NewSkillExecutor(registry *SkillRegistry, chater LLMChatter) *SkillExecutor {
	return &SkillExecutor{registry: registry, chater: chater}
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
		if e.chater != nil {
			return e.chater.Chat("skill-"+skill.Name, prompt)
		}
		return prompt, nil
	}

	return "", fmt.Errorf("unknown skill type")
}
