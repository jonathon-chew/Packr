package install

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/jonathon-chew/Packr/internal/config"
)

type InstalledPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

type State struct {
	Packages []InstalledPackage `json:"packages"`
}

func statePath(p config.Paths) string {
	return filepath.Join(p.StateDir, "state.json")
}

func loadState(p config.Paths) (State, error) {
	path := statePath(p)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return State{}, nil
		}
		return State{}, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return State{}, err
	}
	return s, nil
}

func saveState(p config.Paths, s State) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(statePath(p), data, 0o644)
}

func AppendState(p config.Paths, name, version, path string) error {
	s, err := loadState(p)
	if err != nil {
		return err
	}

	// naive: no dedupe for v0.1
	s.Packages = append(s.Packages, InstalledPackage{
		Name:    name,
		Version: version,
		Path:    path,
	})

	return saveState(p, s)
}
