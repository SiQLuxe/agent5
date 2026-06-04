package runtime

type Message struct {
	Role    string
	Content string
}

type Memory struct {
	ShortTerm  []Message
	WorkingSet map[string]string
}

func NewMemory() *Memory {
	return &Memory{
		ShortTerm:  make([]Message, 0),
		WorkingSet: make(map[string]string),
	}
}

func (m *Memory) Append(role, content string) {
	m.ShortTerm = append(m.ShortTerm, Message{Role: role, Content: content})
}

func (m *Memory) Set(key, value string) {
	m.WorkingSet[key] = value
}

func (m *Memory) Get(key string) string {
	return m.WorkingSet[key]
}

func (m *Memory) Clear() {
	m.ShortTerm = make([]Message, 0)
	m.WorkingSet = make(map[string]string)
}
