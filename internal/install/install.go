package install

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonathon-chew/Packr/internal/archive"
	"github.com/jonathon-chew/Packr/internal/config"
	"github.com/jonathon-chew/Packr/internal/download"
	"github.com/jonathon-chew/Packr/internal/registry"
)

func Install(pkg registry.Package, release registry.Release, target registry.Target, paths config.Paths) error {
	// 1. download into cache
	url := target.URL
	if url == "" {
		return fmt.Errorf("empty URL for target")
	}

	cachePath, err := download.File(paths.CacheDir, pkg.Name, release.Version, url)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	if err := download.VerifySHA256(cachePath, target.SHA256); err != nil {
		return fmt.Errorf("checksum: %w", err)
	}

	archiveType := target.ArchiveType
	if archiveType == "" {
		archiveType = archive.GuessArchiveType(cachePath)
	}

	destName := filepath.Base(target.Bin)
	if destName == "." || destName == "" || destName == string(filepath.Separator) {
		destName = filepath.Base(cachePath)
	}

	// 2. extract to temp dir
	tmpDir, err := os.MkdirTemp("", "packr-extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	var srcBinPath string
	switch archiveType {
	case "tar.gz":
		if err := archive.ExtractTarGz(cachePath, tmpDir); err != nil {
			return fmt.Errorf("extract: %w", err)
		}
		srcBinPath, err = findBinary(tmpDir, target.Bin)
	case "zip":
		if err := archive.ExtractZip(cachePath, tmpDir); err != nil {
			return fmt.Errorf("extract: %w", err)
		}
		srcBinPath, err = findBinary(tmpDir, target.Bin)
	case "binary":
		srcBinPath = cachePath
	default:
		return fmt.Errorf("unsupported archive type for %s", cachePath)
	}
	if err != nil {
		return err
	}

	// 4. copy into ~/.config/packr/bin/<bin>
	destBinPath := filepath.Join(paths.BinDir, destName)
	if err := copyFile(srcBinPath, destBinPath, 0o755); err != nil {
		return err
	}

	// 5. update state (v0: very simple)
	if err := AppendState(paths, pkg.Name, release.Version, destBinPath); err != nil {
		return fmt.Errorf("updating state: %w", err)
	}

	return nil
}

func findBinary(root, binName string) (string, error) {
	if binName != "" {
		var found string
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type().IsRegular() && filepath.Base(path) == filepath.Base(binName) {
				found = path
				return errors.New("found")
			}
			return nil
		})
		if found != "" {
			return found, nil
		}
		if err != nil && err.Error() != "found" {
			return "", err
		}
	}

	var candidates []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if isExecutableCandidate(path, d) {
			candidates = append(candidates, path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) > 1 {
		best := rankCandidates(candidates, binName)
		if best != "" {
			return best, nil
		}
	}
	if binName != "" {
		return "", fmt.Errorf("binary %q not found in archive", filepath.Base(binName))
	}
	return "", fmt.Errorf("could not determine which binary to install from archive")
}

func copyFile(src, dest string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func isExecutableCandidate(path string, d os.DirEntry) bool {
	info, err := d.Info()
	if err != nil {
		return false
	}
	if info.Mode()&0o111 == 0 {
		return false
	}

	base := strings.ToLower(filepath.Base(path))
	for _, blocked := range []string{"readme", "license", "install", "uninstall", ".txt", ".md"} {
		if strings.Contains(base, blocked) {
			return false
		}
	}
	return true
}

func rankCandidates(candidates []string, preferred string) string {
	preferred = strings.ToLower(filepath.Base(preferred))
	if preferred != "" {
		for _, candidate := range candidates {
			if strings.EqualFold(filepath.Base(candidate), preferred) {
				return candidate
			}
		}
	}

	var best string
	bestScore := -1
	for _, candidate := range candidates {
		score := 0
		base := strings.ToLower(filepath.Base(candidate))
		if !strings.Contains(base, ".") {
			score += 2
		}
		if strings.Contains(strings.ToLower(filepath.ToSlash(candidate)), "/bin/") {
			score++
		}
		if score > bestScore {
			best = candidate
			bestScore = score
		}
	}
	return best
}
