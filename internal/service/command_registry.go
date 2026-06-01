package service

import (
	"sync"
)

type CommandCategory string

const (
	CmdBuiltin CommandCategory = "builtin"
	CmdSkill   CommandCategory = "skill"
)

type Command struct {
	Name        string
	Description string
	Category    CommandCategory
	Action      func(args string)
}

type CommandRegistry struct {
	mu       sync.RWMutex
	commands map[string]*Command
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]*Command),
	}
}

func (r *CommandRegistry) Register(cmd *Command) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.commands[cmd.Name] = cmd
}

func (r *CommandRegistry) Get(name string) *Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.commands[name]
}

func (r *CommandRegistry) List() []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*Command, 0, len(r.commands))
	for _, c := range r.commands {
		out = append(out, c)
	}
	return out
}

func (r *CommandRegistry) ListByCategory(cat CommandCategory) []*Command {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*Command
	for _, c := range r.commands {
		if c.Category == cat {
			out = append(out, c)
		}
	}
	return out
}

func (r *CommandRegistry) ClearCategory(cat CommandCategory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for name, cmd := range r.commands {
		if cmd.Category == cat {
			delete(r.commands, name)
		}
	}
}

func (r *CommandRegistry) SyncSkills(registry *SkillRegistry) {
	skills := registry.List()
	for _, s := range skills {
		name := s.Name
		desc := s.Description
		r.Register(&Command{
			Name:        name,
			Description: desc,
			Category:    CmdSkill,
			Action: func(args string) {},
		})
	}
}
