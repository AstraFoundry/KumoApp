// Kumo is a calm desktop client for the Mihomo proxy core. One executable
// is the app, the kumo command (when invoked as kumo, or with arguments)
// and the guardian that cleans up after a crashed app.
package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"

	"github.com/AstraFoundry/KumoApp/internal/application"
	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/paths"
	"github.com/AstraFoundry/KumoApp/internal/interfaces/cli"
	"github.com/AstraFoundry/KumoApp/internal/interfaces/desktop"
)

func main() {
	if parent := os.Getenv(desktop.GuardianEnv); parent != "" {
		guard(parent)
		return
	}
	if isCLI() {
		cli.Version = version()
		os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
	}
	p, err := paths.Default()
	if err != nil {
		log.Fatal(err)
	}
	deps := application.DefaultDependencies(p, core.OwnerApp)
	if resources, err := mygo.App.Path(mygo.PathResources); err == nil {
		deps.BundledDir = resources
	}
	if err := desktop.Run(application.New(deps)); err != nil {
		log.Fatal(err)
	}
}

// isCLI reports whether this run is the kumo command: given a command, or
// on macOS invoked through the lowercase kumo link to Kumo.app's Kumo. The
// app itself starts without arguments (macOS adds a -psn_ one at most).
func isCLI() bool {
	if runtime.GOOS == "darwin" && filepath.Base(os.Args[0]) == "kumo" {
		return true
	}
	return len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-psn_") && os.Getenv("MYGO_GENERATE") == "" &&
		!strings.Contains(os.Args[1], "://")
}

func version() string {
	if v := mygo.App.Version(); v != "" {
		return v
	}
	return "dev"
}

// guard runs as a separate process for the app's lifetime.
func guard(parentPID string) {
	parent, err := strconv.Atoi(parentPID)
	if err != nil {
		return
	}
	p, err := paths.Default()
	if err != nil {
		return
	}
	k := application.New(application.DefaultDependencies(p, core.OwnerApp))
	report := application.Guard(context.Background(), k, parent, 500*time.Millisecond)
	for _, d := range report.Diagnostics {
		log.Println("guardian:", d)
	}
}
