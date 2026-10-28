package scope

import (
	"fmt"
	"regexp"
	"strings"
)

const tokenLen = 40

var namePart = regexp.MustCompile(`^[a-z0-9]{1,40}$`)

// DefaultScope when callers omit scope.
const DefaultScope = "bot"

// Token returns the fixed 40-character scope token for Engram title/query filtering.
func Token(scope string) (string, error) {
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = DefaultScope
	}
	kind, payload, err := parseScope(scope)
	if err != nil {
		return "", err
	}
	body := string([]byte{kind}) + payload
	if len(body) > tokenLen {
		return "", fmt.Errorf("scope token payload too long")
	}
	return body + strings.Repeat("0", tokenLen-len(body)), nil
}

func parseScope(scope string) (kind byte, payload string, err error) {
	switch {
	case scope == "bot":
		return 'b', "cesbot", nil
	case scope == "shared":
		return 's', "cesshared", nil
	case strings.HasPrefix(scope, "routine:"):
		name := strings.TrimPrefix(scope, "routine:")
		if !namePart.MatchString(name) {
			return 0, "", fmt.Errorf("invalid routine scope name")
		}
		return 'r', "cesroutine" + name, nil
	case strings.HasPrefix(scope, "task:"):
		id := strings.TrimPrefix(scope, "task:")
		if !namePart.MatchString(id) {
			return 0, "", fmt.Errorf("invalid task scope id")
		}
		return 't', "cestask" + id, nil
	default:
		return 0, "", fmt.Errorf("unsupported scope: %s", scope)
	}
}

// HandoffTopicKey is normalized handoff/<token> (token is 40 chars, no extra slash).
func HandoffTopicKey(token string) string {
	return "handoff/" + strings.TrimSpace(token)
}

// TitleFirstField returns the first whitespace-separated field of a mem title.
func TitleFirstField(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return ""
	}
	i := strings.IndexByte(title, ' ')
	if i < 0 {
		return title
	}
	return title[:i]
}
