package cli

import (
	"fmt"
	"strings"

	"github.com/AstraFoundry/KumoApp/internal/domain/preferences"
)

type prefs = preferences.Preferences

func topHelp(cmds []*Command) string {
	var b strings.Builder
	b.WriteString("kumo <command>\n\nUsage:\n\n")
	for _, line := range [][2]string{
		{"kumo status --json", "show whether Kumo runs, its mode and system proxy"},
		{"kumo start", "start Mihomo with the current profile"},
		{"kumo doctor --json", "check the core, configuration, ports and system proxy"},
		{"kumo proxies", "list proxy groups and selected nodes"},
		{"kumo prompt overview", "print a ready-made prompt for a coding agent"},
		{"kumo <command> -h", "quick help on a command"},
		{"kumo -l", "usage of every command"},
	} {
		fmt.Fprintf(&b, "%-28s %s\n", line[0], line[1])
	}
	b.WriteString("\nAll commands:\n\n")
	names := []string{}
	for _, c := range cmds {
		names = append(names, c.Name)
	}
	b.WriteString(wrap(strings.Join(names, ", "), 72, "    "))
	b.WriteString("\n\nEvery command takes --json and prints {\"ok\", \"data\", \"error\"}.\n")
	fmt.Fprintf(&b, "\nkumo@%s\n", Version)
	return b.String()
}

func wrap(text string, width int, indent string) string {
	var b strings.Builder
	line := indent
	for _, word := range strings.Fields(text) {
		if len(line)+len(word)+1 > width && strings.TrimSpace(line) != "" {
			b.WriteString(strings.TrimRight(line, " ") + "\n")
			line = indent
		}
		line += word + " "
	}
	b.WriteString(strings.TrimRight(line, " "))
	return b.String()
}

func longHelp(cmds []*Command) string {
	var b strings.Builder
	b.WriteString(topHelp(cmds))
	for _, c := range cmds {
		b.WriteString("\n")
		b.WriteString(commandHelp([]string{c.Name}, c))
	}
	return b.String()
}

func commandHelp(path []string, c *Command) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n\n", c.Summary)
	usage := c.Usage
	if usage == "" {
		usage = "kumo " + strings.Join(path, " ")
	}
	b.WriteString("Usage:\n")
	for _, line := range strings.Split(usage, "\n") {
		b.WriteString("  " + line + "\n")
	}
	if len(c.Sub) > 0 {
		b.WriteString("\nCommands:\n")
		for _, s := range c.Sub {
			fmt.Fprintf(&b, "  %-12s %s\n", s.Name, s.Summary)
		}
	}
	if len(c.Aliases) > 0 {
		fmt.Fprintf(&b, "\nAliases: %s\n", strings.Join(c.Aliases, ", "))
	}
	b.WriteString("\nOptions: --json, --dry-run where it applies, --color <always|auto|never>\n")
	return b.String()
}

func runCompletion(e *Env) (Result, error) {
	names := []string{}
	for _, c := range commands() {
		names = append(names, c.Name)
	}
	words := strings.Join(names, " ")
	var script string
	switch e.Arg(0) {
	case "zsh":
		script = "#compdef kumo\n_arguments '1: :(" + words + ")'\n"
	case "bash":
		script = "complete -W \"" + words + "\" kumo\n"
	case "fish":
		script = "complete -c kumo -f -a \"" + words + "\"\n"
	default:
		return Result{}, usageError("usage: kumo completion <zsh|bash|fish>")
	}
	return Result{Data: map[string]string{"script": script}, Text: func(w *Writer) { fmt.Fprint(w.out, script) }}, nil
}
