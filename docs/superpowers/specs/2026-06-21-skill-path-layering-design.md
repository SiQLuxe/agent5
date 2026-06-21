# Skill Path Layering — Design

Date: 2026-06-21
Status: Approved

## 目标

把 skill 加载路径从硬编码 `"skills"` 改为**双层分层**（项目级 + 用户级），对齐 Claude Code 的 skill 路径约定，并支持配置覆盖。

## 背景

当前 `cmd/agent/main.go:201` 硬编码 `LoadSkillsDir(skillRegistry, "skills")`：
- 必须从项目根目录启动，否则找不到 skills。
- 无用户级路径，个人 skill 无法跨项目复用。
- 无配置项，无法自定义或关闭某层。

Claude Code 约定：项目级 `.claude/skills/` + 用户级 `~/.claude/skills/`。

## 路径分层（默认）

| 层 | 默认路径 | 用途 |
|---|---|---|
| 项目级 | `./skills/` | 随仓库分发，团队共享 |
| 用户级 | `$XDG_CONFIG_HOME/agent-tui/skills/`（无 XDG 时 `~/.config/agent-tui/skills/`） | 个人全局，跨项目 |

## 优先级

**项目级优先**：同名 skill 时项目级胜出，用户级被跳过。

实现方式：**先加载项目级，再加载用户级**。`SkillRegistry.Register` 对已存在的 name 返回 `ErrSkillAlreadyExists`，`LoadSkillsDir` 内部对该错误 `return nil`（忽略），用户级同名 skill 自然被跳过。零 API 变更。

## 配置项

`internal/data/config/config.go` 新增：
```go
type SkillsConfig struct {
    Dirs []string `toml:"dirs"`
}

type Config struct {
    // ...existing code...
    Skills SkillsConfig `toml:"skills"`
}
```

`GetDefaultConfig()` 默认：
```go
Skills: SkillsConfig{
    Dirs: []string{"skills", "~/.config/agent-tui/skills"},
},
```

### 覆盖语义

- 用户在 `config.toml` 写 `[skills] dirs = [...]` → **完全替换**默认（TOML 数组语义，非合并）。
- 不写 → 用默认两层。
- 写 `dirs = ["skills"]` → 仅项目级，关闭用户级。
- 写 `dirs = ["my-skills", "skills", "~/.config/agent-tui/skills"]` → 三层，前者优先。

顺序即优先级：数组前面的先加载，后面的同名被跳过。

## 路径展开

新增 `internal/data/config/paths.go`：
```go
func ExpandSkillPath(p string) string {
    // 1. $XDG_CONFIG_HOME 优先（仅对 ~/.config 前缀）
    if strings.HasPrefix(p, "~/.config") {
        if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
            return strings.Replace(p, "~/.config", xdg, 1)
        }
    }
    // 2. ~ → $HOME
    if strings.HasPrefix(p, "~/") {
        home, _ := os.UserHomeDir()
        return filepath.Join(home, p[2:])
    }
    return p
}
```

## main.go 改动

```go
dirs := cfg.Skills.Dirs
if len(dirs) == 0 {
    dirs = []string{"skills", "~/.config/agent-tui/skills"}
}
for _, d := range dirs {
    expanded := config.ExpandSkillPath(d)
    if err := service.LoadSkillsDir(skillRegistry, expanded); err != nil {
        log.Printf("warning: loading skills from %s: %v", d, err)
    }
}
app.SetSkillsDir(config.ExpandSkillPath(dirs[0]))  // UI watcher 监听项目级
```

## UI 热重载

- 保留 `SetSkillsDir(dirs[0])`（项目级）作为主监听目录。
- 用户级目录暂不监听（YAGNI，用户级改动频率低）。
- 后续如需可扩展 watcher 支持多目录。

## config.example.toml 更新

加注释段：
```toml
# Skill directories. Order = priority (earlier wins on name conflict).
# Defaults to ["skills", "~/.config/agent-tui/skills"] if omitted.
# ~ expands to $HOME; ~/.config honors $XDG_CONFIG_HOME.
[skills]
# dirs = ["skills", "~/.config/agent-tui/skills"]
```

## 测试

### `internal/data/config/config_test.go`
- 默认 `Skills.Dirs` == `["skills", "~/.config/agent-tui/skills"]`
- TOML 解析 `[skills] dirs = ["a","b"]` 正确覆盖默认
- `ExpandSkillPath`：
  - `~/x` → `$HOME/x`
  - `~/.config/x` + 设 `$XDG_CONFIG_HOME` → 替换为 XDG 路径
  - `~/.config/x` 无 XDG → 保持 `~/.config/x` 展开 `$HOME`
  - 相对路径 `skills` 原样返回

### `internal/service/skill_pipeline_test.go`
- 双层加载：tempdir 模拟项目级 + 用户级，同名时项目级胜出
- 仅用户级存在时正常加载
- 空配置 dirs 回退默认两层

## 不变

- `LoadSkillsDir` / `ReloadSkillsDir` / `SkillRegistry` / `SkillExecutor` API 不变
- UI skill overlay 不变
- 现有 `skills/code-review`、`skills/echo-test` 位置不变

## 风险

- **TOML 数组覆盖语义**：用户配 `dirs` 完全替换默认。config.example.toml 注释说明。可接受。
- **XDG 跨平台**：Windows 无 `$XDG_CONFIG_HOME`，回退 `~/.config`。本项目目前 macOS/Linux 为主，YAGNI。
- **热重载仅项目级**：用户级改动需重启。可接受。

## 不在范围

- watcher 支持多目录（YAGNI）
- Windows `%APPDATA%` 支持（YAGNI）
- skill 冲突时的 UI 提示（YAGNI）
