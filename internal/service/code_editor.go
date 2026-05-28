package service

import (
	"os"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type CodeEditorService struct {
	openFiles map[string]string
}

func NewCodeEditorService() *CodeEditorService {
	return &CodeEditorService{
		openFiles: make(map[string]string),
	}
}

func (ces *CodeEditorService) OpenFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	content := string(data)
	ces.openFiles[path] = content
	return content, nil
}

func (ces *CodeEditorService) SaveFile(path, content string) error {
	err := os.WriteFile(path, []byte(content), 0644)
	if err == nil {
		ces.openFiles[path] = content
	}
	return err
}

func (ces *CodeEditorService) GetFileContent(path string) (string, bool) {
	content, exists := ces.openFiles[path]
	return content, exists
}

func (ces *CodeEditorService) CloseFile(path string) {
	delete(ces.openFiles, path)
}

func (ces *CodeEditorService) ListOpenFiles() []string {
	files := make([]string, 0, len(ces.openFiles))
	for f := range ces.openFiles {
		files = append(files, f)
	}
	return files
}

func (ces *CodeEditorService) HighlightSyntax(content, language string) string {
	lexer := lexers.Get(language)
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

	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return content
	}

	var builder strings.Builder
	err = formatter.Format(&builder, style, iterator)
	if err != nil {
		return content
	}

	return builder.String()
}

func (ces *CodeEditorService) DetectLanguage(path string) string {
	lexer := lexers.Match(path)
	if lexer != nil {
		return lexer.Config().Name
	}
	return "text"
}