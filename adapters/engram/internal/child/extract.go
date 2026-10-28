package child

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractArchive verifies sha256 then extracts the engram binary to destPath.
func ExtractArchive(archivePath, expectedSHA256, destPath string) error {
	sum, err := fileSHA256(archivePath)
	if err != nil {
		return err
	}
	if !equalHex(sum, expectedSHA256) {
		return fmt.Errorf("archive checksum mismatch")
	}
	if strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
		return extractZip(archivePath, destPath)
	}
	return extractTarGz(archivePath, destPath)
}

func extractTarGz(archivePath, destPath string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		base := filepath.Base(h.Name)
		if base != "engram" && !strings.HasPrefix(base, "engram") {
			continue
		}
		return writeExe(tr, destPath)
	}
	return fmt.Errorf("engram binary not found in archive")
}

func extractZip(archivePath, destPath string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		if filepath.Base(f.Name) != "engram.exe" && filepath.Base(f.Name) != "engram" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		return writeExe(rc, destPath)
	}
	return fmt.Errorf("engram binary not found in zip")
}

func writeExe(r io.Reader, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, r); err != nil {
		return err
	}
	return nil
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func equalHex(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

// WriteBinarySHA256ToPin updates pin file with extracted binary hash (release/CI helper).
func BinarySHA256(path string) (string, error) {
	return fileSHA256(path)
}
