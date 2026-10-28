package engram

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var createdLine = regexp.MustCompile(`(?m)^Created:\s*(.+)$`)

// UnwrapGetObservationText pulls the observation text from mem_get_observation MCP result.
func UnwrapGetObservationText(resultText string) string {
	resultText = strings.TrimSpace(resultText)
	var wrap struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal([]byte(resultText), &wrap); err == nil && wrap.Result != "" {
		return wrap.Result
	}
	return resultText
}

// ParseObservationBody extracts stored JSON body from mem_get_observation text.
// Body is after the `#id [type] title` line until a `Session:` line.
func ParseObservationBody(observationText string) (body string, createdAt string, err error) {
	observationText = strings.TrimSpace(observationText)
	if m := createdLine.FindStringSubmatch(observationText); len(m) == 2 {
		createdAt = strings.TrimSpace(m[1])
	}
	if strings.HasPrefix(observationText, "{") {
		return observationText, createdAt, nil
	}
	lines := strings.Split(observationText, "\n")
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return "", createdAt, fmt.Errorf("observation missing header line")
	}
	var bodyLines []string
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "Session:") {
			break
		}
		bodyLines = append(bodyLines, lines[i])
	}
	body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	if strings.HasPrefix(body, "{") {
		return body, createdAt, nil
	}
	js := strings.Index(body, "{")
	je := strings.LastIndex(body, "}")
	if js >= 0 && je > js {
		body = body[js : je+1]
	}
	return body, createdAt, nil
}

// ParseSaveID reads numeric id from mem_save JSON envelope text.
func ParseSaveID(resultText string) (string, error) {
	resultText = strings.TrimSpace(resultText)
	var env struct {
		ID json.Number `json:"id"`
	}
	if err := json.Unmarshal([]byte(resultText), &env); err == nil && env.ID.String() != "" {
		return env.ID.String(), nil
	}
	re := regexp.MustCompile(`"id"\s*:\s*(\d+)`)
	if m := re.FindStringSubmatch(resultText); len(m) == 2 {
		return m[1], nil
	}
	return "", fmt.Errorf("save envelope missing id in %q", resultText)
}

// IsUnknownProject detects empty-store mem_search error.
func IsUnknownProject(errText string) bool {
	return strings.Contains(errText, "unknown_project")
}

// ParseSearchResults extracts results array from mem_search MCP text.
func ParseSearchResults(resultText string) ([]struct {
	ID       json.Number `json:"id"`
	TopicKey string      `json:"topic_key"`
	Title    string      `json:"title"`
}, error) {
	resultText = strings.TrimSpace(resultText)
	var wrap struct {
		Results []struct {
			ID       json.Number `json:"id"`
			TopicKey string      `json:"topic_key"`
			Title    string      `json:"title"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(resultText), &wrap); err == nil && wrap.Results != nil {
		return wrap.Results, nil
	}
	if err := json.Unmarshal([]byte(resultText), &wrap); err == nil {
		return wrap.Results, nil
	}
	return nil, fmt.Errorf("unparseable search results")
}
