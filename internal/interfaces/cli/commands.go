package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/application"
	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/inspect"
	"github.com/AstraFoundry/KumoApp/internal/domain/override"
	"github.com/AstraFoundry/KumoApp/internal/domain/profile"
	"github.com/AstraFoundry/KumoApp/internal/domain/proxy"
	"github.com/AstraFoundry/KumoApp/internal/domain/systemproxy"
	"github.com/AstraFoundry/KumoApp/internal/domain/topology"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/clilink"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/skills"
	"github.com/AstraFoundry/KumoApp/internal/infrastructure/storage"
)

func commands() []*Command {
	return []*Command{
		{Name: "status", Aliases: []string{"st"}, Summary: "Show whether Kumo runs, its mode, profile and system proxy", Usage: "kumo status [--json]", Run: runStatus},
		{Name: "doctor", Summary: "Check the core, configuration, ports and system proxy", Usage: "kumo doctor [--json]", Run: runDoctor},
		{Name: "start", Summary: "Start Mihomo with the current profile", Usage: "kumo start [--core <path>] [--resolve-ports] [--json]", Flags: []string{"core=", "resolve-ports"}, Run: runStart},
		{Name: "stop", Summary: "Release the system proxy and stop Mihomo", Usage: "kumo stop [--json]", Run: runStop},
		{Name: "restart", Summary: "Stop and start Mihomo", Usage: "kumo restart [--resolve-ports] [--json]", Flags: []string{"resolve-ports"}, Run: runRestart},
		{Name: "mode", Summary: "Show or set the outbound mode", Usage: "kumo mode [rule|global|direct] [--json]", Run: runMode},
		{Name: "proxies", Aliases: []string{"proxy", "groups"}, Summary: "List proxy groups and their nodes", Usage: "kumo proxies [--group <name>] [--json]", Flags: []string{"group="}, Run: runProxies},
		{Name: "select", Summary: "Use a node in a proxy group", Usage: "kumo select <group> <node> [--json]", Run: runSelect},
		{Name: "delay", Summary: "Measure the delay of a group's nodes, or of one node", Usage: "kumo delay <group> [--json]\nkumo delay --node <name> [--url <test url>] [--json]", Flags: []string{"node=", "url="}, Run: runDelay},
		profileCommand(),
		overrideCommand(),
		{Name: "connections", Aliases: []string{"conns"}, Summary: "List open connections, or close them", Usage: "kumo connections [--limit <n>] [--close <id>] [--close-all] [--json]", Flags: []string{"limit=", "close=", "close-all"}, Run: runConnections},
		{Name: "rules", Summary: "List the running rules, or turn one off for now", Usage: "kumo rules [--disable <index>] [--enable <index>] [--json]", Flags: []string{"disable=", "enable="}, Run: runRules},
		providersCommand(),
		logsCommand(),
		sysproxyCommand(),
		tunCommand(),
		managedSection("dns", "Show or toggle Kumo's DNS settings", func(s *core.RuntimeSettings) *bool { return &s.DNS.Enabled }, func(s core.RuntimeSettings) any { return s.DNS }),
		managedSection("sniffer", "Show or toggle Kumo's sniffer settings", func(s *core.RuntimeSettings) *bool { return &s.Sniffer.Enabled }, func(s core.RuntimeSettings) any { return s.Sniffer }),
		configCommand(),
		coreCommand(),
		{Name: "topology", Aliases: []string{"graph"}, Summary: "Show how traffic flows: mode, rules, groups and nodes", Usage: "kumo topology [--issues] [--json]", Flags: []string{"issues"}, Run: runTopology},
		skillsCommand(),
		{Name: "cli-link", Summary: "Put the kumo command on PATH", Usage: "kumo cli-link status|install|uninstall [--json]", Sub: []*Command{
			{Name: "status", Summary: "Show where kumo is linked", Run: func(e *Env) (Result, error) { return Result{Data: clilink.Check()}, nil }},
			{Name: "install", Summary: "Link kumo into PATH", Run: func(e *Env) (Result, error) { s, err := clilink.Install(); return Result{Data: s}, err }},
			{Name: "uninstall", Summary: "Remove the link", Run: func(e *Env) (Result, error) { s, err := clilink.Uninstall(); return Result{Data: s}, err }},
		}},
		{Name: "prompt", Summary: "Print a prompt that tells a coding agent how to do a task with kumo", Usage: "kumo prompt <page> [--intent <id>] [--json]\nkumo prompt --list", Flags: []string{"intent=", "list"}, Run: runPrompt},
		{Name: "completion", Summary: "Print a shell completion script", Usage: "kumo completion <zsh|bash|fish>", Run: runCompletion},
		{Name: "version", Summary: "Print the version", Usage: "kumo version", Run: func(e *Env) (Result, error) {
			return Result{Data: map[string]string{"version": Version}, Text: func(w *Writer) { w.Line(Version) }}, nil
		}},
		{Name: "help", Summary: "Show help for a command", Usage: "kumo help <command>", Run: runHelp},
	}
}

func kumo(e *Env) (*application.Kumo, error) { return e.Kumo() }

func statusText(s application.Status) func(*Writer) {
	return func(w *Writer) {
		switch s.State {
		case core.StateRunning:
			w.Status("ok", fmt.Sprintf("Mihomo is running (pid %d)", s.PID))
		case core.StateFailed:
			w.Status("error", "Mihomo is not running: "+s.Message)
		default:
			w.Status("info", "Mihomo is "+string(s.State))
		}
		w.Field("mode", s.Mode)
		w.Field("profile", s.Profile.Name)
		w.Field("mixed port", s.MixedPort)
		w.Field("controller", s.Controller)
		proxyState := "off"
		switch {
		case s.SystemProxy.Active && s.SystemProxy.Ownership == systemproxy.TakenOver:
			proxyState = "taken over by another app"
		case s.SystemProxy.Active:
			proxyState = "on (" + string(s.SystemProxy.Settings.Mode) + ")"
		case s.SystemProxy.Enabled:
			proxyState = "on when Mihomo runs"
		}
		w.Field("system proxy", proxyState)
		w.Field("tun", map[bool]string{true: "on", false: "off"}[s.TUN])
	}
}

func runStatus(e *Env) (Result, error) {
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	s, err := k.Lifecycle.Status()
	return Result{Data: s, Text: statusText(s)}, err
}

func runDoctor(e *Env) (Result, error) {
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	report, err := k.Doctor(e.Ctx)
	return Result{Data: report, Text: func(w *Writer) {
		statusText(report.Status)(w)
		w.Line("")
		for _, c := range report.Checks {
			kind := "ok"
			if !c.OK {
				kind = "warn"
			}
			w.Status(kind, c.Name+": "+c.Detail)
		}
		w.Line("")
		w.Field("data", report.DataDir)
	}}, err
}

func runStart(e *Env) (Result, error) {
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	corePath, _ := e.Flag("core")
	opts := application.StartOptions{CorePath: corePath, ResolvePortConflicts: e.Bool("resolve-ports")}
	if e.DryRun {
		s, err := k.Lifecycle.Status()
		return Result{Data: map[string]any{"wouldStart": !s.Running(), "status": s}, Text: func(w *Writer) {
			w.Status("dry-run", "would start Mihomo with profile "+s.Profile.Name)
		}}, err
	}
	s, err := k.Lifecycle.Start(e.Ctx, opts)
	if errors.Is(err, core.ErrAlreadyRunning) {
		return Result{Data: s, Text: statusText(s)}, err
	}
	return Result{Data: s, Text: statusText(s)}, err
}

func runStop(e *Env) (Result, error) {
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	s, err := k.Lifecycle.Stop(e.Ctx)
	return Result{Data: s, Text: statusText(s)}, err
}

func runRestart(e *Env) (Result, error) {
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	s, err := k.Lifecycle.Restart(e.Ctx, application.StartOptions{ResolvePortConflicts: e.Bool("resolve-ports")})
	return Result{Data: s, Text: statusText(s)}, err
}

func runMode(e *Env) (Result, error) {
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	if e.Arg(0) == "" {
		s, err := k.Lifecycle.Status()
		return Result{Data: map[string]core.Mode{"mode": s.Mode}, Text: func(w *Writer) { w.Line(string(s.Mode)) }}, err
	}
	mode, err := core.ParseMode(e.Arg(0))
	if err != nil {
		return Result{}, usageError("%v", err)
	}
	s, err := k.Settings.SetMode(e.Ctx, mode)
	return Result{Data: s, Text: func(w *Writer) { w.Status("ok", "mode is "+string(s.Mode)) }}, err
}

func delayText(delay int) string {
	switch proxy.Grade(delay) {
	case proxy.DelayTimeout:
		return "timeout"
	case proxy.DelayUnknown:
		return "-"
	}
	return strconv.Itoa(delay) + " ms"
}

func runProxies(e *Env) (Result, error) {
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	groups, err := k.Proxies.Groups(e.Ctx)
	if err != nil {
		return Result{}, err
	}
	if name, ok := e.Flag("group"); ok {
		filtered := []proxy.Group{}
		for _, g := range groups {
			if g.Name == name {
				filtered = append(filtered, g)
			}
		}
		if len(filtered) == 0 {
			return Result{}, fmt.Errorf("no group named %q", name)
		}
		groups = filtered
	}
	return Result{Data: groups, Text: func(w *Writer) {
		for i, g := range groups {
			if i > 0 {
				w.Line("")
			}
			title := g.Name + " " + w.Dim(g.Type)
			if !g.Live {
				title += w.Dim(" (from profile; start Kumo for live selection)")
			}
			w.Heading(title)
			rows := [][]string{}
			for _, n := range g.Nodes {
				mark := " "
				if n.Name == g.Now {
					mark = "*"
				}
				rows = append(rows, []string{mark, n.Name, n.Type, delayText(n.Delay)})
			}
			w.Table(nil, rows)
		}
	}}, nil
}

func runSelect(e *Env) (Result, error) {
	if e.Arg(0) == "" || e.Arg(1) == "" {
		return Result{}, usageError("usage: kumo select <group> <node>")
	}
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	err = k.Proxies.Select(e.Ctx, e.Arg(0), e.Arg(1))
	data := map[string]string{"group": e.Arg(0), "selected": e.Arg(1)}
	return Result{Data: data, Text: func(w *Writer) { w.Status("ok", e.Arg(0)+" uses "+e.Arg(1)) }}, err
}

func runDelay(e *Env) (Result, error) {
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	if node, ok := e.Flag("node"); ok {
		url, _ := e.Flag("url")
		delay, err := k.Proxies.TestNode(e.Ctx, node, url)
		return Result{Data: map[string]int{node: delay}, Text: func(w *Writer) { w.Line(node + "  " + delayText(delay)) }}, err
	}
	if e.Arg(0) == "" {
		return Result{}, usageError("usage: kumo delay <group> | kumo delay --node <name>")
	}
	results, err := k.Proxies.TestGroup(e.Ctx, e.Arg(0))
	return Result{Data: results, Text: func(w *Writer) {
		rows := [][]string{}
		for _, name := range sortedKeys(results) {
			rows = append(rows, []string{name, delayText(results[name])})
		}
		w.Table(nil, rows)
	}}, err
}

func profileCommand() *Command {
	list := func(e *Env) (Result, error) {
		k, err := kumo(e)
		if err != nil {
			return Result{}, err
		}
		profiles, err := k.Profiles.List()
		return Result{Data: profiles, Text: func(w *Writer) {
			rows := [][]string{}
			for _, p := range profiles {
				mark := " "
				if p.Current {
					mark = "*"
				}
				updated := "-"
				if p.UpdatedAt != nil {
					updated = p.UpdatedAt.Local().Format("2006-01-02 15:04")
				}
				rows = append(rows, []string{mark, p.ID, p.Name, p.Source, updated})
			}
			w.Table([]string{"", "ID", "NAME", "SOURCE", "UPDATED"}, rows)
		}}, err
	}
	summary := func(s profile.Summary) func(*Writer) {
		return func(w *Writer) {
			w.Status("ok", s.Name+" ("+s.ID+")")
			w.Field("source", s.Source)
			if s.Usage != nil && s.Usage.Total > 0 {
				w.Field("usage", fmt.Sprintf("%.1f / %.1f GB", float64(s.Usage.Used())/1e9, float64(s.Usage.Total)/1e9))
			}
		}
	}
	return &Command{Name: "profile", Aliases: []string{"profiles"}, Summary: "List, import, refresh and switch profiles", Usage: "kumo profile <list|current|use|import|refresh|show|delete>", Sub: []*Command{
		{Name: "list", Aliases: []string{"ls"}, Summary: "List profiles", Run: list},
		{Name: "current", Summary: "Show the current profile", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			s, err := k.Lifecycle.Status()
			return Result{Data: s.Profile, Text: summary(s.Profile)}, err
		}},
		{Name: "use", Aliases: []string{"switch"}, Summary: "Make a profile current", Usage: "kumo profile use <id>", Run: func(e *Env) (Result, error) {
			if e.Arg(0) == "" {
				return Result{}, usageError("usage: kumo profile use <id>")
			}
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			if err := k.Profiles.Select(e.Ctx, e.Arg(0)); err != nil {
				return Result{}, err
			}
			s, err := k.Profiles.Get(e.Arg(0))
			return Result{Data: s, Text: summary(s)}, err
		}},
		{Name: "import", Aliases: []string{"add"}, Summary: "Import a profile from a URL or a file", Usage: "kumo profile import <url|path> [--name <name>] [--use-proxy] [--activate]", Flags: []string{"name=", "use-proxy", "activate"}, Run: func(e *Env) (Result, error) {
			source := e.Arg(0)
			if source == "" {
				return Result{}, usageError("usage: kumo profile import <url|path>")
			}
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			name, _ := e.Flag("name")
			req := application.ImportRequest{Name: name, UseProxy: e.Bool("use-proxy"), Activate: e.Bool("activate")}
			if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
				req.URL = source
			} else {
				req.Path = source
			}
			s, err := k.Profiles.Import(e.Ctx, req)
			return Result{Data: s, Text: summary(s)}, err
		}},
		{Name: "refresh", Aliases: []string{"update"}, Summary: "Download remote profiles again", Usage: "kumo profile refresh <id> | --due", Flags: []string{"due"}, Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			if e.Bool("due") || e.Arg(0) == "" {
				refreshed, err := k.Profiles.RefreshDue(e.Ctx)
				if refreshed == nil {
					refreshed = []profile.Summary{}
				}
				return Result{Data: refreshed, Text: func(w *Writer) { w.Status("ok", fmt.Sprintf("refreshed %d profiles", len(refreshed))) }}, err
			}
			s, err := k.Profiles.Refresh(e.Ctx, e.Arg(0))
			return Result{Data: s, Text: summary(s)}, err
		}},
		{Name: "show", Summary: "Print a profile's YAML", Usage: "kumo profile show <id>", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			id := e.Arg(0)
			if id == "" {
				s, _ := k.Lifecycle.Status()
				id = s.Profile.ID
			}
			yaml, err := k.Profiles.Content(id)
			return Result{Data: map[string]string{"id": id, "yaml": yaml}, Text: func(w *Writer) { w.Line(strings.TrimRight(yaml, "\n")) }}, err
		}},
		{Name: "delete", Aliases: []string{"rm"}, Summary: "Delete a profile", Usage: "kumo profile delete <id>", Run: func(e *Env) (Result, error) {
			if e.Arg(0) == "" {
				return Result{}, usageError("usage: kumo profile delete <id>")
			}
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			if e.DryRun {
				return Result{Data: map[string]string{"wouldDelete": e.Arg(0)}}, nil
			}
			err = k.Profiles.Delete(e.Ctx, e.Arg(0))
			return Result{Data: map[string]string{"deleted": e.Arg(0)}, Text: func(w *Writer) { w.Status("ok", "deleted "+e.Arg(0)) }}, err
		}},
	}}
}

func overrideCommand() *Command {
	set := func(disabled bool) func(*Env) (Result, error) {
		return func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			err = k.Overrides.SetDisabled(e.Ctx, e.Arg(0), disabled)
			return Result{Data: map[string]any{"id": e.Arg(0), "disabled": disabled}}, err
		}
	}
	return &Command{Name: "override", Aliases: []string{"overrides"}, Summary: "Manage overrides applied on top of every profile", Usage: "kumo override <list|add|show|enable|disable|delete|order>", Sub: []*Command{
		{Name: "list", Aliases: []string{"ls"}, Summary: "List overrides in the order they apply", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			items, err := k.Overrides.List()
			return Result{Data: items, Text: func(w *Writer) {
				rows := [][]string{}
				for i, o := range items {
					state := "on"
					if o.Disabled {
						state = "off"
					}
					rows = append(rows, []string{strconv.Itoa(i + 1), o.ID, o.Name, string(o.Format), state})
				}
				w.Table([]string{"#", "ID", "NAME", "FORMAT", "STATE"}, rows)
			}}, err
		}},
		{Name: "add", Summary: "Add an override from a file, a URL or text", Usage: "kumo override add (--file <path> | --url <url> | --yaml <text>) [--name <name>] [--format yaml|javascript]", Flags: []string{"file=", "url=", "yaml=", "name=", "format="}, Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			name, _ := e.Flag("name")
			format, _ := e.Flag("format")
			req := application.SaveRequest{Name: name, Format: override.Format(format)}
			if path, ok := e.Flag("file"); ok {
				data, err := os.ReadFile(path)
				if err != nil {
					return Result{}, err
				}
				req.Content = string(data)
				if req.Name == "" {
					req.Name = strings.TrimSuffix(path[strings.LastIndexAny(path, `/\`)+1:], ".yaml")
				}
				if req.Format == "" && strings.HasSuffix(path, ".js") {
					req.Format = override.FormatJavaScript
				}
			} else if text, ok := e.Flag("yaml"); ok {
				req.Content = text
			} else if url, ok := e.Flag("url"); ok {
				req.URL = url
			} else {
				return Result{}, usageError("give --file, --url or --yaml")
			}
			item, err := k.Overrides.Save(e.Ctx, req)
			return Result{Data: item, Text: func(w *Writer) { w.Status("ok", "added "+item.Name+" ("+item.ID+")") }}, err
		}},
		{Name: "show", Summary: "Print an override", Usage: "kumo override show <id>", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			detail, err := k.Overrides.Get(e.Arg(0))
			return Result{Data: detail, Text: func(w *Writer) { w.Line(detail.Content) }}, err
		}},
		{Name: "enable", Summary: "Apply an override again", Usage: "kumo override enable <id>", Run: set(false)},
		{Name: "disable", Summary: "Stop applying an override", Usage: "kumo override disable <id>", Run: set(true)},
		{Name: "delete", Aliases: []string{"rm"}, Summary: "Delete an override", Usage: "kumo override delete <id>", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			err = k.Overrides.Delete(e.Ctx, e.Arg(0))
			return Result{Data: map[string]string{"deleted": e.Arg(0)}}, err
		}},
		{Name: "order", Summary: "Apply overrides in this order", Usage: "kumo override order <id>...", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			err = k.Overrides.Reorder(e.Ctx, e.Args)
			return Result{Data: map[string][]string{"order": e.Args}}, err
		}},
	}}
}

func runConnections(e *Env) (Result, error) {
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	if id, ok := e.Flag("close"); ok {
		err := k.Inspect.CloseConnection(e.Ctx, id)
		return Result{Data: map[string]string{"closed": id}}, err
	}
	if e.Bool("close-all") {
		err := k.Inspect.CloseConnection(e.Ctx, "")
		return Result{Data: map[string]string{"closed": "all"}, Text: func(w *Writer) { w.Status("ok", "closed every connection") }}, err
	}
	conns, err := k.Inspect.Connections(e.Ctx)
	if err != nil {
		return Result{}, err
	}
	limit, err := e.Int("limit", 0)
	if err != nil {
		return Result{}, err
	}
	if limit > 0 && len(conns.Connections) > limit {
		conns.Connections = conns.Connections[:limit]
	}
	return Result{Data: conns, Text: func(w *Writer) {
		rows := [][]string{}
		for _, c := range conns.Connections {
			rows = append(rows, []string{c.Host, c.Process, c.Rule, strings.Join(c.Chain, " ← "), bytes(c.Download)})
		}
		w.Table([]string{"HOST", "PROCESS", "RULE", "CHAIN", "DOWN"}, rows)
	}}, nil
}

func bytes(n int64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(n)/(1<<10))
	}
	return fmt.Sprintf("%d B", n)
}

func runRules(e *Env) (Result, error) {
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	for flag, enabled := range map[string]bool{"disable": false, "enable": true} {
		if _, ok := e.Flag(flag); ok {
			index, err := e.Int(flag, 0)
			if err != nil {
				return Result{}, err
			}
			err = k.Inspect.SetRuleEnabled(e.Ctx, index, enabled)
			return Result{Data: map[string]any{"index": index, "enabled": enabled}}, err
		}
	}
	rules, err := k.Inspect.Rules(e.Ctx)
	return Result{Data: rules, Text: func(w *Writer) {
		rows := [][]string{}
		for _, r := range rules {
			state := ""
			if !r.Enabled {
				state = "off"
			}
			rows = append(rows, []string{strconv.Itoa(r.Index), r.Type, r.Payload, r.Proxy, state})
		}
		w.Table([]string{"#", "TYPE", "PAYLOAD", "POLICY", ""}, rows)
	}}, err
}

func providersCommand() *Command {
	return &Command{Name: "providers", Summary: "List and update proxy and rule providers", Usage: "kumo providers [--json]\nkumo providers update <proxy|rule> <name>\nkumo providers geo-update", Run: func(e *Env) (Result, error) {
		k, err := kumo(e)
		if err != nil {
			return Result{}, err
		}
		proxies, rules, err := k.Inspect.Providers(e.Ctx)
		data := map[string]any{"proxy": proxies, "rule": rules}
		return Result{Data: data, Text: func(w *Writer) {
			rows := [][]string{}
			for _, p := range proxies {
				rows = append(rows, []string{"proxy", p.Name, p.VehicleType, strconv.Itoa(p.Count), ago(p.UpdatedAt)})
			}
			for _, r := range rules {
				rows = append(rows, []string{"rule", r.Name, r.Behavior, strconv.Itoa(r.Count), ago(r.UpdatedAt)})
			}
			w.Table([]string{"KIND", "NAME", "TYPE", "COUNT", "UPDATED"}, rows)
		}}, err
	}, Sub: []*Command{
		{Name: "update", Summary: "Refresh a provider", Usage: "kumo providers update <proxy|rule> <name>", Run: func(e *Env) (Result, error) {
			if e.Arg(0) != "proxy" && e.Arg(0) != "rule" || e.Arg(1) == "" {
				return Result{}, usageError("usage: kumo providers update <proxy|rule> <name>")
			}
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			err = k.Inspect.UpdateProvider(e.Ctx, e.Arg(0), e.Arg(1))
			return Result{Data: map[string]string{"updated": e.Arg(1)}}, err
		}},
		{Name: "geo-update", Summary: "Download fresh GeoIP and GeoSite data", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			err = k.Inspect.UpgradeGeoData(e.Ctx)
			return Result{Data: map[string]bool{"updated": err == nil}}, err
		}},
	}}
}

func ago(t time.Time) string {
	if t.IsZero() || t.Year() < 2000 {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04")
}

func logsCommand() *Command {
	return &Command{Name: "logs", Aliases: []string{"log"}, Summary: "Show Mihomo's log", Usage: "kumo logs [--limit <n>] [--level <level>] [--follow] [--json]\nkumo logs path", Flags: []string{"limit=", "level=", "follow", "f"}, Run: func(e *Env) (Result, error) {
		k, err := kumo(e)
		if err != nil {
			return Result{}, err
		}
		level, _ := e.Flag("level")
		if e.Bool("follow") || e.Bool("f") {
			if level == "" {
				level = "info"
			}
			unsubscribe := k.Events.Subscribe(func(ev application.Event) {
				if l, ok := ev.(application.LogReceived); ok {
					if e.JSON {
						(&Writer{out: e.Stdout}).Value(l.Entry)
					} else {
						fmt.Fprintf(e.Stdout, "%s %-7s %s\n", l.Entry.Time.Format("15:04:05"), l.Entry.Level, l.Entry.Message)
					}
				}
			})
			defer unsubscribe()
			k.Inspect.StreamLogs(e.Ctx, level)
			return Result{Text: func(*Writer) {}}, nil
		}
		limit, err := e.Int("limit", 100)
		if err != nil {
			return Result{}, err
		}
		entries, err := k.Inspect.RecentLogs(limit * 4)
		if err != nil {
			return Result{}, err
		}
		entries = filterLevel(entries, level)
		if len(entries) > limit {
			entries = entries[len(entries)-limit:]
		}
		return Result{Data: entries, Text: func(w *Writer) {
			for _, entry := range entries {
				w.Line(fmt.Sprintf("%-7s %s", entry.Level, entry.Message))
			}
		}}, nil
	}, Sub: []*Command{
		{Name: "path", Summary: "Print the log directory", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			dir := k.Paths().Logs()
			return Result{Data: map[string]string{"path": dir, "core": k.Paths().CoreLog()}, Text: func(w *Writer) { w.Line(dir) }}, nil
		}},
	}}
}

var levelRank = map[string]int{"debug": 0, "info": 1, "kumo": 1, "warning": 2, "error": 3}

func filterLevel(entries []inspect.LogEntry, level string) []inspect.LogEntry {
	if level == "" {
		return entries
	}
	if level == "warn" {
		level = "warning"
	}
	min := levelRank[level]
	out := []inspect.LogEntry{}
	for _, entry := range entries {
		if levelRank[entry.Level] >= min {
			out = append(out, entry)
		}
	}
	return out
}

func sysproxyCommand() *Command {
	status := func(e *Env) (Result, error) {
		k, err := kumo(e)
		if err != nil {
			return Result{}, err
		}
		report, err := k.Doctor(e.Ctx)
		data := map[string]any{"systemProxy": report.Status.SystemProxy, "observed": report.SystemProxy}
		return Result{Data: data, Text: func(w *Writer) { statusText(report.Status)(w) }}, err
	}
	toggle := func(enabled bool) func(*Env) (Result, error) {
		return func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			if e.DryRun {
				s, err := k.Lifecycle.Status()
				settings := s.SystemProxy.Settings
				settings.Port = s.MixedPort
				return Result{Data: map[string]any{"enabled": enabled, "settings": settings, "appliesNow": s.Running()}, Text: func(w *Writer) {
					w.Status("dry-run", fmt.Sprintf("would turn the system proxy %s (%s %s:%d on %s)", map[bool]string{true: "on", false: "off"}[enabled], settings.Mode, settings.Host, settings.Port, settings.NetworkService))
				}}, err
			}
			s, err := k.SystemProxy.SetEnabled(enabled)
			return Result{Data: s.SystemProxy, Text: statusText(s)}, err
		}
	}
	return &Command{Name: "sysproxy", Aliases: []string{"system-proxy"}, Summary: "Point the system proxy at Kumo, or release it", Usage: "kumo sysproxy <status|on|off|services|set>", Sub: []*Command{
		{Name: "status", Summary: "Show the system proxy and who owns it", Run: status},
		{Name: "on", Aliases: []string{"enable"}, Summary: "Turn the system proxy on (now, if Kumo runs)", Run: toggle(true)},
		{Name: "off", Aliases: []string{"disable"}, Summary: "Turn the system proxy off", Run: toggle(false)},
		{Name: "services", Summary: "List network services", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			services, err := k.SystemProxy.Services()
			return Result{Data: services, Text: func(w *Writer) {
				for _, s := range services {
					w.Line(s)
				}
			}}, err
		}},
		{Name: "set", Summary: "Change system proxy settings", Usage: "kumo sysproxy set [--mode manual|pac] [--service <name|Automatic>] [--host <host>] [--bypass <a,b,...>] [--pac-file <path>] [--reclaim on|off]", Flags: []string{"mode=", "service=", "host=", "bypass=", "pac-file=", "reclaim="}, Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			if v, ok := e.Flag("reclaim"); ok {
				if _, err := k.Preferences.Update(func(p *prefs) { p.ReclaimSystemProxy = v == "on" || v == "true" }); err != nil {
					return Result{}, err
				}
			}
			s, err := k.Lifecycle.Status()
			if err != nil {
				return Result{}, err
			}
			settings := s.SystemProxy.Settings
			if v, ok := e.Flag("mode"); ok {
				settings.Mode = systemproxy.Mode(v)
			}
			if v, ok := e.Flag("service"); ok {
				settings.NetworkService = v
			}
			if v, ok := e.Flag("host"); ok {
				settings.Host = v
			}
			if v, ok := e.Flag("bypass"); ok {
				settings.BypassList = strings.Split(v, ",")
			}
			if v, ok := e.Flag("pac-file"); ok {
				data, err := os.ReadFile(v)
				if err != nil {
					return Result{}, err
				}
				settings.PACScript = string(data)
			}
			s, err = k.SystemProxy.UpdateSettings(settings)
			return Result{Data: s.SystemProxy, Text: statusText(s)}, err
		}},
	}}
}

func tunCommand() *Command {
	toggle := func(enabled bool) func(*Env) (Result, error) {
		return func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			s, err := k.Settings.SetTUN(e.Ctx, enabled)
			return Result{Data: s, Text: statusText(s)}, err
		}
	}
	return &Command{Name: "tun", Summary: "Route all traffic through TUN", Usage: "kumo tun <status|on|off|grant>", Sub: []*Command{
		{Name: "status", Summary: "Show TUN settings and permission", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			settings, err := k.Settings.Get()
			data := map[string]any{"enabled": settings.Settings.TUN.Enabled, "settings": settings.Settings.TUN, "permission": settings.TUNPermission}
			return Result{Data: data, Text: func(w *Writer) {
				w.Field("tun", map[bool]string{true: "on", false: "off"}[settings.Settings.TUN.Enabled])
				w.Field("permission", map[bool]string{true: "granted (" + settings.TUNPermission.Method + ")", false: "not granted; run `kumo tun grant`"}[settings.TUNPermission.Granted])
			}}, err
		}},
		{Name: "on", Aliases: []string{"enable"}, Summary: "Turn TUN on", Run: toggle(true)},
		{Name: "off", Aliases: []string{"disable"}, Summary: "Turn TUN off", Run: toggle(false)},
		{Name: "grant", Summary: "Allow the core to create a TUN interface (asks for a password)", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			p, err := k.Settings.GrantTUN(e.Ctx)
			return Result{Data: p}, err
		}},
	}}
}

func managedSection(name, summary string, enabled func(*core.RuntimeSettings) *bool, view func(core.RuntimeSettings) any) *Command {
	toggle := func(on bool) func(*Env) (Result, error) {
		return func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			settings, err := k.Settings.Get()
			if err != nil {
				return Result{}, err
			}
			*enabled(&settings.Settings) = on
			s, err := k.Settings.Update(e.Ctx, settings.Settings)
			return Result{Data: view(settings.Settings), Text: statusText(s)}, err
		}
	}
	return &Command{Name: name, Summary: summary, Usage: "kumo " + name + " <show|on|off>", Sub: []*Command{
		{Name: "show", Summary: "Show the settings", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			settings, err := k.Settings.Get()
			return Result{Data: view(settings.Settings)}, err
		}},
		{Name: "on", Aliases: []string{"enable"}, Summary: "Use Kumo's settings instead of the profile's", Run: toggle(true)},
		{Name: "off", Aliases: []string{"disable"}, Summary: "Use the profile's own settings", Run: toggle(false)},
	}}
}

func configCommand() *Command {
	return &Command{Name: "config", Aliases: []string{"c"}, Summary: "Show the generated configuration and change core settings", Usage: "kumo config <show|path|set>", Sub: []*Command{
		{Name: "show", Summary: "Print the configuration Mihomo runs with", Flags: []string{"settings"}, Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			settings, err := k.Settings.Get()
			if err != nil {
				return Result{}, err
			}
			if e.Bool("settings") || e.JSON {
				return Result{Data: settings}, nil
			}
			rt, _ := storage.NewRuntimeStore(k.Paths()).Load()
			yaml, err := k.Lifecycle.RuntimeConfig(rt)
			return Result{Data: map[string]string{"yaml": yaml}, Text: func(w *Writer) { w.Line(strings.TrimRight(yaml, "\n")) }}, err
		}},
		{Name: "path", Summary: "Print where Kumo keeps its data", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			p := k.Paths()
			data := map[string]string{"data": p.Root, "runtimeConfig": p.RuntimeConfig(), "state": p.State(), "profiles": p.Profiles(), "logs": p.Logs()}
			return Result{Data: data, Text: func(w *Writer) {
				for _, key := range sortedKeys(data) {
					w.Field(key, data[key])
				}
			}}, nil
		}},
		{Name: "set", Summary: "Change a core setting", Usage: "kumo config set <mixed-port|allow-lan|log-level|ipv6|find-process-mode|controller> <value> [--secret <secret>]", Flags: []string{"secret="}, Run: runConfigSet},
	}}
}

func onOff(value string) (bool, error) {
	switch value {
	case "on", "true", "yes", "1":
		return true, nil
	case "off", "false", "no", "0":
		return false, nil
	}
	return false, usageError("use on or off, not %q", value)
}

func runConfigSet(e *Env) (Result, error) {
	key, value := e.Arg(0), e.Arg(1)
	if key == "" || value == "" {
		return Result{}, usageError("usage: kumo config set <key> <value>")
	}
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	settings, err := k.Settings.Get()
	if err != nil {
		return Result{}, err
	}
	s := settings.Settings
	switch key {
	case "mixed-port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return Result{}, usageError("mixed-port takes a number")
		}
		s.MixedPort = port
	case "allow-lan":
		if s.AllowLAN, err = onOff(value); err != nil {
			return Result{}, err
		}
	case "ipv6":
		if s.IPv6, err = onOff(value); err != nil {
			return Result{}, err
		}
	case "log-level":
		s.LogLevel = value
	case "find-process-mode":
		s.FindProcessMode = value
	case "controller":
		host, portText, ok := strings.Cut(value, ":")
		port, err := strconv.Atoi(portText)
		if !ok || err != nil {
			return Result{}, usageError("controller takes host:port")
		}
		var secret *string
		if v, ok := e.Flag("secret"); ok {
			secret = &v
		}
		status, err := k.Settings.SetController(e.Ctx, host, port, secret)
		return Result{Data: status, Text: statusText(status)}, err
	default:
		return Result{}, usageError("unknown setting %q", key)
	}
	status, err := k.Settings.Update(e.Ctx, s)
	return Result{Data: status, Text: statusText(status)}, err
}

func coreCommand() *Command {
	return &Command{Name: "core", Summary: "Find, choose and install the Mihomo core", Usage: "kumo core <status|list|use|install|auto>", Sub: []*Command{
		{Name: "status", Summary: "Show the core Kumo uses", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			settings, err := k.Settings.Get()
			data := map[string]any{"path": settings.CorePath, "candidates": settings.Candidates}
			if v, verr := k.Inspect.Version(e.Ctx); verr == nil {
				data["version"] = v.Version
			}
			return Result{Data: data, Text: func(w *Writer) {
				w.Field("core", settings.CorePath)
				if v, ok := data["version"]; ok {
					w.Field("version", v)
				}
			}}, err
		}},
		{Name: "list", Summary: "List Mihomo executables Kumo can find", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			settings, err := k.Settings.Get()
			return Result{Data: settings.Candidates, Text: func(w *Writer) {
				rows := [][]string{}
				for _, c := range settings.Candidates {
					rows = append(rows, []string{c.Path, c.Source})
				}
				w.Table([]string{"PATH", "SOURCE"}, rows)
			}}, err
		}},
		{Name: "use", Summary: "Use a specific executable", Usage: "kumo core use <path>", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			err = k.Settings.SetCorePath(e.Arg(0))
			return Result{Data: map[string]string{"path": e.Arg(0)}}, err
		}},
		{Name: "auto", Summary: "Go back to finding the core automatically", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			return Result{Data: map[string]string{"path": ""}}, k.Settings.SetCorePath("")
		}},
		{Name: "install", Summary: "Download the latest Mihomo for this machine", Run: func(e *Env) (Result, error) {
			k, err := kumo(e)
			if err != nil {
				return Result{}, err
			}
			installed, err := k.Settings.InstallCore(e.Ctx)
			return Result{Data: installed, Text: func(w *Writer) { w.Status("ok", "installed Mihomo "+installed.Version+" at "+installed.Path) }}, err
		}},
	}}
}

func runTopology(e *Env) (Result, error) {
	k, err := kumo(e)
	if err != nil {
		return Result{}, err
	}
	graph, err := k.Topology.Graph(e.Ctx)
	if err != nil {
		return Result{}, err
	}
	if e.Bool("issues") {
		return Result{Data: graph.Issues, Text: func(w *Writer) {
			if len(graph.Issues) == 0 {
				w.Status("ok", "every reference resolves")
			}
			for _, issue := range graph.Issues {
				w.Status("warn", issue.Message)
			}
		}}, nil
	}
	return Result{Data: graph, Text: func(w *Writer) {
		counts := map[topology.Kind]int{}
		for _, n := range graph.Nodes {
			counts[n.Kind]++
		}
		w.Field("mode", graph.Mode)
		w.Field("rule nodes", counts[topology.KindRules]+counts[topology.KindRuleSet]+counts[topology.KindFinal])
		w.Field("groups", counts[topology.KindGroup])
		w.Field("proxies", counts[topology.KindProxy])
		w.Field("providers", counts[topology.KindProvider])
		w.Field("issues", len(graph.Issues))
		active := []string{}
		for _, edge := range graph.Edges {
			if edge.Active && edge.Kind != topology.EdgeRoute {
				active = append(active, strings.SplitN(edge.Source, ":", 2)[1]+" → "+strings.SplitN(edge.Target, ":", 2)[len(strings.SplitN(edge.Target, ":", 2))-1])
			}
		}
		if len(active) > 0 {
			w.Line("")
			w.Heading("Active path")
			for _, a := range active {
				w.Line("  " + a)
			}
		}
	}}, nil
}

func skillsCommand() *Command {
	parse := func(e *Env) (*skills.Installer, []skills.Target, skills.Scope, error) {
		k, err := kumo(e)
		if err != nil {
			return nil, nil, "", err
		}
		installer, err := skills.New(k.Paths().AgentSkillsState())
		if err != nil {
			return nil, nil, "", err
		}
		scope := skills.ScopeGlobal
		if v, ok := e.Flag("scope"); ok {
			if v != "global" && v != "project" {
				return nil, nil, "", usageError("--scope is global or project")
			}
			scope = skills.Scope(v)
		}
		var targets []skills.Target
		if v, ok := e.Flag("agent"); ok && v != "all" {
			for _, name := range strings.Split(v, ",") {
				t, err := skills.ParseTarget(strings.TrimSpace(name))
				if err != nil {
					return nil, nil, "", usageError("%v", err)
				}
				targets = append(targets, t)
			}
		}
		return installer, targets, scope, nil
	}
	flags := []string{"agent=", "scope=", "force"}
	report := func(r skills.Report) func(*Writer) {
		return func(w *Writer) {
			kind := "ok"
			if r.DryRun {
				kind = "dry-run"
			}
			for _, root := range r.DestinationRoots {
				w.Status(kind, strings.Join(r.SkillIDs, ", ")+" → "+root)
			}
		}
	}
	return &Command{Name: "skills", Summary: "Install the Kumo skill for coding agents", Usage: "kumo skills <status|install|uninstall> [--agent <cursor|claude|codex|gemini|agents|all>] [--scope global|project] [--dry-run] [--force]", Sub: []*Command{
		{Name: "status", Summary: "Show where the skill is installed", Flags: flags, Run: func(e *Env) (Result, error) {
			installer, targets, scope, err := parse(e)
			if err != nil {
				return Result{}, err
			}
			status, err := installer.Status(targets, scope)
			return Result{Data: status, Text: func(w *Writer) {
				rows := [][]string{}
				for _, t := range status.Targets {
					state := "-"
					if t.Installed {
						state = map[bool]string{true: "installed", false: "outdated"}[t.UpToDate]
					}
					rows = append(rows, []string{string(t.Target), state, t.DestinationRoot})
				}
				w.Table([]string{"AGENT", "STATE", "PATH"}, rows)
			}}, err
		}},
		{Name: "install", Aliases: []string{"add"}, Summary: "Install the skill", Flags: flags, Run: func(e *Env) (Result, error) {
			installer, targets, scope, err := parse(e)
			if err != nil {
				return Result{}, err
			}
			r, err := installer.Install(targets, scope, e.DryRun, e.Bool("force"))
			return Result{Data: r, Text: report(r)}, err
		}},
		{Name: "uninstall", Aliases: []string{"remove"}, Summary: "Remove the skill Kumo installed", Flags: flags, Run: func(e *Env) (Result, error) {
			installer, targets, scope, err := parse(e)
			if err != nil {
				return Result{}, err
			}
			r, err := installer.Uninstall(targets, scope, e.DryRun)
			return Result{Data: r, Text: report(r)}, err
		}},
	}}
}

func runPrompt(e *Env) (Result, error) {
	if e.Bool("list") || e.Arg(0) == "" {
		guides := application.Guides()
		return Result{Data: guides, Text: func(w *Writer) {
			for _, g := range guides {
				ids := []string{}
				for _, i := range g.Intents {
					ids = append(ids, i.ID)
				}
				w.Field(g.Page, strings.Join(ids, ", "))
			}
		}}, nil
	}
	intent, _ := e.Flag("intent")
	prompt, err := application.Prompt(e.Arg(0), intent)
	return Result{Data: map[string]string{"prompt": prompt}, Text: func(w *Writer) { w.Line(prompt) }}, err
}

func runHelp(e *Env) (Result, error) {
	cmds := commands()
	path, cmd, _ := route(cmds, e.Args)
	if cmd == nil {
		return Result{Text: func(w *Writer) { w.out.Write([]byte(topHelp(cmds))) }}, nil
	}
	return Result{Text: func(w *Writer) { w.out.Write([]byte(commandHelp(path, cmd))) }}, nil
}
