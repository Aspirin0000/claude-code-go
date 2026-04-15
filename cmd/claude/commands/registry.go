package commands

import (
	"errors"
	"strings"
	"sync"
)

var ErrCommandAlreadyExists = errors.New("command already exists")

type Registry struct {
	commands   map[string]Command
	aliases    map[string]string
	categories map[CommandCategory][]Command
	mu         sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		commands:   make(map[string]Command),
		aliases:    make(map[string]string),
		categories: make(map[CommandCategory][]Command),
	}
}

func (r *Registry) Register(cmd Command) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := cmd.Name()
	if _, exists := r.commands[name]; exists {
		return ErrCommandAlreadyExists
	}

	r.commands[name] = cmd
	r.categories[cmd.Category()] = append(r.categories[cmd.Category()], cmd)

	for _, alias := range cmd.Aliases() {
		r.aliases[alias] = name
	}

	return nil
}

func (r *Registry) RegisterMultiple(cmds ...Command) {
	for _, cmd := range cmds {
		_ = r.Register(cmd)
	}
}

func (r *Registry) Get(name string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, exists := r.commands[name]
	return cmd, exists
}

func (r *Registry) GetByAlias(alias string) (Command, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	name, exists := r.aliases[alias]
	if !exists {
		return nil, false
	}

	cmd, exists := r.commands[name]
	return cmd, exists
}

func (r *Registry) List() []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Command, 0, len(r.commands))
	for _, cmd := range r.commands {
		result = append(result, cmd)
	}
	return result
}

func (r *Registry) ListNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0, len(r.commands))
	for name := range r.commands {
		result = append(result, name)
	}
	return result
}

func (r *Registry) ListByCategory(cat CommandCategory) []Command {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Command, len(r.categories[cat]))
	copy(result, r.categories[cat])
	return result
}

func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cmd, exists := r.commands[name]
	if !exists {
		return
	}

	delete(r.commands, name)

	for _, alias := range cmd.Aliases() {
		delete(r.aliases, alias)
	}

	cat := cmd.Category()
	cmds := r.categories[cat]
	for i, c := range cmds {
		if c.Name() == name {
			r.categories[cat] = append(cmds[:i], cmds[i+1:]...)
			break
		}
	}
}

func (r *Registry) Match(input string) (Command, []string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil, nil
	}

	name := parts[0]
	args := parts[1:]

	if cmd, exists := r.commands[name]; exists {
		return cmd, args
	}

	if target, exists := r.aliases[name]; exists {
		if cmd, exists := r.commands[target]; exists {
			return cmd, args
		}
	}

	return nil, args
}

var defaultRegistry = NewRegistry()

func GetRegistry() *Registry {
	return defaultRegistry
}

func Register(cmd Command) error {
	return defaultRegistry.Register(cmd)
}

func Get(name string) (Command, bool) {
	return defaultRegistry.Get(name)
}

func List() []Command {
	return defaultRegistry.List()
}
