package tool

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

func (r *Registry) Register(t Tool) {
	r.tools[t.Name()] = t
}

func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) List() []Tool {
	list := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

func (r *Registry) AsToolDefinitions() []map[string]interface{} {
	defs := make([]map[string]interface{}, 0, len(r.tools))
	for _, t := range r.tools {
		schema := t.Schema()
		properties := make(map[string]interface{})
		for name, ps := range schema.Parameters {
			prop := map[string]interface{}{
				"type":        ps.Type,
				"description": ps.Description,
			}
			if len(ps.Enum) > 0 {
				prop["enum"] = ps.Enum
			}
			properties[name] = prop
		}
		defs = append(defs, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        t.Name(),
				"description": t.Description(),
				"parameters": map[string]interface{}{
					"type":       "object",
					"properties": properties,
					"required":   schema.Required,
				},
			},
		})
	}
	return defs
}
