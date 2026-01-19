package install

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

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

	// 2. extract to temp dir
	tmpDir, err := os.MkdirTemp("", "packr-extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	switch archive.GuessArchiveType(cachePath) {
	case "tar.gz":
		if err := archive.ExtractTarGz(cachePath, tmpDir); err != nil {
			return fmt.Errorf("extract: %w", err)
		}
	default:
		return fmt.Errorf("unsupported archive type for %s", cachePath)
	}

	// 3. locate the binary file
	srcBinPath, err := findBinary(tmpDir, target.Bin)
	if err != nil {
		return err
	}

	// 4. copy into ~/.packr/bin/<bin>
	destBinPath := filepath.Join(paths.BinDir, target.Bin)
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
	var found string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type().IsRegular() && filepath.Base(path) == binName {
			found = path
			return fmt.Errorf("found") // early stop
		}
		return nil
	})
	if found != "" {
		return found, nil
	}
	if err != nil && err.Error() == "found" {
		// handled above
	}
	return "", fmt.Errorf("binary %q not found in archive", binName)
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
