package core

import (
	"fmt"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
)

// Runtime is the aggregate of everything Kumo controls about the core: the
// current session, the settings the next launch uses, and the system proxy.
// It is persisted as one document so the app, the CLI and the service see
// the same state.
type Runtime struct {
	State     RunState        `json:"state"`
	PID       int             `json:"pid,omitempty"`
	CorePath  string          `json:"corePath,omitempty"`
	Owner     Owner           `json:"owner,omitempty"`
	Readiness Readiness       `json:"readiness,omitempty"`
	StartedAt *time.Time      `json:"startedAt,omitempty"`
	Message   string          `json:"message,omitempty"`
	Mode      Mode            `json:"mode"`
	Endpoint  Endpoint        `json:"endpoint"`
	Settings  RuntimeSettings `json:"runtimeSettings"`
	// SystemProxy is the system proxy as Kumo tracks it.
	SystemProxy systemproxy.State `json:"systemProxy"`
	// LastExit describes how the previous session ended, when it ended on
	// its own.
	LastExit *Exit `json:"lastExit,omitempty"`
}

// Exit is how a core process ended.
type Exit struct {
	At      time.Time `json:"at"`
	Code    int       `json:"code"`
	Message string    `json:"message"`
	// Expected is false when the core died without Kumo stopping it.
	Expected bool `json:"expected"`
}

// NewRuntime is the state of a fresh install.
func NewRuntime() Runtime {
	return Runtime{
		State:       StateStopped,
		Mode:        ModeRule,
		Endpoint:    DefaultEndpoint(),
		Settings:    DefaultRuntimeSettings(),
		SystemProxy: systemproxy.DefaultState(),
	}
}

// IsActive reports whether a core process is (or should be) alive.
func (r Runtime) IsActive() bool {
	return r.State == StateRunning || r.State == StateStarting || r.State == StateStopping
}

// BeginStart moves a stopped or failed runtime into starting. StartedAt
// records the attempt until the process launches.
func (r *Runtime) BeginStart(owner Owner, corePath string, at time.Time) error {
	if r.IsActive() {
		return fmt.Errorf("%w (pid %d)", ErrAlreadyRunning, r.PID)
	}
	r.StartedAt = &at
	r.State = StateStarting
	r.Owner = owner
	r.CorePath = corePath
	r.PID = 0
	r.Readiness = ""
	r.Message = "Starting Mihomo."
	return nil
}

// Launched records the core's process.
func (r *Runtime) Launched(pid int, at time.Time) {
	r.State = StateRunning
	r.PID = pid
	r.Readiness = ReadinessLaunched
	r.StartedAt = &at
	r.LastExit = nil
	r.Message = fmt.Sprintf("Mihomo started with pid %d.", pid)
}

// Ready records that the controller answers.
func (r *Runtime) Ready() {
	if r.State == StateRunning {
		r.Readiness = ReadinessControllerReady
		r.Message = "Mihomo is ready."
	}
}

// Fail ends a start or a session with an error.
func (r *Runtime) Fail(message string) {
	r.State = StateFailed
	r.PID = 0
	r.Readiness = ""
	r.StartedAt = nil
	r.Message = message
}

// Stopped ends the session.
func (r *Runtime) Stopped(message string) {
	r.State = StateStopped
	r.PID = 0
	r.Readiness = ""
	r.StartedAt = nil
	r.Owner = ""
	r.Message = message
}

// Exited records a core that ended without Kumo stopping it.
func (r *Runtime) Exited(exit Exit) {
	r.LastExit = &exit
	r.Fail(exit.Message)
}

// Store persists the runtime. Update runs fn under a lock shared by every
// process, so concurrent writers never lose each other's changes.
type Store interface {
	Load() (Runtime, error)
	Update(fn func(*Runtime) error) (Runtime, error)
}
