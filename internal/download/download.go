package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func File(cacheDir, pkgName, version, filename string) (string, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return "", err
	}

	dest := filepath.Join(cacheDir, fmt.Sprintf("%s-%s-%s", pkgName, version, filename))
	if _, err := os.Stat(dest); err == nil {
		return dest, nil // already cached
	}

	out, err := os.Create(dest)
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(filename) // we'll pass full URL as filename; rename arg if you prefer
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %s", resp.Status)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", err
	}

	return dest, nil
}

func VerifySHA256(path, expected string) error {
	if expected == "" {
		return nil // optional for now
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expected, actual)
	}
	return nil
}
