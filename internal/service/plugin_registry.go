package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Plugin struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Path        string `json:"path"`
	Enabled     bool   `json:"enabled"`
}

type PluginRegistry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

func NewPluginRegistry() *PluginRegistry {
	return &PluginRegistry{
		plugins: make(map[string]Plugin),
	}
}

func (pr *PluginRegistry) LoadPlugins(pluginDir string) error {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(pluginDir, 0755)
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			pluginPath := filepath.Join(pluginDir, entry.Name())
			manifestPath := filepath.Join(pluginPath, "plugin.json")

			if _, err := os.Stat(manifestPath); err == nil {
				data, err := os.ReadFile(manifestPath)
				if err != nil {
					continue
				}

				var plugin Plugin
				if err := json.Unmarshal(data, &plugin); err != nil {
					continue
				}

				plugin.Path = pluginPath
				pr.mu.Lock()
				pr.plugins[plugin.ID] = plugin
				pr.mu.Unlock()
			}
		}
	}

	return nil
}

func (pr *PluginRegistry) RegisterPlugin(plugin Plugin) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if _, exists := pr.plugins[plugin.ID]; exists {
		return os.ErrExist
	}

	pr.plugins[plugin.ID] = plugin
	return nil
}

func (pr *PluginRegistry) GetPlugin(id string) (*Plugin, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	plugin, exists := pr.plugins[id]
	if !exists {
		return nil, os.ErrNotExist
	}
	return &plugin, nil
}

func (pr *PluginRegistry) ListPlugins() []Plugin {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	plugins := make([]Plugin, 0, len(pr.plugins))
	for _, p := range pr.plugins {
		plugins = append(plugins, p)
	}
	return plugins
}

func (pr *PluginRegistry) EnablePlugin(id string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	plugin, exists := pr.plugins[id]
	if !exists {
		return os.ErrNotExist
	}

	plugin.Enabled = true
	pr.plugins[id] = plugin
	return nil
}

func (pr *PluginRegistry) DisablePlugin(id string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	plugin, exists := pr.plugins[id]
	if !exists {
		return os.ErrNotExist
	}

	plugin.Enabled = false
	pr.plugins[id] = plugin
	return nil
}

func (pr *PluginRegistry) UnregisterPlugin(id string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if _, exists := pr.plugins[id]; !exists {
		return os.ErrNotExist
	}

	delete(pr.plugins, id)
	return nil
}

func (pr *PluginRegistry) ListEnabledPlugins() []Plugin {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	var enabled []Plugin
	for _, p := range pr.plugins {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}
	return enabled
}