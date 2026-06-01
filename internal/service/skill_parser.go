package service

import "strings"

type ParsedCommand struct {
	Name string
	Args string
}

func ParseCommand(input string) *ParsedCommand {
	input = strings.TrimSpace(input)
	if len(input) < 2 || input[0] != '/' {
		return nil
	}
	rest := input[1:]
	firstSpace := strings.Index(rest, " ")
	if firstSpace < 0 {
		return &ParsedCommand{Name: rest}
	}
	return &ParsedCommand{
		Name: rest[:firstSpace],
		Args: strings.TrimSpace(rest[firstSpace+1:]),
	}
}
