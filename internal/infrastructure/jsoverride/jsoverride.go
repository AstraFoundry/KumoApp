// Package jsoverride runs JavaScript overrides: scripts that define
// main(config) and return the changed configuration, as Clash Verge's
// "scripts" do.
package jsoverride

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"gopkg.in/yaml.v3"
)

// Timeout stops a script that runs too long.
var Timeout = 5 * time.Second

// Runner implements runtimeconfig.Transformer with goja.
type Runner struct{}

// Transform runs script's main on config and returns its result.
func (Runner) Transform(name, script string, config *yaml.Node) (*yaml.Node, error) {
	var value any
	if err := config.Decode(&value); err != nil {
		return nil, err
	}
	vm := goja.New()
	vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
	timer := time.AfterFunc(Timeout, func() { vm.Interrupt("timed out") })
	defer timer.Stop()
	_ = vm.Set("console", map[string]any{"log": func(...any) {}})
	if _, err := vm.RunScript(name, script); err != nil {
		return nil, fmt.Errorf("script error: %w", err)
	}
	main, ok := goja.AssertFunction(vm.Get("main"))
	if !ok {
		return nil, errors.New("the script defines no function main(config)")
	}
	// The configuration crosses as JSON so the script gets real JavaScript
	// objects and arrays, whose changes (push, unshift, delete) all count.
	input, err := json.Marshal(normalize(value))
	if err != nil {
		return nil, err
	}
	parse, _ := goja.AssertFunction(vm.Get("JSON").ToObject(vm).Get("parse"))
	stringify, _ := goja.AssertFunction(vm.Get("JSON").ToObject(vm).Get("stringify"))
	arg, err := parse(goja.Undefined(), vm.ToValue(string(input)))
	if err != nil {
		return nil, err
	}
	result, err := main(goja.Undefined(), arg)
	if err != nil {
		return nil, fmt.Errorf("main failed: %w", err)
	}
	if result == nil || goja.IsUndefined(result) || goja.IsNull(result) {
		return nil, errors.New("main must return the configuration object")
	}
	text, err := stringify(goja.Undefined(), result)
	if err != nil {
		return nil, err
	}
	var exported map[string]any
	if err := json.Unmarshal([]byte(text.String()), &exported); err != nil {
		return nil, errors.New("main must return the configuration object")
	}
	var node yaml.Node
	if err := node.Encode(exported); err != nil {
		return nil, err
	}
	return &node, nil
}

// normalize turns the map[string]any YAML decodes into plain values goja
// exposes as JavaScript objects and arrays.
func normalize(v any) any {
	switch v := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, item := range v {
			out[k] = normalize(item)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = normalize(item)
		}
		return out
	}
	return v
}
