package domain

// Scope values from docs/plan.md.
type Scope string

const (
	ScopeBot    Scope = "bot"
	ScopeShared Scope = "shared"
)

// MemoryItem is returned by recall.
type MemoryItem struct {
	ID         string   `json:"id"`
	Text       string   `json:"text"`
	Tags       []string `json:"tags"`
	Importance int      `json:"importance"`
	Scope      string   `json:"scope"`
	CreatedAt  string   `json:"createdAt"`
}

// Handoff shape from docs/plan.md.
type Handoff struct {
	Goal       string   `json:"goal"`
	OpenItems  []string `json:"openItems"`
	Decisions  []string `json:"decisions"`
	FilePaths  []string `json:"filePaths"`
	NextSteps  []string `json:"nextSteps"`
}

// RetainBody is stored in mem_save content for retain.
type RetainBody struct {
	Text       string   `json:"text"`
	Tags       []string `json:"tags"`
	Importance int      `json:"importance"`
	Scope      string   `json:"scope"`
}
