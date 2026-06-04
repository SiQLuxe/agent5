package tool

type LLMProvider interface {
	Chat(model, systemPrompt, userPrompt string) (string, error)
}

type ChatLLMTool struct {
	Provider LLMProvider
}

func (t *ChatLLMTool) Name() string { return "chat_llm" }

func (t *ChatLLMTool) Description() string { return "Ask an LLM to respond to a prompt" }

func (t *ChatLLMTool) Schema() ToolSchema {
	return ToolSchema{
		Parameters: map[string]ParamSchema{
			"prompt": {Type: "string", Description: "The prompt to send to the LLM"},
			"model":  {Type: "string", Description: "Model name (optional, uses default if empty)"},
		},
		Required: []string{"prompt"},
	}
}

func (t *ChatLLMTool) Execute(ctx ToolContext, params map[string]interface{}) ToolResult {
	prompt, _ := params["prompt"].(string)
	if prompt == "" {
		return ToolResult{Error: "prompt parameter is required"}
	}
	model, _ := params["model"].(string)

	resp, err := t.Provider.Chat(model, "", prompt)
	if err != nil {
		return ToolResult{Error: err.Error()}
	}
	return ToolResult{Success: true, Data: resp}
}
