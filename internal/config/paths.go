package config

import (
	"os"
	"path/filepath"
)

type Paths struct {
	Root     string
	BinDir   string
	CacheDir string
	StateDir string
}

func DefaultPaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Paths{}, err
	}

	root := filepath.Join(home, ".packr")
	return Paths{
		Root:     root,
		BinDir:   filepath.Join(root, "bin"),
		CacheDir: filepath.Join(root, "cache"),
		StateDir: filepath.Join(root, "state"),
	}, nil
}

func EnsureDirs(p Paths) error {
	for _, dir := range []string{p.Root, p.BinDir, p.CacheDir, p.StateDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}
