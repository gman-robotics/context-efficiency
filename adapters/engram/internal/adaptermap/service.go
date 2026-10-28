package adaptermap

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/child"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/content"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/domain"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/engram"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/scope"
)

// Service maps domain tools to engram child calls.
type Service struct {
	Client child.ToolCaller
}

func (s *Service) Retain(ctx context.Context, text string, tags []string, importance int, scopeStr string) (string, error) {
	scopeStr = defaultScope(scopeStr)
	if err := content.RejectPrivateInString(text); err != nil {
		return "", err
	}
	token, err := scope.Token(scopeStr)
	if err != nil {
		return "", err
	}
	body := domain.RetainBody{Text: text, Tags: tags, Importance: importance, Scope: scopeStr}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	if len(raw) > 40000 {
		return "", fmt.Errorf("retain body too large")
	}
	if len(text) > 8000 {
		return "", fmt.Errorf("text too large")
	}
	prepared, err := content.PrepareForStore(string(raw))
	if err != nil {
		return "", err
	}
	title, err := content.PrepareTitle(content.RetainTitle(token, text))
	if err != nil {
		return "", err
	}
	return s.save(ctx, title, prepared, "")
}

func (s *Service) Recall(ctx context.Context, query string, scopeStr string, limit int) ([]domain.MemoryItem, error) {
	scopeStr = defaultScope(scopeStr)
	token, err := scope.Token(scopeStr)
	if err != nil {
		return nil, err
	}
	if limit < 1 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}
	q := token + " " + strings.TrimSpace(query)
	resText, isErr, err := s.Client.CallTool(ctx, "mem_search", map[string]interface{}{
		"query": q,
		"limit": limit,
	})
	if err != nil {
		return nil, err
	}
	if isErr {
		if engram.IsUnknownProject(resText) {
			return []domain.MemoryItem{}, nil
		}
		return nil, fmt.Errorf("mem_search: %s", resText)
	}
	results, err := engram.ParseSearchResults(resText)
	if err != nil {
		return []domain.MemoryItem{}, nil
	}
	if results == nil {
		return []domain.MemoryItem{}, nil
	}
	var out []domain.MemoryItem
	for _, r := range results {
		if scope.TitleFirstField(r.Title) != token {
			continue
		}
		if strings.HasPrefix(strings.ToLower(r.Title), content.HandoffTitle) {
			continue
		}
		id := r.ID.String()
		obs, isErr, err := s.Client.CallTool(ctx, "mem_get_observation", map[string]interface{}{"id": observationIDArg(id)})
		if err != nil || isErr {
			continue
		}
		body, createdAt, err := engram.ParseObservationBody(engram.UnwrapGetObservationText(obs))
		if err != nil {
			continue
		}
		var rb domain.RetainBody
		if err := json.Unmarshal([]byte(body), &rb); err != nil {
			continue
		}
		out = append(out, domain.MemoryItem{
			ID:         id,
			Text:       rb.Text,
			Tags:       rb.Tags,
			Importance: rb.Importance,
			Scope:      rb.Scope,
			CreatedAt:  createdAt,
		})
	}
	return out, nil
}

func (s *Service) HandoffWrite(ctx context.Context, h domain.Handoff, scopeStr string) (string, error) {
	scopeStr = defaultScope(scopeStr)
	if err := content.RejectPrivateInHandoff(h); err != nil {
		return "", err
	}
	token, err := scope.Token(scopeStr)
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	if len(raw) > 40000 {
		return "", fmt.Errorf("handoff body too large")
	}
	prepared, err := content.PrepareForStore(string(raw))
	if err != nil {
		return "", err
	}
	title := content.HandoffTitle
	topic := scope.HandoffTopicKey(token)
	id, err := s.save(ctx, title, prepared, topic)
	if err != nil {
		return "", err
	}
	if err := s.verifyStoredBody(ctx, id, prepared); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) HandoffRead(ctx context.Context, scopeStr string) (*domain.Handoff, error) {
	scopeStr = defaultScope(scopeStr)
	token, err := scope.Token(scopeStr)
	if err != nil {
		return nil, err
	}
	topic := scope.HandoffTopicKey(token)
	resText, isErr, err := s.Client.CallTool(ctx, "mem_search", map[string]interface{}{
		"query": topic,
		"limit": 20,
	})
	if err != nil {
		return nil, err
	}
	if isErr {
		if engram.IsUnknownProject(resText) {
			return nil, nil
		}
		return nil, fmt.Errorf("mem_search: %s", resText)
	}
	results, err := engram.ParseSearchResults(resText)
	if err != nil || results == nil {
		return nil, nil
	}
	var id string
	for _, r := range results {
		if r.TopicKey == topic {
			id = r.ID.String()
			break
		}
	}
	if id == "" {
		return nil, nil
	}
	obs, isErr, err := s.Client.CallTool(ctx, "mem_get_observation", map[string]interface{}{"id": observationIDArg(id)})
	if err != nil || isErr {
		return nil, fmt.Errorf("mem_get_observation failed")
	}
	body, _, err := engram.ParseObservationBody(engram.UnwrapGetObservationText(obs))
	if err != nil {
		return nil, err
	}
	var h domain.Handoff
	if err := json.Unmarshal([]byte(body), &h); err != nil {
		return nil, err
	}
	return &h, nil
}

func (s *Service) save(ctx context.Context, title, contentBody, topicKey string) (string, error) {
	args := map[string]interface{}{
		"title":          title,
		"content":        contentBody,
		"type":           "manual",
		"scope":          "project",
		"capture_prompt": false,
	}
	if topicKey != "" {
		args["topic_key"] = topicKey
	}
	resText, isErr, err := s.Client.CallTool(ctx, "mem_save", args)
	if err != nil {
		return "", err
	}
	if isErr {
		return "", fmt.Errorf("mem_save: %s", resText)
	}
	id, err := engram.ParseSaveID(resText)
	if err != nil {
		return "", err
	}
	if err := s.verifyStoredBody(ctx, id, contentBody); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) verifyStoredBody(ctx context.Context, id, want string) error {
	obs, isErr, err := s.Client.CallTool(ctx, "mem_get_observation", map[string]interface{}{"id": observationIDArg(id)})
	if err != nil || isErr {
		return fmt.Errorf("stored body verification failed: %v isErr=%v", err, isErr)
	}
	got, _, err := engram.ParseObservationBody(engram.UnwrapGetObservationText(obs))
	if err != nil {
		return err
	}
	if !jsonEqual(got, want) {
		return fmt.Errorf("stored body differs from sent body")
	}
	return nil
}

func observationIDArg(id string) interface{} {
	if n, err := json.Number(id).Int64(); err == nil {
		return n
	}
	return id
}

func jsonEqual(a, b string) bool {
	var va, vb interface{}
	if err := json.Unmarshal([]byte(a), &va); err != nil {
		return strings.TrimSpace(a) == strings.TrimSpace(b)
	}
	if err := json.Unmarshal([]byte(b), &vb); err != nil {
		return false
	}
	ba, _ := json.Marshal(va)
	bb, _ := json.Marshal(vb)
	return string(ba) == string(bb)
}

func defaultScope(s string) string {
	if strings.TrimSpace(s) == "" {
		return scope.DefaultScope
	}
	return strings.TrimSpace(s)
}
