package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func parseFrontmatter(input string) (name, description, body string, err error) {
	if !strings.HasPrefix(input, "---\n") {
		return "", "", "", fmt.Errorf("missing opening ---")
	}

	rest := input[4:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return "", "", "", fmt.Errorf("missing closing ---")
	}

	yamlPart := rest[:idx]
	body = strings.TrimLeft(rest[idx+4:], "\n")

	for _, line := range strings.Split(yamlPart, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:colonIdx])
		val := strings.TrimSpace(line[colonIdx+1:])
		switch key {
		case "name":
			name = val
		case "description":
			description = val
		}
	}

	if name == "" || description == "" {
		return "", "", "", fmt.Errorf("name and description are required")
	}
	return name, description, body, nil
}

func LoadSkillsDir(registry *SkillRegistry, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillPath := filepath.Join(dir, entry.Name(), "SKILL.md")
		data, err := os.ReadFile(skillPath)
		if err != nil {
			continue
		}
		name, desc, body, err := parseFrontmatter(string(data))
		if err != nil {
			continue
		}
		if name != entry.Name() {
			continue
		}
		skillType := SkillPrompt
		if registry.IsHandler(name) {
			skillType = SkillHandler
		}
		registry.Register(&Skill{
			Name:        name,
			Description: desc,
			Type:        skillType,
			Prompt:      body,
		})
	}
	return nil
}
