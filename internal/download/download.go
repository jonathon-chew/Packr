package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func File(cacheDir, pkgName, version, fileURL string) (string, error) {
	// Ensure cacheDir is an absolute path
	absCacheDir, err := filepath.Abs(cacheDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve cache directory: %w", err)
	}

	if err := os.MkdirAll(absCacheDir, 0o755); err != nil {
		return "", err
	}

	// Extract filename from URL
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Get the filename from the URL path
	urlPath := parsedURL.Path
	urlFilename := filepath.Base(urlPath)
	if urlFilename == "" || urlFilename == "/" {
		return "", fmt.Errorf("could not determine filename from URL: %s", fileURL)
	}

	dest := filepath.Join(absCacheDir, fmt.Sprintf("%s-%s-%s", pkgName, version, urlFilename))

	// Check if file already exists in cache
	if _, err := os.Stat(dest); err == nil {
		return dest, nil // already cached
	}

	// Download the file
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", fileURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed: %s (status: %d)", fileURL, resp.StatusCode)
	}

	// Create the destination file
	out, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("failed to create cache file: %w", err)
	}
	defer out.Close()

	// Copy the response body to the file
	if _, err := io.Copy(out, resp.Body); err != nil {
		os.Remove(dest) // cleanup on error
		return "", fmt.Errorf("failed to write cache file: %w", err)
	}

	return dest, nil
}

func VerifySHA256(path, expected string) error {
	if expected == "" {
		return nil // optional for now
	}

	// Strip "sha256:" prefix if present
	expected = strings.TrimPrefix(expected, "sha256:")
	expected = strings.TrimSpace(expected)

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
