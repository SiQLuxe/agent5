package runtime

import (
	"time"

	"github.com/example/agent-tui/internal/agent/tool"
)

type LogEntry struct {
	Timestamp  time.Time
	Phase      string
	Content    string
	ToolName   string
	ToolParams map[string]interface{}
	ToolResult *tool.ToolResult
	Duration   time.Duration
}

type Logger struct {
	entries    []LogEntry
	maxEntries int
}

func NewLogger(max int) *Logger {
	return &Logger{
		entries:    make([]LogEntry, 0, max),
		maxEntries: max,
	}
}

func (l *Logger) Log(phase, content, toolName string, params map[string]interface{}, result *tool.ToolResult, duration time.Duration) {
	entry := LogEntry{
		Timestamp:  time.Now(),
		Phase:      phase,
		Content:    content,
		ToolName:   toolName,
		ToolParams: params,
		ToolResult: result,
		Duration:   duration,
	}
	if len(l.entries) >= l.maxEntries {
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, entry)
}

func (l *Logger) Entries() []LogEntry {
	return l.entries
}

func (l *Logger) Clear() {
	l.entries = make([]LogEntry, 0, l.maxEntries)
}
