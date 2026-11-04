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
