package shadow

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/adaptermap"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/child"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/domain"
	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/laya"
)

var _ Retainer = (*adaptermap.Service)(nil)

var hexID = regexp.MustCompile(`^[0-9a-f]{32}$`)

type spyEngram struct {
	*child.FakeEngram
	saves        int
	obsIDs       []interface{}
	searchLimits []interface{}
	saveArgs     []map[string]interface{}
}

func (s *spyEngram) CallTool(ctx context.Context, name string, args map[string]interface{}) (string, bool, error) {
	switch name {
	case "mem_save":
		s.saves++
		cp := map[string]interface{}{}
		for k, v := range args {
			cp[k] = v
		}
		s.saveArgs = append(s.saveArgs, cp)
	case "mem_get_observation":
		s.obsIDs = append(s.obsIDs, args["id"])
	case "mem_search":
		s.searchLimits = append(s.searchLimits, args["limit"])
	}
	return s.FakeEngram.CallTool(ctx, name, args)
}

type countCaller struct {
	n      int
	choice string
	err    error
}

func (c *countCaller) SystemOne(ctx context.Context, body string, criteria map[string]string) (laya.Parsed, error) {
	c.n++
	if c.err != nil {
		return laya.Parsed{}, c.err
	}
	choice := c.choice
	if choice == "" {
		choice = "go"
	}
	return laya.Parsed{Choice: choice, Confidence: 0.5, AnswerConfidence: 0.25, Model: "english"}, nil
}

func TestDecideRetainsMetadataAndReadsBack(t *testing.T) {
	const body = "STATEBODYMARKER routine gate text for the shadow record"
	fake := &child.FakeEngram{}
	spy := &spyEngram{FakeEngram: fake}
	svc := &adaptermap.Service{Client: spy}
	when := time.Date(2026, 4, 1, 3, 4, 5, 6, time.UTC)
	caller := &countCaller{choice: "go"}
	d := NewDecider(caller, svc)
	d.Now = func() time.Time { return when }
	out, err := d.Decide(context.Background(), Input{
		Path:      PathGo,
		StateBody: body,
		Actual:    "go",
		Action:    "write_file",
		RequestID: "req-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if caller.n != 1 {
		t.Fatalf("calls %d", caller.n)
	}
	if out.Effect != "none" || out.LogID == "" || !out.Agree || out.Shadow != "go" || out.Actual != "go" {
		t.Fatalf("%#v", out)
	}
	if out.RequestID != "req-1" || out.Model != "english" || out.Confidence != 0.5 || out.AnswerConfidence != 0.25 {
		t.Fatalf("%#v", out)
	}
	if out.StateHash != StateHash(body) || out.DecidedAt != when.Format(time.RFC3339Nano) || out.Kind != RecordKind {
		t.Fatalf("%#v", out)
	}
	if spy.saves != 1 || len(spy.saveArgs) != 1 {
		t.Fatalf("saves %d", spy.saves)
	}
	args := spy.saveArgs[0]
	if args["capture_prompt"] != false {
		t.Fatalf("capture %#v", args["capture_prompt"])
	}
	if _, ok := args["topic_key"]; ok {
		t.Fatal("topic_key set")
	}
	if fake.LastTopicKey() != "" {
		t.Fatalf("topic %q", fake.LastTopicKey())
	}
	content, _ := args["content"].(string)
	title, _ := args["title"].(string)
	if strings.Contains(content, body) || strings.Contains(title, body) || strings.Contains(content, "STATEBODYMARKER") || strings.Contains(title, "STATEBODYMARKER") {
		t.Fatal("state stored")
	}
	var rb domain.RetainBody
	if err := json.Unmarshal([]byte(content), &rb); err != nil {
		t.Fatal(err)
	}
	if rb.Scope != ScopeShadow || rb.Importance != Importance {
		t.Fatalf("%#v", rb)
	}
	if len(rb.Tags) != 2 || rb.Tags[0] != TagShadow || rb.Tags[1] != PathGo {
		t.Fatalf("tags %#v", rb.Tags)
	}
	var rec Record
	if err := json.Unmarshal([]byte(rb.Text), &rec); err != nil {
		t.Fatal(err)
	}
	if rec.StateHash != StateHash(body) || rec.Shadow != "go" || rec.Actual != "go" || !rec.Agree {
		t.Fatalf("%#v", rec)
	}
	items, err := svc.Recall(context.Background(), "laya_shadow", ScopeShadow, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Text != rb.Text || items[0].Scope != ScopeShadow {
		t.Fatalf("%#v", items)
	}
	if len(spy.searchLimits) != 1 || spy.searchLimits[0] != 20 {
		t.Fatalf("limits %#v", spy.searchLimits)
	}
	if len(spy.obsIDs) == 0 {
		t.Fatal("no observation id")
	}
	for _, id := range spy.obsIDs {
		if _, ok := id.(int64); !ok {
			t.Fatalf("id %T", id)
		}
	}
}

func TestMergeInStateIsPostedNotStored(t *testing.T) {
	const body = "please merge the notes in this state body"
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		var req laya.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(req.State.Body, "merge") {
			t.Fatal("state not forwarded")
		}
		_, _ = io.WriteString(w, `{"model":"english","answers":{"verdict":{"choice":"go","confidence":0.5,"answer_confidence":0.25}}}`)
	}))
	defer srv.Close()
	client, err := laya.NewClient(srv.URL, "k")
	if err != nil {
		t.Fatal(err)
	}
	fake := &child.FakeEngram{}
	spy := &spyEngram{FakeEngram: fake}
	d := NewDecider(client, &adaptermap.Service{Client: spy})
	out, err := d.Decide(context.Background(), Input{Path: PathGo, StateBody: body, Actual: "go", Action: "write_file"})
	if err != nil {
		t.Fatal(err)
	}
	if hits != 1 || !out.Agree {
		t.Fatalf("hits %d %#v", hits, out)
	}
	content, _ := spy.saveArgs[0]["content"].(string)
	title, _ := spy.saveArgs[0]["title"].(string)
	if strings.Contains(content, "merge") || strings.Contains(title, "merge") || strings.Contains(content, body) {
		t.Fatal("state stored")
	}
}

func TestRefusalSkipsHTTPAndSave(t *testing.T) {
	caller := &countCaller{}
	fake := &child.FakeEngram{}
	spy := &spyEngram{FakeEngram: fake}
	d := NewDecider(caller, &adaptermap.Service{Client: spy})
	cases := []Input{
		{Path: PathGo, StateBody: "hello state", Actual: "merge", Action: "write_file"},
		{Path: PathGo, StateBody: "", Actual: "go", Action: "write_file"},
		{Path: PathGo, StateBody: strings.Repeat("a", 2001), Actual: "go", Action: "write_file"},
		{Path: PathGo, StateBody: "see <private>x</private>", Actual: "go", Action: "write_file"},
		{Path: PathGo, StateBody: "user@host name", Actual: "go", Action: "write_file"},
		{Path: PathGo, StateBody: "id 1234567 end", Actual: "go", Action: "write_file"},
		{Path: "send", StateBody: "hello state", Actual: "go", Action: "write_file"},
	}
	for _, in := range cases {
		_, err := d.Decide(context.Background(), in)
		if err == nil || !strings.HasPrefix(err.Error(), "refused:") {
			t.Fatal(err)
		}
	}
	if caller.n != 0 || spy.saves != 0 {
		t.Fatalf("calls %d saves %d", caller.n, spy.saves)
	}
}

func TestServerErrorsDoNotRetain(t *testing.T) {
	const key = "super-secret-key"
	const body = "STATEBODYMARKER"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, `{"detail":"inference failed `+key+` `+body+`"}`)
	}))
	defer srv.Close()
	client, err := laya.NewClient(srv.URL, key)
	if err != nil {
		t.Fatal(err)
	}
	fake := &child.FakeEngram{}
	spy := &spyEngram{FakeEngram: fake}
	d := NewDecider(client, &adaptermap.Service{Client: spy})
	_, err = d.Decide(context.Background(), Input{Path: PathGo, StateBody: body, Actual: "go", Action: "write_file"})
	if err == nil || !strings.Contains(err.Error(), "laya http 500:") {
		t.Fatal(err)
	}
	if strings.Contains(err.Error(), key) || strings.Contains(err.Error(), body) || spy.saves != 0 {
		t.Fatalf("err %v saves %d", err, spy.saves)
	}

	trunc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"model":"english","answers":{"verdict":{"choice":"go","confidence":1,"answer_confidence":1}},"usage":{"truncated":true}}`)
	}))
	defer trunc.Close()
	client, err = laya.NewClient(trunc.URL, key)
	if err != nil {
		t.Fatal(err)
	}
	d = NewDecider(client, &adaptermap.Service{Client: spy})
	_, err = d.Decide(context.Background(), Input{Path: PathGo, StateBody: body, Actual: "go", Action: "write_file"})
	if !errors.Is(err, laya.ErrTruncated) || spy.saves != 0 {
		t.Fatal(err, spy.saves)
	}
}

func TestChoiceRejectedDoesNotRetainOrEcho(t *testing.T) {
	caller := &countCaller{choice: "maybe"}
	fake := &child.FakeEngram{}
	spy := &spyEngram{FakeEngram: fake}
	d := NewDecider(caller, &adaptermap.Service{Client: spy})
	_, err := d.Decide(context.Background(), Input{Path: PathGo, StateBody: "hello state", Actual: "go", Action: "write_file"})
	if err == nil || err.Error() != "laya verdict choice rejected" || strings.Contains(err.Error(), "maybe") {
		t.Fatal(err)
	}
	if spy.saves != 0 {
		t.Fatal("saved")
	}
}

func TestInvalidRequestIDReplaced(t *testing.T) {
	const rawID = "RAWREQUESTIDVALUE"
	caller := &countCaller{choice: "go"}
	fake := &child.FakeEngram{}
	spy := &spyEngram{FakeEngram: fake}
	d := NewDecider(caller, &adaptermap.Service{Client: spy})
	out, err := d.Decide(context.Background(), Input{
		Path: PathGo, StateBody: "hello state", Actual: "go", Action: "write_file", RequestID: rawID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.RequestID == rawID || !hexID.MatchString(out.RequestID) {
		t.Fatalf("id %q", out.RequestID)
	}
	content, _ := spy.saveArgs[0]["content"].(string)
	if strings.Contains(content, rawID) {
		t.Fatal("raw id stored")
	}
}

func TestDisagreementStillRetains(t *testing.T) {
	caller := &countCaller{choice: "no_go"}
	fake := &child.FakeEngram{}
	spy := &spyEngram{FakeEngram: fake}
	d := NewDecider(caller, &adaptermap.Service{Client: spy})
	out, err := d.Decide(context.Background(), Input{Path: PathGo, StateBody: "hello state", Actual: "go", Action: "write_file"})
	if err != nil {
		t.Fatal(err)
	}
	if out.Agree || out.Shadow != "no_go" || out.LogID == "" || spy.saves != 1 {
		t.Fatalf("%#v saves %d", out, spy.saves)
	}
}

func TestRetainFailureIsLogError(t *testing.T) {
	caller := &countCaller{choice: "go"}
	d := NewDecider(caller, errStore{})
	const body = "STATEBODYMARKER"
	_, err := d.Decide(context.Background(), Input{Path: PathGo, StateBody: body, Actual: "go", Action: "write_file"})
	if err == nil || !strings.HasPrefix(err.Error(), "laya log:") || strings.HasPrefix(err.Error(), "refused:") {
		t.Fatal(err)
	}
	if strings.Contains(err.Error(), body) {
		t.Fatal(err)
	}
}

type errStore struct{}

func (errStore) Retain(ctx context.Context, text string, tags []string, importance int, scopeStr string) (string, error) {
	return "", errors.New("disk " + text)
}

func TestEmptyKeyDoesNotHit(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
	}))
	defer srv.Close()
	client, err := laya.NewClient(srv.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	fake := &child.FakeEngram{}
	spy := &spyEngram{FakeEngram: fake}
	d := NewDecider(client, &adaptermap.Service{Client: spy})
	_, err = d.Decide(context.Background(), Input{Path: PathGo, StateBody: "hello state", Actual: "go", Action: "write_file"})
	if !errors.Is(err, laya.ErrKey) || hits != 0 || spy.saves != 0 {
		t.Fatal(err, hits, spy.saves)
	}
}
