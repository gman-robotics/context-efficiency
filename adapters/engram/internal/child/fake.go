package child

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// FakeEngram is an in-memory fake child for unit tests.
type FakeEngram struct {
	mu       sync.Mutex
	saves    []saveRec
	nextID   int
	UnknownProject bool
	NetworkFS      bool
}

type saveRec struct {
	Title    string
	Content  string
	TopicKey string
	ID       string
}

func (f *FakeEngram) CallTool(ctx context.Context, name string, args map[string]interface{}) (string, bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.NetworkFS {
		return "NetworkFilesystemError", true, nil
	}
	switch name {
	case "mem_save":
		title, _ := args["title"].(string)
		content, _ := args["content"].(string)
		topic, _ := args["topic_key"].(string)
		if f.UnknownProject {
			f.UnknownProject = false
		}
		f.nextID++
		id := fmt.Sprintf("%d", f.nextID)
		f.saves = append(f.saves, saveRec{Title: title, Content: content, TopicKey: topic, ID: id})
		b, _ := json.Marshal(map[string]interface{}{"id": id})
		return string(b), false, nil
	case "mem_search":
		if f.UnknownProject {
			return `{"error_code":"unknown_project"}`, true, nil
		}
		query, _ := args["query"].(string)
		var results []map[string]interface{}
		for _, s := range f.saves {
			if s.TopicKey != "" && query == s.TopicKey {
				results = append(results, map[string]interface{}{"id": s.ID, "topic_key": s.TopicKey, "title": s.Title})
			} else if s.TopicKey == "" {
				results = append(results, map[string]interface{}{"id": s.ID, "topic_key": "", "title": s.Title})
			}
		}
		b, _ := json.Marshal(map[string]interface{}{"results": results})
		return string(b), false, nil
	case "mem_get_observation":
		id := fmt.Sprint(args["id"])
		for _, s := range f.saves {
			if s.ID == id {
				text := fmt.Sprintf("#%s [manual] %s\n%s\n\nSession: test\nCreated: 2026-01-01T00:00:00Z\n", s.ID, s.Title, s.Content)
				return text, false, nil
			}
		}
		return "not found", true, nil
	default:
		return "unknown tool", true, nil
	}
}

func (f *FakeEngram) LastSaveTitle() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.saves) == 0 {
		return ""
	}
	return f.saves[len(f.saves)-1].Title
}
