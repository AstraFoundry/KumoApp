// Command kumo is Kumo's command-line interface, built on its own for
// development and CI. The app's binary runs the same CLI when invoked as
// kumo, which is how the installed command works.
package main

import (
	"os"

	"github.com/AstraFoundry/KumoApp/internal/interfaces/cli"
)

func main() { os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr)) }
