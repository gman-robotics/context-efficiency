package shadow

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/laya"
)

var requestIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

// Caller is the Laya transport.
type Caller interface {
	SystemOne(ctx context.Context, body string, criteria map[string]string) (laya.Parsed, error)
}

// Retainer is the Engram retain path. *adaptermap.Service implements it.
type Retainer interface {
	Retain(ctx context.Context, text string, tags []string, importance int, scopeStr string) (string, error)
}

// Input is one decide_shadow call.
type Input struct {
	Path      string
	StateBody string
	Actual    string
	Action    string
	Labels    []string
	RequestID string
}

// Outcome is the tool result. Effect is none.
type Outcome struct {
	Record
	LogID  string `json:"log_id"`
	Effect string `json:"effect"`
}

// Decider logs a shadow label. A nil Store skips retain.
type Decider struct {
	Caller Caller
	Store  Retainer
	Now    func() time.Time
}

// NewDecider builds a decider. store may be nil for the in-memory report.
func NewDecider(caller Caller, store Retainer) *Decider {
	return &Decider{Caller: caller, Store: store}
}

func (d *Decider) now() time.Time {
	if d != nil && d.Now != nil {
		return d.Now()
	}
	return time.Now()
}

// Decide screens the call, asks Laya, and retains metadata only.
// Caller-supplied actual is stored as given.
func (d *Decider) Decide(ctx context.Context, in Input) (Outcome, error) {
	if err := Screen(in.Path, in.Action, in.Actual, in.Labels, in.StateBody); err != nil {
		return Outcome{}, err
	}
	criteria, err := Criteria(in.Path)
	if err != nil {
		return Outcome{}, err
	}
	requestID, err := normalizeRequestID(in.RequestID)
	if err != nil {
		return Outcome{}, errors.New("laya request id")
	}
	if d == nil || d.Caller == nil {
		return Outcome{}, errors.New("laya decider not set")
	}
	parsed, err := d.Caller.SystemOne(ctx, in.StateBody, criteria)
	if err != nil {
		return Outcome{}, err
	}
	if !verdictOK(in.Path, parsed.Choice) {
		return Outcome{}, errors.New("laya verdict choice rejected")
	}
	rec := Record{
		Kind:             RecordKind,
		Path:             in.Path,
		RequestID:        requestID,
		StateHash:        StateHash(in.StateBody),
		Actual:           in.Actual,
		Shadow:           parsed.Choice,
		Agree:            parsed.Choice == in.Actual,
		Confidence:       parsed.Confidence,
		AnswerConfidence: parsed.AnswerConfidence,
		Model:            parsed.Model,
		DecidedAt:        d.now().UTC().Format(time.RFC3339Nano),
	}
	out := Outcome{Record: rec, Effect: "none"}
	if d.Store == nil {
		return out, nil
	}
	raw, err := json.Marshal(rec)
	if err != nil {
		return Outcome{}, errors.New("laya log: marshal")
	}
	id, err := d.Store.Retain(ctx, string(raw), []string{TagShadow, in.Path}, Importance, ScopeShadow)
	if err != nil {
		return Outcome{}, fmt.Errorf("laya log: %s", safeErr(err, in.StateBody))
	}
	out.LogID = id
	return out, nil
}

func normalizeRequestID(raw string) (string, error) {
	if requestIDPattern.MatchString(raw) {
		return raw, nil
	}
	return randomHexID()
}

func randomHexID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
