package core

import (
	"errors"
	"fmt"
)

// Errors callers branch on. Each surface (CLI exit codes, UI states) maps
// them to its own presentation.
var (
	ErrAlreadyRunning = errors.New("mihomo is already running")
	ErrNotRunning     = errors.New("mihomo is not running")
	ErrCoreNotFound   = errors.New("no mihomo executable found")
)

// PortConflict reports that a port Kumo needs is taken by another process.
type PortConflict struct {
	// Role is what Kumo wanted the port for: "mixed" or "controller".
	Role string `json:"role"`
	Port int    `json:"port"`
	// PID and Process identify the holder when the system tells.
	PID     int    `json:"pid,omitempty"`
	Process string `json:"process,omitempty"`
	// Suggested is a free port Kumo can use instead.
	Suggested int `json:"suggested,omitempty"`
}

func (c *PortConflict) Error() string {
	holder := "another process"
	if c.Process != "" {
		holder = fmt.Sprintf("%s (pid %d)", c.Process, c.PID)
	} else if c.PID > 0 {
		holder = fmt.Sprintf("pid %d", c.PID)
	}
	return fmt.Sprintf("the %s port %d is in use by %s", c.Role, c.Port, holder)
}

// PortConflicts is every conflict found before a launch.
type PortConflicts []*PortConflict

func (c PortConflicts) Error() string {
	if len(c) == 1 {
		return c[0].Error()
	}
	message := "ports in use:"
	for _, conflict := range c {
		message += " " + conflict.Error() + ";"
	}
	return message
}
