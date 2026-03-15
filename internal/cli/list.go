package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonathon-chew/Packr/internal/config"
)

/*
{
  "name": "ripgrep",
  "version": "14.0.0",
  "bin": "rg"
}
*/

type State struct {
	Packages []PackageJSON `json:"packages"`
}

type PackageJSON struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Path    string `json:"path"`
}

func loadState(statePath string) (State, error) {
	var s State

	b, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return State{Packages: []PackageJSON{}}, nil
		}
		return State{}, err
	}

	if len(b) == 0 {
		return State{Packages: []PackageJSON{}}, nil
	}

	if err := json.Unmarshal(b, &s); err != nil {
		return State{}, err
	}

	if s.Packages == nil {
		s.Packages = []PackageJSON{}
	}

	for _, pkg := range s.Packages {
		fmt.Printf("%s %s (%s)\n", pkg.Name, pkg.Version, pkg.Path)
	}

	return s, nil
}

func runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	if err := fs.Parse(args); err != nil {
		return 2
	}

	// fmt.Println("TODO: list installed packages")

	DefaultDir, err := config.DefaultPaths()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error getting default paths:", err)
		return 2
	}

	statePath := filepath.Join(DefaultDir.StateDir, "state.json")
	state, err := loadState(statePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error loading state:", err)
		return 2
	}

	if len(state.Packages) == 0 {
		fmt.Println("No packages installed.")
		return 0
	}

	return 0
}
