package laya

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateBaseURLRejects(t *testing.T) {
	for _, raw := range []string{
		"http://u:p@127.0.0.1:8792",
		"ftp://127.0.0.1:8792",
		"http://127.0.0.1:8792/v1",
		"http://127.0.0.1:8792/?q=1",
		"http://127.0.0.1:8792/#f",
		"http://",
		"http://example.com:8792",
	} {
		if err := validateBaseURL(raw); err == nil {
			t.Fatalf("accepted %q", raw)
		}
	}
	if err := validateBaseURL("http://127.0.0.1:8792/"); err != nil {
		t.Fatal(err)
	}
}

func serve(t *testing.T, status int, body string) *Client {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient(srv.URL, "k-test-secret")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestSystemOneResponseErrors(t *testing.T) {
	cases := []struct {
		name, body, want string
		status           int
	}{
		{"unreadable", "not json", "laya response unreadable", 200},
		{"missing", `{"answers":{}}`, "laya verdict choice missing", 200},
		{"blank", `{"answers":{"verdict":{"choice":"  "}}}`, "laya verdict choice missing", 200},
		{"detail object", `{"detail":{"msg":"bad"}}`, `laya http 422: {"msg":"bad"}`, 422},
		{"detail string", `{"detail":"nope"}`, "laya http 400: nope", 400},
		{"plain", `oops`, "laya http 500: oops", 500},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := serve(t, tc.status, tc.body)
			_, err := c.SystemOne(nil, "state body text", map[string]string{"go": "g"}) //nolint:staticcheck
			if err == nil || err.Error() != tc.want {
				t.Fatalf("got %v want %q", err, tc.want)
			}
		})
	}
}

func TestSystemOneTransportErrorScrubbed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()
	c, err := NewClient(url, "k-test-secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = c.SystemOne(context.Background(), "state body text", nil)
	if !errors.Is(err, ErrTransport) || strings.Contains(err.Error(), "k-test-secret") {
		t.Fatal(err)
	}
}

func TestDetailTextFallsBack(t *testing.T) {
	if got := detailText([]byte(`{"detail":null}`)); got != `{"detail":null}` {
		t.Fatal(got)
	}
	if got := detailText([]byte(`{"detail":["a"]}`)); got != `["a"]` {
		t.Fatal(got)
	}
}
