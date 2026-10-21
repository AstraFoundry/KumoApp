//go:build !windows

package application

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

// TUNPermission says whether the core can create a TUN interface.
type TUNPermission struct {
	Granted bool `json:"granted"`
	// Method is how it was granted: "root" when Kumo itself runs as root,
	// "setuid" when the core is owned by root with the setuid bit, or
	// "capability" for Linux file capabilities.
	Method string `json:"method,omitempty"`
}

func tunPermission(corePath string) TUNPermission {
	if os.Geteuid() == 0 {
		return TUNPermission{Granted: true, Method: "root"}
	}
	info, err := os.Stat(corePath)
	if err != nil {
		return TUNPermission{}
	}
	if st, ok := info.Sys().(*syscall.Stat_t); ok && st.Uid == 0 && info.Mode()&os.ModeSetuid != 0 {
		return TUNPermission{Granted: true, Method: "setuid"}
	}
	if runtime.GOOS == "linux" {
		if out, err := exec.Command("getcap", corePath).Output(); err == nil && strings.Contains(string(out), "cap_net_admin") {
			return TUNPermission{Granted: true, Method: "capability"}
		}
	}
	return TUNPermission{}
}

// grantTUN makes the core run as root through the setuid bit on macOS, or
// gives it CAP_NET_ADMIN on Linux, after the system asks for an
// administrator's password.
func grantTUN(ctx context.Context, corePath string) error {
	quoted := "'" + strings.ReplaceAll(corePath, "'", `'\''`) + "'"
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`do shell script "chown root:admin %s && chmod u+s %s" with administrator privileges with prompt "Kumo needs to let Mihomo create a network interface for TUN."`,
			strings.ReplaceAll(quoted, `"`, `\"`), strings.ReplaceAll(quoted, `"`, `\"`))
		out, err := exec.CommandContext(ctx, "/usr/bin/osascript", "-e", script).CombinedOutput()
		if err != nil {
			return fmt.Errorf("permission was not granted: %s", strings.TrimSpace(string(out)))
		}
		return nil
	default:
		out, err := exec.CommandContext(ctx, "pkexec", "setcap", "cap_net_admin,cap_net_bind_service=+eip", corePath).CombinedOutput()
		if err != nil {
			return fmt.Errorf("permission was not granted: %s", strings.TrimSpace(string(out)))
		}
		return nil
	}
}
