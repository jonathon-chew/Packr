package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/jonathon-chew/Packr/internal/config"
	"github.com/jonathon-chew/Packr/internal/install"
	"github.com/jonathon-chew/Packr/internal/platform"
	"github.com/jonathon-chew/Packr/internal/registry"
)

func runInstall(args []string) int {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var version string
	fs.StringVar(&version, "version", "", "install a specific version (default: latest)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "install requires exactly 1 package name")
		return 2
	}
	name := rest[0]

	paths, err := config.DefaultPaths()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error determining paths: %v\n", err)
		return 1
	}

	if err := config.EnsureDirs(paths); err != nil {
		fmt.Fprintf(os.Stderr, "error creating packr directories: %v\n", err)
		return 1
	}

	pkg, err := registry.LoadLocalPackage(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading package %q: %v\n", name, err)
		return 1
	}

	release, err := registry.ResolveRelease(pkg, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving version for %q: %v\n", name, err)
		return 1
	}

	pk := platform.Key()
	target, ok := release.Targets[pk]
	if !ok {
		fmt.Fprintf(os.Stderr, "no binary for %s on %s\n", name, pk)
		return 1
	}

	if err := install.Install(pkg, release, target, paths); err != nil {
		fmt.Fprintf(os.Stderr, "install failed: %v\n", err)
		return 1
	}

	fmt.Printf("Installed %s %s\n", pkg.Name, release.Version)
	fmt.Printf("Make sure %s is on your PATH\n", paths.BinDir)
	return 0
}
