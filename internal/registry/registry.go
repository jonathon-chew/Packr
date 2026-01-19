package registry

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadLocalPackage loads registry/<name>.yaml relative to the executable.
func LoadLocalPackage(name string) (Package, error) {
	exe, err := os.Executable()
	if err != nil {
		return Package{}, err
	}

	exeDir := filepath.Dir(exe)
	registryDir := filepath.Join(exeDir, "internal", "registry") // adjust if layout differs

	path := filepath.Join(registryDir, name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Package{}, fmt.Errorf("reading %s: %w", path, err)
	}

	var pkg Package
	if err := yaml.Unmarshal(data, &pkg); err != nil {
		return Package{}, fmt.Errorf("parsing %s: %w", path, err)
	}

	return pkg, nil
}

// ResolveRelease picks a release by version; if version == "", returns the first.
func ResolveRelease(pkg Package, version string) (Release, error) {
	if len(pkg.Releases) == 0 {
		return Release{}, fmt.Errorf("no releases defined")
	}

	if version == "" {
		return pkg.Releases[0], nil // later you can sort or do semver
	}

	for _, r := range pkg.Releases {
		if r.Version == version {
			return r, nil
		}
	}

	return Release{}, fmt.Errorf("version %q not found", version)
}
