package shadow

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/laya"
)

//go:embed fixtures/*.json
var fixtureFS embed.FS

var fixtureOrder = []string{
	"fixtures/routine_go_nogo.json",
	"fixtures/pr_class.json",
	"fixtures/tool_failure_stop.json",
	"fixtures/audit_sort.json",
}

type rawFixture struct {
	ID     string         `json:"id"`
	Actual string         `json:"actual"`
	Go     *GoFeatures    `json:"go"`
	PR     *PRFeatures    `json:"pr"`
	Stop   *StopFeatures  `json:"stop"`
	Audit  *AuditFeatures `json:"audit"`
}

// LoadFixtures reads the embedded rows in path order and renders each state body.
func LoadFixtures() ([]Case, error) {
	var out []Case
	for _, name := range fixtureOrder {
		raw, err := fixtureFS.ReadFile(name)
		if err != nil {
			return nil, err
		}
		var rows []rawFixture
		if err := json.Unmarshal(raw, &rows); err != nil {
			return nil, fmt.Errorf("fixture %s: %w", name, err)
		}
		path := pathByFile(name)
		for _, row := range rows {
			c, err := caseFor(path, row)
			if err != nil {
				return nil, err
			}
			out = append(out, c)
		}
	}
	return out, nil
}

func pathByFile(name string) string {
	switch name {
	case "fixtures/routine_go_nogo.json":
		return PathGo
	case "fixtures/pr_class.json":
		return PathPR
	case "fixtures/tool_failure_stop.json":
		return PathStop
	case "fixtures/audit_sort.json":
		return PathAudit
	default:
		return ""
	}
}

func caseFor(path string, row rawFixture) (Case, error) {
	if row.ID == "" {
		return Case{}, fmt.Errorf("fixture row missing id")
	}
	c := Case{ID: row.ID, Path: path, Actual: row.Actual}
	switch path {
	case PathGo:
		if row.Go == nil {
			return Case{}, fmt.Errorf("fixture %s missing go", row.ID)
		}
		c.Go = *row.Go
		c.Action = row.Go.Action
		c.Body = RenderGo(c.Go)
	case PathPR:
		if row.PR == nil {
			return Case{}, fmt.Errorf("fixture %s missing pr", row.ID)
		}
		c.PR = *row.PR
		c.Body = RenderPR(c.PR)
	case PathStop:
		if row.Stop == nil {
			return Case{}, fmt.Errorf("fixture %s missing stop", row.ID)
		}
		c.Stop = *row.Stop
		c.Body = RenderStop(c.Stop)
	case PathAudit:
		if row.Audit == nil {
			return Case{}, fmt.Errorf("fixture %s missing audit", row.ID)
		}
		c.Audit = *row.Audit
		c.Body = RenderAudit(c.Audit)
	default:
		return Case{}, fmt.Errorf("fixture path")
	}
	return c, nil
}

// RunFixtures loads the embedded rows and runs them. The bool is true when Laya is down.
func RunFixtures(ctx context.Context, caller Caller, now time.Time) (Report, bool, error) {
	cases, err := LoadFixtures()
	if err != nil {
		return Report{}, false, err
	}
	return RunCases(ctx, caller, cases, now)
}

// RunCases refuses to call Laya when a fixture actual disagrees with Rule.
// A key or transport error with zero successes is down. Other Laya errors become items.
func RunCases(ctx context.Context, caller Caller, cases []Case, now time.Time) (Report, bool, error) {
	dec := NewDecider(caller, nil)
	dec.Now = func() time.Time { return now }
	by := map[string][]Item{
		PathGo:    {},
		PathPR:    {},
		PathStop:  {},
		PathAudit: {},
	}
	successes := 0
	for _, c := range cases {
		want, err := Rule(c)
		if err != nil {
			return Report{}, false, fmt.Errorf("fixture rule: %s", c.ID)
		}
		if want != c.Actual {
			return Report{}, false, fmt.Errorf("fixture actual drift: %s", c.ID)
		}
		out, err := dec.Decide(ctx, Input{
			Path:      c.Path,
			StateBody: c.Body,
			Actual:    c.Actual,
			Action:    c.Action,
			Labels:    c.Labels,
		})
		if err != nil {
			if successes == 0 && errors.Is(err, laya.ErrKey) {
				return Report{}, true, err
			}
			if successes == 0 && errors.Is(err, laya.ErrTransport) {
				return Report{}, true, nil
			}
			if strings.HasPrefix(err.Error(), "refused:") {
				return Report{}, false, fmt.Errorf("fixture refused: %s", c.ID)
			}
			by[c.Path] = append(by[c.Path], Item{
				ID:        c.ID,
				Actual:    c.Actual,
				StateHash: StateHash(c.Body),
				Agree:     false,
				Err:       safeErr(err, c.Body),
			})
			continue
		}
		successes++
		by[c.Path] = append(by[c.Path], Item{
			ID:               c.ID,
			Actual:           out.Actual,
			Shadow:           out.Shadow,
			Confidence:       out.Confidence,
			AnswerConfidence: out.AnswerConfidence,
			Model:            out.Model,
			StateHash:        out.StateHash,
			Agree:            out.Agree,
		})
	}
	return BuildReport(now, by), false, nil
}
