package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunExitsBeforeReport(t *testing.T) {
	closed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	closedURL := closed.URL
	closed.Close()
	cases := []struct {
		name, base, key, want string
	}{
		{"bad url", "http://example.com:8792", "k-secret", "laya base url rejected"},
		{"no key", "", "", "laya api key not set"},
		{"down", closedURL, "k-secret", "laya down"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			var errOut bytes.Buffer
			if code := run(dir, tc.base, tc.key, &errOut); code != 1 {
				t.Fatalf("code %d", code)
			}
			got := strings.TrimSpace(errOut.String())
			if got != tc.want || strings.Contains(got, "k-secret") {
				t.Fatal(got)
			}
		})
	}
}
