package config

import (
	"os"
	"path/filepath"
	"strings"
)

// ExpandSkillPath expands a skill directory path:
//   - "~/.config/..." honors $XDG_CONFIG_HOME when set (replaces the
//     "~/.config" prefix with $XDG_CONFIG_HOME).
//   - "~/..." expands to "$HOME/...".
//   - Relative and absolute paths are returned unchanged.
func ExpandSkillPath(p string) string {
	if strings.HasPrefix(p, "~/.config") {
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			return strings.Replace(p, "~/.config", xdg, 1)
		}
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return p
		}
		return filepath.Join(home, p[2:])
	}
	return p
}
