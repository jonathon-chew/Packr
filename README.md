# packr

Packr is a fast, Go-native installer for CLI tools from prebuilt GitHub releases.

It installs **prebuilt binaries only**, into **user space**, with
deterministic and atomic installs.

## Goals
- single static binary
- no system dependencies
- no source builds
- cross-platform (macOS / Linux)
- boring, predictable behavior

## Install (dev)
```bash
go build ./cmd/packr
./packr version
```

## Usage
```bash
packr install ripgrep
packr install sharkdp/bat
packr list
```

## Package Sources
Packr can install tools in two ways:

1. Bundled or user-provided YAML package definitions in `~/.config/packr/registry`.
2. Direct GitHub repositories using `owner/repo`, where Packr resolves the latest
   release asset for the current platform automatically.

The GitHub mode is heuristic-based, so curated YAML remains useful for projects
with unusual asset naming or multiple binaries per archive.
