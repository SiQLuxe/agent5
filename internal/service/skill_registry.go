package service

import (
	"os"
	"sync"
)

type SkillType int

const (
	SkillPrompt  SkillType = iota
	SkillHandler
)

type Skill struct {
	Name        string
	Description string
	Type        SkillType
	Prompt      string
}

type SkillContext struct {
	Input     string
	SessionID string
	AI        *AIAssistant
}

type SkillRegistry struct {
	mu       sync.RWMutex
	skills   map[string]*Skill
	handlers map[string]func(ctx SkillContext) string
}

func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills:   make(map[string]*Skill),
		handlers: make(map[string]func(ctx SkillContext) string),
	}
}

func (r *SkillRegistry) Register(skill *Skill) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[skill.Name]; exists {
		return os.ErrExist
	}

	if _, handlerExists := r.handlers[skill.Name]; handlerExists {
		skill.Type = SkillHandler
	}

	r.skills[skill.Name] = skill
	return nil
}

func (r *SkillRegistry) Get(name string) (*Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skill, ok := r.skills[name]
	if !ok {
		return nil, false
	}
	return skill, true
}

func (r *SkillRegistry) List() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	skills := make([]*Skill, 0, len(r.skills))
	for _, s := range r.skills {
		skills = append(skills, s)
	}
	return skills
}

func (r *SkillRegistry) RegisterHandler(name string, fn func(ctx SkillContext) string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers[name] = fn

	if skill, exists := r.skills[name]; exists {
		skill.Type = SkillHandler
	}
}

func (r *SkillRegistry) GetHandler(name string) (func(ctx SkillContext) string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	fn, ok := r.handlers[name]
	return fn, ok
}

func (r *SkillRegistry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.skills, name)
	delete(r.handlers, name)
}

func (r *SkillRegistry) IsHandler(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.handlers[name]
	return ok
}
