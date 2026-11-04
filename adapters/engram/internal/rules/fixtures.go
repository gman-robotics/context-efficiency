package rules

import (
	"embed"
	"encoding/json"
	"fmt"
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
	fill, ok := fillers[path]
	if !ok {
		return Case{}, fmt.Errorf("fixture path")
	}
	c := Case{ID: row.ID, Path: path, Actual: row.Actual}
	if !fill(&c, row) {
		return Case{}, fmt.Errorf("fixture %s missing %s", row.ID, featureKey[path])
	}
	return c, nil
}

var featureKey = map[string]string{PathGo: "go", PathPR: "pr", PathStop: "stop", PathAudit: "audit"}

// fillers copy one path's features into c and render its body. False means the features are missing.
var fillers = map[string]func(c *Case, row rawFixture) bool{
	PathGo: func(c *Case, row rawFixture) bool {
		if row.Go == nil {
			return false
		}
		c.Go, c.Action = *row.Go, row.Go.Action
		c.Body = RenderGo(c.Go)
		return true
	},
	PathPR: func(c *Case, row rawFixture) bool {
		if row.PR == nil {
			return false
		}
		c.PR = *row.PR
		c.Body = RenderPR(c.PR)
		return true
	},
	PathStop: func(c *Case, row rawFixture) bool {
		if row.Stop == nil {
			return false
		}
		c.Stop = *row.Stop
		c.Body = RenderStop(c.Stop)
		return true
	},
	PathAudit: func(c *Case, row rawFixture) bool {
		if row.Audit == nil {
			return false
		}
		c.Audit = *row.Audit
		c.Body = RenderAudit(c.Audit)
		return true
	},
}
