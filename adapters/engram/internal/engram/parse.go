package engram

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var createdLine = regexp.MustCompile(`(?m)^Created:\s*(.+)$`)

// ParseObservationBody extracts stored JSON body from mem_get_observation text.
// Body is after the `#id [type] title` line until a `Session:` line.
func ParseObservationBody(observationText string) (body string, createdAt string, err error) {
	lines := strings.Split(observationText, "\n")
	start := -1
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return "", "", fmt.Errorf("observation missing header line")
	}
	var bodyLines []string
	for i := start; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "Session:") {
			break
		}
		bodyLines = append(bodyLines, lines[i])
	}
	body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
	if m := createdLine.FindStringSubmatch(observationText); len(m) == 2 {
		createdAt = strings.TrimSpace(m[1])
	}
	return body, createdAt, nil
}

// ParseSaveID reads numeric id from mem_save JSON envelope text.
func ParseSaveID(resultText string) (string, error) {
	var env struct {
		ID json.Number `json:"id"`
	}
	if err := json.Unmarshal([]byte(resultText), &env); err != nil {
		return "", fmt.Errorf("save envelope: %w", err)
	}
	if env.ID.String() == "" {
		return "", fmt.Errorf("save envelope missing id")
	}
	return env.ID.String(), nil
}

// IsUnknownProject detects empty-store mem_search error.
func IsUnknownProject(errText string) bool {
	return strings.Contains(errText, "unknown_project")
}
