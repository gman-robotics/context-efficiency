package child

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// FakeEngram is an in-memory fake child for unit tests.
type FakeEngram struct {
	mu             sync.Mutex
	saves          []saveRec
	nextID         int
	gets           int
	UnknownProject bool
	NetworkFS      bool
	// SearchText, when set, is the mem_search result. Empty keeps the default search.
	SearchText string
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
	return f.callNamed(name, args)
}

func (f *FakeEngram) callNamed(name string, args map[string]interface{}) (string, bool, error) {
	handler := f.toolHandler(name)
	if handler == nil {
		return "unknown tool", true, nil
	}
	return handler(args)
}

func (f *FakeEngram) toolHandler(name string) func(map[string]interface{}) (string, bool, error) {
	return map[string]func(map[string]interface{}) (string, bool, error){
		"mem_save":            f.memSave,
		"mem_search":          f.memSearch,
		"mem_get_observation": f.memGetObservation,
	}[name]
}

func (f *FakeEngram) memSave(args map[string]interface{}) (string, bool, error) {
	title := stringArg(args, "title")
	content := stringArg(args, "content")
	topic := stringArg(args, "topic_key")
	f.noteSaveProject()
	return saveEnvelope(f.pushSave(title, content, topic)), false, nil
}

func (f *FakeEngram) noteSaveProject() {
	if f.UnknownProject {
		f.UnknownProject = false
	}
}

func (f *FakeEngram) pushSave(title, content, topic string) string {
	f.nextID++
	id := fmt.Sprintf("%d", f.nextID)
	f.saves = append(f.saves, saveRec{Title: title, Content: content, TopicKey: topic, ID: id})
	return id
}

func stringArg(args map[string]interface{}, key string) string {
	v, _ := args[key].(string)
	return v
}

func saveEnvelope(id string) string {
	b, _ := json.Marshal(map[string]interface{}{"id": id})
	return string(b)
}

func (f *FakeEngram) memSearch(args map[string]interface{}) (string, bool, error) {
	if f.UnknownProject {
		return `{"error_code":"unknown_project"}`, true, nil
	}
	return f.searchKnown(args), false, nil
}

func (f *FakeEngram) searchKnown(args map[string]interface{}) string {
	if f.SearchText != "" {
		return f.SearchText
	}
	query, _ := args["query"].(string)
	return f.savedSearchJSON(query)
}

func (f *FakeEngram) savedSearchJSON(query string) string {
	var results []map[string]interface{}
	for _, s := range f.saves {
		results = appendMatch(results, s, query)
	}
	b, _ := json.Marshal(map[string]interface{}{"results": results})
	return string(b)
}

func appendMatch(results []map[string]interface{}, s saveRec, query string) []map[string]interface{} {
	if !rowVisible(s.TopicKey, query) {
		return results
	}
	return append(results, searchHit(s))
}

func rowVisible(topic, query string) bool {
	if topic == "" {
		return true
	}
	return query == topic
}

func searchHit(s saveRec) map[string]interface{} {
	return map[string]interface{}{"id": s.ID, "topic_key": s.TopicKey, "title": s.Title}
}

func (f *FakeEngram) memGetObservation(args map[string]interface{}) (string, bool, error) {
	f.gets++
	s, ok := indexSaves(f.saves)[fmt.Sprint(args["id"])]
	if !ok {
		return "not found", true, nil
	}
	return formatObservation(s), false, nil
}

func indexSaves(saves []saveRec) map[string]saveRec {
	out := map[string]saveRec{}
	for _, s := range saves {
		out[s.ID] = s
	}
	return out
}

func formatObservation(s saveRec) string {
	return fmt.Sprintf("#%s [manual] %s\n%s\n\nSession: test\nCreated: 2026-01-01T00:00:00Z\n", s.ID, s.Title, s.Content)
}

func (f *FakeEngram) LastSaveTitle() string {
	return f.lastSave().Title
}

func (f *FakeEngram) GetCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.gets
}

func (f *FakeEngram) LastTopicKey() string {
	return f.lastSave().TopicKey
}

func (f *FakeEngram) lastSave() saveRec {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.saves) == 0 {
		return saveRec{}
	}
	return f.saves[len(f.saves)-1]
}
