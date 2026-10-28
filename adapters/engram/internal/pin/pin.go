package pin

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Archive struct {
	URL    string `json:"url"`
	SHA256 string `json:"sha256"`
}

type File struct {
	Version       string              `json:"version"`
	Upstream      string              `json:"upstream"`
	BinarySHA256  string              `json:"binary_sha256"`
	Archives      map[string]Archive  `json:"archives"`
}

func Load(path string) (*File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f File
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, err
	}
	return &f, nil
}

func VerifyBinary(path string, expected string) error {
	if expected == "" {
		return fmt.Errorf("pin missing binary_sha256")
	}
	sum, err := fileSHA256(path)
	if err != nil {
		return err
	}
	if !strings.EqualFold(sum, expected) {
		return fmt.Errorf("binary sha256 mismatch")
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
