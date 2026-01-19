package main

import (
	"os"

	"github.com/jonathon-chew/Packr/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:]))
}
