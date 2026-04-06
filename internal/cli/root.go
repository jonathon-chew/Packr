package cli

import (
	"fmt"
	"os"
)

func Run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}

	switch args[0] {
	case "help", "-h", "--help":
		usage()
		return 0
	case "version":
		fmt.Println("packr v0.1.0-dev")
		return 0
	case "install":
		return runInstall(args[1:])
	case "list":
		return runList(args[1:])
	case "remove":
		// return runRemove(args[1:])
		return 0
	case "upgrade":
		// return runUpgrade(args[1:])
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Print(`packr — fast, user-local installer for CLI tools

Usage:
  packr <command> [options]

Commands:
  install   Install a tool
  remove    Remove a tool
  list      List installed tools
  upgrade   Upgrade tools
  version   Show version
  help      Show this help

Examples:
  packr install ripgrep
  packr install sharkdp/bat
  packr list
`)
}
