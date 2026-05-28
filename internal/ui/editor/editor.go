package editor

import (
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type CodeEditor struct {
	FilePath   string
	Content    string
	Language   string
	CursorLine int
	CursorCol  int
}

func NewCodeEditor() *CodeEditor {
	return &CodeEditor{
		Content:    "",
		Language:   "go",
		CursorLine: 0,
		CursorCol:  0,
	}
}

func (ce *CodeEditor) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	ce.FilePath = path
	ce.Content = string(data)
	ce.detectLanguage()
	return nil
}

func (ce *CodeEditor) SaveFile(path string) error {
	return os.WriteFile(path, []byte(ce.Content), 0644)
}

func (ce *CodeEditor) detectLanguage() {
	lexer := lexers.Match(ce.FilePath)
	if lexer != nil {
		ce.Language = lexer.Config().Name
	}
}

func (ce *CodeEditor) SetContent(content string) {
	ce.Content = content
}

func (ce *CodeEditor) GetContent() string {
	return ce.Content
}

func (ce *CodeEditor) SetLanguage(lang string) {
	ce.Language = lang
}

func (ce *CodeEditor) GetLanguage() string {
	return ce.Language
}

func (ce *CodeEditor) SetCursorPosition(line, col int) {
	ce.CursorLine = line
	ce.CursorCol = col
}

func (ce *CodeEditor) GetCursorPosition() (int, int) {
	return ce.CursorLine, ce.CursorCol
}

func (ce *CodeEditor) Highlight() string {
	lexer := lexers.Get(ce.Language)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	formatter := formatters.Get("terminal")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, ce.Content)
	if err != nil {
		return ce.Content
	}

	var builder strings.Builder
	err = formatter.Format(&builder, style, iterator)
	if err != nil {
		return ce.Content
	}

	return builder.String()
}

func (ce *CodeEditor) GetLines() []string {
	return strings.Split(ce.Content, "\n")
}

func (ce *CodeEditor) InsertLine(lineNum int, content string) {
	lines := ce.GetLines()
	if lineNum < 0 {
		lineNum = 0
	}
	if lineNum > len(lines) {
		lineNum = len(lines)
	}
	lines = append(lines[:lineNum], append([]string{content}, lines[lineNum:]...)...)
	ce.Content = strings.Join(lines, "\n")
}

func (ce *CodeEditor) DeleteLine(lineNum int) {
	lines := ce.GetLines()
	if lineNum >= 0 && lineNum < len(lines) {
		lines = append(lines[:lineNum], lines[lineNum+1:]...)
		ce.Content = strings.Join(lines, "\n")
	}
}