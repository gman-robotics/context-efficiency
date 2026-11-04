package laya

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestSystemOnePostsVerdict(t *testing.T) {
	var got Request
	var auth, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		path = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		_, _ = io.WriteString(w, `{"model":"english","answers":{"verdict":{"type":"choice","choice":"go","confidence":0.5,"answer_confidence":0.25}},"usage":{"input_tokens":3,"output_tokens":1}}`)
	}))
	defer srv.Close()

	c, err := NewClient(srv.URL, "test-key")
	if err != nil {
		t.Fatal(err)
	}
	if c.httpClient.Timeout != requestTimeout {
		t.Fatalf("timeout %s", c.httpClient.Timeout)
	}
	parsed, err := c.SystemOne(context.Background(), "abc", map[string]string{"go": "yes", "no_go": "no"})
	if err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer test-key" {
		t.Fatalf("auth %q", auth)
	}
	if path != "/v1/systemone" {
		t.Fatalf("path %s", path)
	}
	if got.State.Body != "abc" || got.Model != "english" {
		t.Fatalf("body %#v", got)
	}
	q, ok := got.Questions["verdict"]
	if !ok || q.Type != "choice" || q.Instructions != Instructions {
		t.Fatalf("question %#v", q)
	}
	if q.Criteria["go"] != "yes" || q.Criteria["no_go"] != "no" {
		t.Fatalf("criteria %#v", q.Criteria)
	}
	if parsed.Choice != "go" || parsed.Confidence != 0.5 || parsed.AnswerConfidence != 0.25 || parsed.Model != "english" {
		t.Fatalf("parsed %#v", parsed)
	}
}

func TestSystemOneURLKeepsSchemeSlash(t *testing.T) {
	c, err := NewClient("http://127.0.0.1:8792/", "")
	if err != nil {
		t.Fatal(err)
	}
	got, err := c.systemOneURL()
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://127.0.0.1:8792/v1/systemone" {
		t.Fatalf("url %s", got)
	}
	v6, err := NewClient("http://[::1]:8792", "")
	if err != nil {
		t.Fatal(err)
	}
	got, err = v6.systemOneURL()
	if err != nil {
		t.Fatal(err)
	}
	if got != "http://[::1]:8792/v1/systemone" {
		t.Fatalf("v6 url %s", got)
	}
}

func TestEmptyKeyDoesNotDial(t *testing.T) {
	c, err := NewClient("http://127.0.0.1:9", "")
	if err != nil {
		t.Fatal(err)
	}
	c.httpClient.Transport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("dialed")
		return nil, errors.New("dialed")
	})
	_, err = c.SystemOne(context.Background(), "hello-body", map[string]string{"go": "a", "no_go": "b"})
	if !errors.Is(err, ErrKey) {
		t.Fatalf("err %v", err)
	}
}

func TestNonLoopbackDoesNotDial(t *testing.T) {
	for _, raw := range []string{
		"http://127.0.0.1.example/v1",
		"http://127.0.0.1:8792@evil.example/",
		"http://10.1.2.3:8792",
		"http://localhost:8792",
		"http://127.0.0.1:8792/v1",
		"ftp://127.0.0.1:8792",
	} {
		_, err := NewClient(raw, "k")
		if err == nil {
			t.Fatalf("accepted %s", raw)
		}
		if strings.Contains(err.Error(), "evil.example") || strings.Contains(err.Error(), "127.0.0.1.example") || strings.Contains(err.Error(), "10.1.2.3") {
			t.Fatalf("url leaked in %v", err)
		}
	}
	c := &Client{
		baseURL: "http://10.1.2.3:8792",
		apiKey:  "k",
		httpClient: &http.Client{Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			t.Fatal("dialed")
			return nil, errors.New("dialed")
		})},
	}
	_, err := c.SystemOne(context.Background(), "hello-body", map[string]string{"go": "a", "no_go": "b"})
	if err == nil {
		t.Fatal("expected reject")
	}
	if strings.Contains(err.Error(), "10.1.2.3") || strings.Contains(err.Error(), "hello-body") {
		t.Fatalf("leaked %v", err)
	}
}

func TestLoopbackHostsAccepted(t *testing.T) {
	for _, raw := range []string{
		"",
		"http://127.0.0.1:8792",
		"http://127.0.0.2:8792",
		"https://127.0.0.1",
		"http://[::1]:8792",
	} {
		if _, err := NewClient(raw, ""); err != nil {
			t.Fatalf("%q: %v", raw, err)
		}
	}
	c, err := NewClient("", "")
	if err != nil {
		t.Fatal(err)
	}
	if c.baseURL != defaultBaseURL {
		t.Fatalf("base %s", c.baseURL)
	}
}

func TestRedirectDoesNotForwardBearer(t *testing.T) {
	hit := 0
	downstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit++
		if r.Header.Get("Authorization") != "" {
			t.Fatal("bearer forwarded")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer downstream.Close()
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer secret-key" {
			t.Fatalf("upstream auth %q", r.Header.Get("Authorization"))
		}
		http.Redirect(w, r, downstream.URL+"/landed", http.StatusFound)
	}))
	defer upstream.Close()

	c, err := NewClient(upstream.URL, "secret-key")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.SystemOne(context.Background(), "hello-body", map[string]string{"go": "a", "no_go": "b"})
	if err == nil {
		t.Fatal("expected redirect error")
	}
	if hit != 0 {
		t.Fatalf("downstream hits %d", hit)
	}
	if strings.Contains(err.Error(), "secret-key") || strings.Contains(err.Error(), "hello-body") {
		t.Fatalf("leaked %v", err)
	}
	if !errors.Is(err, ErrTransport) {
		t.Fatalf("err %v", err)
	}
}

func TestHTTPErrorHidesKeyAndBody(t *testing.T) {
	const key = "super-secret-key"
	const body = "state-sentinel-value"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"detail":"nope `+key+` and `+body+`"}`)
	}))
	defer srv.Close()
	c, err := NewClient(srv.URL, key)
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.SystemOne(context.Background(), body, map[string]string{"go": "a", "no_go": "b"})
	if err == nil {
		t.Fatal("expected 401")
	}
	msg := err.Error()
	if !strings.Contains(msg, "laya http 401:") {
		t.Fatalf("msg %s", msg)
	}
	if strings.Contains(msg, key) || strings.Contains(msg, body) || strings.Contains(msg, "Bearer") {
		t.Fatalf("leaked %s", msg)
	}
}

func TestTruncation(t *testing.T) {
	cases := []struct {
		name string
		body string
		cut  bool
	}{
		{name: "absent", body: `{"model":"english","answers":{"verdict":{"choice":"go","confidence":0.5,"answer_confidence":0.25}},"usage":{}}`, cut: false},
		{name: "false", body: `{"model":"english","answers":{"verdict":{"choice":"go","confidence":0.5,"answer_confidence":0.25}},"usage":{"truncated":false}}`, cut: false},
		{name: "zero", body: `{"model":"english","answers":{"verdict":{"choice":"go","confidence":0.5,"answer_confidence":0.25}},"usage":{"truncated":0}}`, cut: false},
		{name: "null", body: `{"model":"english","answers":{"verdict":{"choice":"go","confidence":0.5,"answer_confidence":0.25}},"usage":{"truncated":null}}`, cut: false},
		{name: "text", body: `{"model":"english","answers":{"verdict":{"choice":"go","confidence":0.5,"answer_confidence":0.25}},"usage":{"truncated":"no"}}`, cut: false},
		{name: "true", body: `{"model":"english","answers":{"verdict":{"choice":"go","confidence":0.5,"answer_confidence":0.25}},"usage":{"truncated":true}}`, cut: true},
		{name: "two", body: `{"model":"english","answers":{"verdict":{"choice":"go","confidence":0.5,"answer_confidence":0.25}},"usage":{"truncated":2}}`, cut: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = io.WriteString(w, tc.body)
			}))
			defer srv.Close()
			c, err := NewClient(srv.URL, "k")
			if err != nil {
				t.Fatal(err)
			}
			_, err = c.SystemOne(context.Background(), "hello-body", map[string]string{"go": "a", "no_go": "b"})
			if tc.cut {
				if !errors.Is(err, ErrTruncated) {
					t.Fatalf("err %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestTimeoutConstant(t *testing.T) {
	if requestTimeout != 60*time.Second {
		t.Fatalf("timeout %s", requestTimeout)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
