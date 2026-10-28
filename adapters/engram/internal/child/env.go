package child

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// Config for spawning the engram child.
type Config struct {
	DataDir     string
	Project     string
	SentinelHome string
	EngramBin   string
}

func ValidateStartup(dataDir, project string) error {
	if strings.TrimSpace(dataDir) == "" {
		return fmt.Errorf("ENGRAM_DATA_DIR missing")
	}
	if !filepath.IsAbs(dataDir) {
		return fmt.Errorf("ENGRAM_DATA_DIR must be absolute")
	}
	if strings.TrimSpace(project) == "" {
		return fmt.Errorf("ENGRAM_PROJECT missing")
	}
	if strings.Contains(project, "/") || strings.Contains(project, "\\") {
		return fmt.Errorf("ENGRAM_PROJECT must not contain path separators")
	}
	for _, r := range project {
		if unicode.IsControl(r) {
			return fmt.Errorf("ENGRAM_PROJECT must not contain control characters")
		}
	}
	return nil
}

// ChildEnv returns the explicit environment list for the engram child (no host inheritance).
func ChildEnv(cfg Config) ([]string, error) {
	if err := ValidateStartup(cfg.DataDir, cfg.Project); err != nil {
		return nil, err
	}
	home := cfg.SentinelHome
	if home == "" {
		return nil, fmt.Errorf("sentinel home missing")
	}
	if !filepath.IsAbs(home) {
		return nil, fmt.Errorf("sentinel home must be absolute")
	}
	if filepath.Clean(home) == filepath.Clean(cfg.DataDir) {
		return nil, fmt.Errorf("sentinel home must differ from ENGRAM_DATA_DIR")
	}
	return []string{
		"ENGRAM_DATA_DIR=" + cfg.DataDir,
		"HOME=" + home,
		"USERPROFILE=" + home,
	}, nil
}

func MkdirSentinel(dir string) error {
	return os.MkdirAll(dir, 0o700)
}
