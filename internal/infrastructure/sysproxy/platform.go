package sysproxy

import (
	"os"
	"runtime"
	"strings"

	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
)

// ForOS returns the platform of the running operating system.
// KUMO_SYSTEM_PROXY=file:<path> swaps in the file-backed pretend system
// for lifecycle simulations; it never changes the real settings.
func ForOS() systemproxy.Platform {
	if path, ok := strings.CutPrefix(os.Getenv("KUMO_SYSTEM_PROXY"), "file:"); ok && path != "" {
		return File{Path: path}
	}
	switch runtime.GOOS {
	case "darwin":
		return Darwin{Runner: ExecRunner{}}
	case "windows":
		return Windows{Runner: ExecRunner{}}
	default:
		return Linux{Runner: ExecRunner{}}
	}
}
