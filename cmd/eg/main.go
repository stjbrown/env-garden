// Command eg is env-garden: a per-shell environment profile switcher.
package main

import (
	"os"

	"github.com/stjbrown/env-garden/internal/cli"
)

// version is overridden at release time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	cli.SetVersion(version)
	os.Exit(cli.Execute())
}
