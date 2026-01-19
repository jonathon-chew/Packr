# packr

Packr is a fast, Go-native installer for CLI tools.

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

## Useage
```bash
packr install ripgrep
packr list
```
