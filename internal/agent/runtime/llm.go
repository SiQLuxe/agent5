package runtime

type ToolCall struct {
	Name      string
	Arguments map[string]interface{}
}

type LLMResponse struct {
	Type     string   // "tool_call" or "final"
	Content  string
	ToolCall *ToolCall
}

type LLMClient interface {
	ChatWithTools(messages []Message, tools []map[string]interface{}, model string) (*LLMResponse, error)
}

type MockLLMClient struct {
	Responses []LLMResponse
	index     int
}

func NewMockLLMClient(responses []LLMResponse) *MockLLMClient {
	return &MockLLMClient{Responses: responses}
}

func (m *MockLLMClient) ChatWithTools(messages []Message, tools []map[string]interface{}, model string) (*LLMResponse, error) {
	if m.index >= len(m.Responses) {
		return &LLMResponse{Type: "final", Content: "done"}, nil
	}
	resp := m.Responses[m.index]
	m.index++
	return &resp, nil
}
