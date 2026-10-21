// Package core models the Mihomo runtime that Kumo controls: how traffic is
// routed, how the core is reached, the settings Kumo owns in the generated
// configuration, and the lifecycle of one core session.
package core

import "fmt"

// Mode is the outbound mode Mihomo routes traffic with.
type Mode string

const (
	ModeRule   Mode = "rule"
	ModeGlobal Mode = "global"
	ModeDirect Mode = "direct"
)

// Modes lists every outbound mode in display order.
var Modes = []Mode{ModeRule, ModeGlobal, ModeDirect}

// ParseMode validates a user-supplied mode.
func ParseMode(value string) (Mode, error) {
	for _, mode := range Modes {
		if string(mode) == value {
			return mode, nil
		}
	}
	return "", fmt.Errorf("unknown mode %q: use rule, global or direct", value)
}

// RunState is where a core session is in its lifecycle.
type RunState string

const (
	StateStopped  RunState = "stopped"
	StateStarting RunState = "starting"
	StateRunning  RunState = "running"
	StateStopping RunState = "stopping"
	StateFailed   RunState = "failed"
)

// Readiness is how far a running core has come since it was launched.
type Readiness string

const (
	ReadinessLaunched        Readiness = "processLaunched"
	ReadinessControllerReady Readiness = "controllerReady"
)

// Owner is the surface that started the current core session. A session the
// app started ends with the app; one the CLI started outlives it.
type Owner string

const (
	OwnerApp     Owner = "app"
	OwnerCLI     Owner = "cli"
	OwnerService Owner = "service"
)
