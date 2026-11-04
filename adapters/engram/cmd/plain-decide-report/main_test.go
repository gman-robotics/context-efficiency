package main

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWritesReport(t *testing.T) {
	dir := t.TempDir()
	if code := run(dir, io.Discard); code != 0 {
		t.Fatalf("exit %d", code)
	}
	for _, f := range []string{"plain-decide-agreement.json", "plain-decide-agreement.md"} {
		if _, err := os.Stat(filepath.Join(dir, f)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRunWriteError(t *testing.T) {
	f := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(f, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if code := run(filepath.Join(f, "sub"), io.Discard); code != 1 {
		t.Fatalf("exit %d", code)
	}
}
