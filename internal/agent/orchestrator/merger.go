package orchestrator

import "strings"

type MergeStrategy func(results []*Task) string

type Merger struct {
	strategy MergeStrategy
}

func NewMerger() *Merger {
	return &Merger{strategy: ConcatMerge}
}

func (m *Merger) SetStrategy(s MergeStrategy) {
	m.strategy = s
}

func (m *Merger) Merge(results []*Task) string {
	return m.strategy(results)
}

func ConcatMerge(results []*Task) string {
	var b strings.Builder
	for _, r := range results {
		if r.Status == StatusCompleted {
			b.WriteString("## " + string(r.Type) + "\n")
			b.WriteString(r.Result + "\n\n")
		} else if r.Status == StatusFailed {
			b.WriteString("## " + string(r.Type) + " (FAILED)\n")
			b.WriteString(r.Error + "\n\n")
		}
	}
	return b.String()
}
