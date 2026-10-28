package startup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gman-robotics/context-efficiency/adapters/engram/internal/pin"
)

// ResolveEngramBinary returns path to engram after optional ENGRAM_BIN override and pin hash check.
func ResolveEngramBinary(exeDir string, pinPath string) (string, error) {
	p, err := pin.Load(pinPath)
	if err != nil {
		return "", err
	}
	if p.BinarySHA256 == "" {
		return "", fmt.Errorf("pin missing binary_sha256")
	}
	candidate := os.Getenv("ENGRAM_BIN")
	if candidate == "" {
		candidate = filepath.Join(exeDir, "engram")
	}
	if err := pin.VerifyBinary(candidate, p.BinarySHA256); err != nil {
		return "", err
	}
	return candidate, nil
}

func PinPath(exe string) string {
	return filepath.Join(filepath.Dir(exe), "pin.json")
}
