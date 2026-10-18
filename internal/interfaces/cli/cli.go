// Package cli is the kumo command: the same use cases as the app, for
// people at a terminal and for coding agents. Every command accepts
// --json and then prints {"ok", "data", "error"} on stdout.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"slices"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/AstraFoundry/KumoApp/internal/application"
	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/paths"
)

// Version is set by the build.
var Version = "dev"

// Env is what a command runs with.
type Env struct {
	Stdout, Stderr io.Writer
	JSON           bool
	Color          bool
	DryRun         bool
	Args           []string
	Flags          map[string]string
	Kumo           func() (*application.Kumo, error)
	Ctx            context.Context
}

// Arg returns the nth positional argument, or "".
func (e *Env) Arg(n int) string {
	if n < len(e.Args) {
		return e.Args[n]
	}
	return ""
}

// Flag returns a flag's value and whether it was given.
func (e *Env) Flag(name string) (string, bool) {
	v, ok := e.Flags[name]
	return v, ok
}

// Bool reports a boolean flag.
func (e *Env) Bool(name string) bool {
	v, ok := e.Flags[name]
	return ok && v != "false"
}

// Int reads a numeric flag with a default.
func (e *Env) Int(name string, fallback int) (int, error) {
	v, ok := e.Flags[name]
	if !ok {
		return fallback, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, usageError("--%s takes a number", name)
	}
	return n, nil
}

// Result is a command's data and how to show it as text.
type Result struct {
	Data any
	Text func(w *Writer)
}

// Command is one kumo command.
type Command struct {
	Name    string
	Aliases []string
	Summary string
	Usage   string
	// Flags lists the flags the command takes; those ending in "=" take a value.
	Flags []string
	Sub   []*Command
	Run   func(e *Env) (Result, error)
}

type usage struct{ message string }

func (u usage) Error() string { return u.message }

func usageError(format string, args ...any) error { return usage{fmt.Sprintf(format, args...)} }

// globalFlags apply to every command.
var globalFlags = []string{"json", "color=", "dry-run", "help", "h", "silent", "version", "v", "l", "long"}

// Run executes kumo with args and returns the exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	env := &Env{Stdout: stdout, Stderr: stderr, Ctx: ctx, Flags: map[string]string{}}
	var kumo *application.Kumo
	env.Kumo = func() (*application.Kumo, error) {
		if kumo != nil {
			return kumo, nil
		}
		p, err := paths.Default()
		if err != nil {
			return nil, err
		}
		deps := application.DefaultDependencies(p, core.OwnerCLI)
		deps.BundledDir = bundledDir()
		kumo = application.New(deps)
		return kumo, nil
	}

	cmds := commands()
	path, cmd, rest := route(cmds, args)
	flagSpec := append(slices.Clone(globalFlags), flagsOf(cmd)...)
	positional, err := parseFlags(rest, flagSpec, env.Flags)
	env.Args = positional
	env.JSON = env.Bool("json")
	env.DryRun = env.Bool("dry-run")
	env.Color = colorEnabled(env, stdout)
	if err != nil {
		return finish(env, Result{}, err)
	}
	switch {
	case env.Bool("version") || env.Bool("v") && cmd == nil:
		if env.JSON {
			return finish(env, Result{Data: map[string]string{"version": Version}}, nil)
		}
		fmt.Fprintln(stdout, Version)
		return 0
	case env.Bool("l") || env.Bool("long"):
		fmt.Fprint(stdout, longHelp(cmds))
		return 0
	case cmd == nil:
		if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
			return finish(env, Result{}, usageError("unknown command %q; run kumo --help", args[0]))
		}
		fmt.Fprint(stdout, topHelp(cmds))
		return 0
	case env.Bool("help") || env.Bool("h") || cmd.Run == nil:
		fmt.Fprint(stdout, commandHelp(path, cmd))
		return 0
	}
	result, err := cmd.Run(env)
	return finish(env, result, err)
}

func flagsOf(cmd *Command) []string {
	if cmd == nil {
		return nil
	}
	return cmd.Flags
}

// route finds the deepest command named by the leading arguments.
func route(cmds []*Command, args []string) ([]string, *Command, []string) {
	var path []string
	var found *Command
	list := cmds
	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			// Global flags may come before the command.
			i++
			if arg == "--color" {
				i++
			}
			continue
		}
		next := lookup(list, arg)
		if next == nil {
			break
		}
		found, list = next, next.Sub
		path = append(path, next.Name)
		args = append(args[:i:i], args[i+1:]...)
	}
	return path, found, args
}

func lookup(cmds []*Command, name string) *Command {
	for _, c := range cmds {
		if c.Name == name || slices.Contains(c.Aliases, name) {
			return c
		}
	}
	return nil
}

// parseFlags splits args into flags (--name, --name value, --name=value,
// -h) and positional arguments.
func parseFlags(args, spec []string, out map[string]string) ([]string, error) {
	takesValue := map[string]bool{}
	known := map[string]bool{}
	for _, f := range spec {
		name := strings.TrimSuffix(f, "=")
		known[name] = true
		takesValue[name] = strings.HasSuffix(f, "=")
	}
	var positional []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" || isNumber(arg) {
			positional = append(positional, arg)
			continue
		}
		name := strings.TrimLeft(arg, "-")
		value, hasValue := "", false
		if eq := strings.IndexByte(name, '='); eq >= 0 {
			name, value, hasValue = name[:eq], name[eq+1:], true
		}
		if !known[name] {
			return positional, usageError("unknown flag --%s", name)
		}
		if takesValue[name] && !hasValue {
			if i+1 >= len(args) {
				return positional, usageError("--%s needs a value", name)
			}
			i++
			value = args[i]
		} else if !hasValue {
			value = "true"
		}
		out[name] = value
	}
	return positional, nil
}

func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func colorEnabled(env *Env, out io.Writer) bool {
	switch env.Flags["color"] {
	case "always":
		return true
	case "never":
		return false
	}
	if env.JSON || os.Getenv("NO_COLOR") != "" || os.Getenv("CLICOLOR") == "0" {
		return false
	}
	f, ok := out.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

type envelope struct {
	OK    bool    `json:"ok"`
	Data  any     `json:"data"`
	Error *string `json:"error"`
}

func finish(env *Env, result Result, err error) int {
	code := 0
	if err != nil {
		code = 1
		var u usage
		if errors.As(err, &u) {
			code = 2
		}
	}
	if env.JSON {
		out := envelope{OK: err == nil, Data: result.Data}
		if err != nil {
			message := err.Error()
			out.Error = &message
			out.Data = errorData(err)
		}
		enc := json.NewEncoder(env.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		_ = enc.Encode(out)
		return code
	}
	w := &Writer{out: env.Stdout, color: env.Color}
	if err == nil {
		switch {
		case result.Text != nil:
			result.Text(w)
		case result.Data != nil:
			w.Value(result.Data)
		}
		return 0
	}
	ew := &Writer{out: env.Stderr, color: env.Color}
	ew.Status("error", err.Error())
	var conflicts core.PortConflicts
	if errors.As(err, &conflicts) {
		for _, c := range conflicts {
			if c.Suggested > 0 {
				ew.Line(fmt.Sprintf("  run `kumo start --resolve-ports` to use port %d for %s instead", c.Suggested, c.Role))
			}
		}
	}
	if code == 2 {
		ew.Line("  run `kumo --help` for usage")
	}
	return code
}

// errorData carries structured details of an error, such as port
// conflicts, for JSON callers.
func errorData(err error) any {
	var conflicts core.PortConflicts
	if errors.As(err, &conflicts) {
		return map[string]any{"portConflicts": conflicts}
	}
	return nil
}

// sortedKeys returns a map's keys in order.
func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func bundledDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	// Inside Kumo.app the core ships in Contents/Resources.
	return resourcesDir(exe)
}
