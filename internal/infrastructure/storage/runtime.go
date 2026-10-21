package storage

import (
	"encoding/json"

	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/paths"
)

// RuntimeStore keeps the runtime in state.json.
type RuntimeStore struct{ paths paths.Paths }

// NewRuntimeStore stores the runtime under p.
func NewRuntimeStore(p paths.Paths) *RuntimeStore { return &RuntimeStore{paths: p} }

// Load reads the runtime without locking; a missing file is a fresh install.
func (s *RuntimeStore) Load() (core.Runtime, error) {
	return s.read()
}

// Update changes the runtime under the shared lock.
func (s *RuntimeStore) Update(fn func(*core.Runtime) error) (core.Runtime, error) {
	lock, err := Lock(s.paths.StateLock())
	if err != nil {
		return core.Runtime{}, err
	}
	defer lock.Unlock()
	runtime, err := s.read()
	if err != nil {
		return core.Runtime{}, err
	}
	if err := fn(&runtime); err != nil {
		return runtime, err
	}
	return runtime, WriteJSON(s.paths.State(), runtime)
}

// legacyState is the shape the Swift app wrote, read once to migrate.
type legacyState struct {
	ProxyPorts *struct {
		MixedPort int `json:"mixedPort"`
	} `json:"proxyPorts"`
	SystemProxyEnabled  *bool                 `json:"systemProxyEnabled"`
	SystemProxySettings *systemproxy.Settings `json:"systemProxySettings"`
	SystemProxy         *json.RawMessage      `json:"systemProxy"`
	RuntimeSettings     *json.RawMessage      `json:"runtimeSettings"`
}

func (s *RuntimeStore) read() (core.Runtime, error) {
	runtime := core.NewRuntime()
	var raw json.RawMessage
	found, err := ReadJSON(s.paths.State(), &raw)
	if err != nil || !found {
		return runtime, err
	}
	if err := json.Unmarshal(raw, &runtime); err != nil {
		return runtime, err
	}
	var legacy legacyState
	if json.Unmarshal(raw, &legacy) == nil && legacy.SystemProxy == nil {
		if legacy.ProxyPorts != nil && legacy.ProxyPorts.MixedPort > 0 && legacy.RuntimeSettings == nil {
			runtime.Settings.MixedPort = legacy.ProxyPorts.MixedPort
		}
		if legacy.SystemProxySettings != nil {
			runtime.SystemProxy.Settings = *legacy.SystemProxySettings
		}
		if legacy.SystemProxyEnabled != nil {
			runtime.SystemProxy.Enabled = *legacy.SystemProxyEnabled
		}
	}
	if runtime.Mode == "" {
		runtime.Mode = core.ModeRule
	}
	if runtime.State == "" {
		runtime.State = core.StateStopped
	}
	if runtime.SystemProxy.Settings.Mode == "" {
		runtime.SystemProxy.Settings.Mode = systemproxy.ModeManual
	}
	return runtime, nil
}
