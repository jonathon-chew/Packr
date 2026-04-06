package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Paths struct {
	Root        string
	BinDir      string
	CacheDir    string
	StateDir    string
	RegistryDir string
}

func DefaultPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	root := filepath.Join(home, ".config", "packr")
	// Ensure root is absolute (should already be, but be explicit)
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return Paths{}, fmt.Errorf("failed to resolve root directory: %w", err)
	}

	return Paths{
		Root:        absRoot,
		BinDir:      filepath.Join(absRoot, "bin"),
		CacheDir:    filepath.Join(absRoot, "cache"),
		StateDir:    filepath.Join(absRoot, "state"),
		RegistryDir: filepath.Join(absRoot, "registry"),
	}, nil
}

func EnsureDirs(p Paths) error {
	for _, dir := range []string{p.Root, p.BinDir, p.CacheDir, p.StateDir, p.RegistryDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
