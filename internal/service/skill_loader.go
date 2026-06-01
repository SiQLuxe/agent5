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
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() {
		return nil
	}

	return filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fi.IsDir() || fi.Name() != "SKILL.md" {
			return nil
		}
		parentDir := filepath.Base(filepath.Dir(path))
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		name, desc, body, err := parseFrontmatter(string(data))
		if err != nil {
			return nil
		}
		if name != parentDir {
			return nil
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
		return nil
	})
}

func ReloadSkillsDir(registry *SkillRegistry, dir string) error {
	registry.ClearPrompts()
	return LoadSkillsDir(registry, dir)
}
