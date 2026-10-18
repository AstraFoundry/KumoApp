package desktop

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/egoist/mygo"
)

// DriverEnv names a directory through which end-to-end tests drive the
// app: they write a script to cmd.js, the app evaluates it in the main
// window's page and writes the result to out.json. Unset, nothing runs.
const DriverEnv = "KUMO_E2E_DRIVER"

func (s *Shell) startDriver() {
	dir := os.Getenv(DriverEnv)
	if dir == "" {
		return
	}
	_ = os.MkdirAll(dir, 0o755)
	cmd, out := filepath.Join(dir, "cmd.js"), filepath.Join(dir, "out.json")
	go func() {
		for range time.Tick(100 * time.Millisecond) {
			script, err := os.ReadFile(cmd)
			if err != nil {
				continue
			}
			_ = os.Remove(cmd)
			s.mu.Lock()
			win := s.window
			s.mu.Unlock()
			result := map[string]any{}
			if win == nil {
				result["error"] = "no window"
			} else if value, err := mygo.EvalAs[any](win.Page(), string(script)); err != nil {
				result["error"] = err.Error()
			} else {
				result["value"] = value
			}
			data, _ := json.Marshal(result)
			_ = os.WriteFile(out+".tmp", data, 0o644)
			_ = os.Rename(out+".tmp", out)
		}
	}()
}
