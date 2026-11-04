package main

import "testing"

func TestOpenLaya(t *testing.T) {
	t.Setenv("LAYA_API_KEY", "  ")
	t.Setenv("LAYA_BASE_URL", "http://example.com:1")
	if c, err := openLaya(); err != nil || c == nil {
		t.Fatalf("empty key must still start: %v", err)
	}
	t.Setenv("LAYA_API_KEY", "k-test")
	if _, err := openLaya(); err == nil {
		t.Fatal("non-loopback url accepted with key set")
	}
	t.Setenv("LAYA_BASE_URL", "")
	if c, err := openLaya(); err != nil || c == nil {
		t.Fatal(err)
	}
}
