package startup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveEngramBinaryMissingBinarySHA256(t *testing.T) {
	dir := t.TempDir()
	pinPath := filepath.Join(dir, "pin.json")
	body := []byte("{\n  \"version\": \"3.6.0\",\n  \"binary_sha256\": \"\"\n}\n")
	if err := os.WriteFile(pinPath, body, 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ResolveEngramBinary(dir, pinPath)
	if err == nil || err.Error() != "pin missing binary_sha256" {
		t.Fatalf("err %v", err)
	}
}
