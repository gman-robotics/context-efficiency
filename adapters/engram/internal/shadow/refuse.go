package shadow

import (
	"errors"
	"regexp"
	"strings"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/content"
)

var (
	nonAlnum = regexp.MustCompile(`[^a-z0-9]+`)
	digitRun = regexp.MustCompile(`[0-9]{7,}`)
)

var verdicts = map[string]map[string]struct{}{
	PathGo: {
		"go":    {},
		"no_go": {},
	},
	PathPR: {
		"T0": {},
		"T1": {},
		"T2": {},
		"T3": {},
	},
	PathStop: {
		"continue": {},
		"stop":     {},
	},
	PathAudit: {
		"blocker":    {},
		"should_fix": {},
		"nit":        {},
		"note":       {},
	},
}

// DeniedToken returns the first denied token in s, or empty.
// Prefixes send, merge, deploy, and pay match. sender and payload do not.
// charge, refund, wire, invoice, and money match only as whole tokens.
func DeniedToken(s string) string {
	for _, tok := range tokensOf(s) {
		if tokenDenied(tok) {
			return tok
		}
	}
	return ""
}

func tokensOf(s string) []string {
	low := strings.ToLower(s)
	parts := nonAlnum.Split(low, -1)
	var out []string
	seen := map[string]struct{}{}
	add := func(tok string) {
		if tok == "" {
			return
		}
		if _, ok := seen[tok]; ok {
			return
		}
		seen[tok] = struct{}{}
		out = append(out, tok)
	}
	for _, part := range parts {
		add(part)
	}
	add(strings.TrimSpace(low))
	return out
}

func tokenDenied(tok string) bool {
	switch tok {
	case "sender", "payload":
		return false
	}
	for _, prefix := range []string{"send", "merge", "deploy", "pay"} {
		if strings.HasPrefix(tok, prefix) {
			return true
		}
	}
	switch tok {
	case "charge", "refund", "wire", "invoice", "money":
		return true
	default:
		return false
	}
}

func verdictOK(path, actual string) bool {
	set, ok := verdicts[path]
	if !ok {
		return false
	}
	_, ok = set[actual]
	return ok
}

// Refuse rejects a path, action, actual, or label that is denied or outside the allow-list.
// State text is not scanned here.
func Refuse(path, action, actual string, labels []string) error {
	fields := make([]string, 0, 3+len(labels))
	fields = append(fields, path, action, actual)
	fields = append(fields, labels...)
	for _, field := range fields {
		if DeniedToken(field) != "" {
			return errors.New("refused: denied token")
		}
	}
	if _, ok := verdicts[path]; !ok {
		return errors.New("refused: path")
	}
	if !verdictOK(path, actual) {
		return errors.New("refused: actual")
	}
	return nil
}

// Screen is Refuse plus the state-body checks. Failure means no HTTP and no save.
func Screen(path, action, actual string, labels []string, body string) error {
	if err := Refuse(path, action, actual, labels); err != nil {
		return err
	}
	if strings.TrimSpace(body) == "" {
		return errors.New("refused: empty body")
	}
	if len(body) > 2000 {
		return errors.New("refused: body too large")
	}
	if strings.Contains(body, "@") || digitRun.MatchString(body) {
		return errors.New("refused: body")
	}
	fields := make([]string, 0, 3+len(labels))
	fields = append(fields, body, actual, action)
	fields = append(fields, labels...)
	for _, field := range fields {
		if err := content.RejectPrivateInString(field); err != nil {
			return errors.New("refused: private")
		}
	}
	return nil
}
