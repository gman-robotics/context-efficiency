package shadow

import (
	"crypto/sha256"
	"encoding/hex"
)

const (
	RecordKind  = "laya_shadow"
	TagShadow   = "laya-shadow"
	ScopeShadow = "routine:layashadow"
	Importance  = 1
)

// Record is the retain text. It has no state body.
type Record struct {
	Kind             string  `json:"kind"`
	Path             string  `json:"path"`
	RequestID        string  `json:"request_id"`
	StateHash        string  `json:"state_hash"`
	Actual           string  `json:"actual"`
	Shadow           string  `json:"shadow"`
	Agree            bool    `json:"agree"`
	Confidence       float64 `json:"confidence"`
	AnswerConfidence float64 `json:"answer_confidence"`
	Model            string  `json:"model"`
	DecidedAt        string  `json:"decided_at"`
}

// StateHash is the hex sha256 of the state body.
func StateHash(body string) string {
	sum := sha256.Sum256([]byte(body))
	return hex.EncodeToString(sum[:])
}
