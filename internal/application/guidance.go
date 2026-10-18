package application

import (
	"fmt"
	"strings"
)

// Intent is one thing a user may want an agent to do on a page, with the
// kumo commands that do it.
type Intent struct {
	ID    string `json:"id"`
	Goal  string `json:"goal"`
	Steps []Step `json:"steps"`
}

// Step is one command and why it runs.
type Step struct {
	Command string `json:"command"`
	Why     string `json:"why"`
}

// Guide is the intents of one page.
type Guide struct {
	Page    string   `json:"page"`
	Intents []Intent `json:"intents"`
}

var guides = []Guide{
	{"overview", []Intent{
		{"connect", "Start Kumo and route this Mac's traffic through it", []Step{
			{"kumo status --json", "see whether Mihomo already runs"},
			{"kumo start --json", "start Mihomo with the current profile"},
			{"kumo sysproxy on --json", "point the system proxy at Kumo"},
			{"kumo status --json", "confirm state is running and systemProxy.active is true"},
		}},
		{"diagnose", "Find out why Kumo is not connecting", []Step{
			{"kumo doctor --json", "collect runtime, profile, core and port checks in one report"},
			{"kumo logs --limit 50 --level warning --json", "read recent warnings and errors from Mihomo"},
			{"kumo sysproxy status --json", "check whether another app took over the system proxy"},
		}},
		{"mode", "Switch the outbound mode", []Step{
			{"kumo mode --json", "read the current mode"},
			{"kumo mode rule --json", "use rule, global or direct"},
		}},
	}},
	{"profiles", []Intent{
		{"import", "Import a subscription and use it", []Step{
			{"kumo profile import '<subscription URL>' --activate --json", "download, validate and switch to the profile"},
			{"kumo proxies --json", "check that its proxy groups are loaded"},
		}},
		{"refresh", "Refresh subscriptions", []Step{
			{"kumo profile list --json", "find remote profiles and their last update"},
			{"kumo profile refresh <id> --json", "download one again, or use --due for every profile past its interval"},
		}},
		{"switch", "Switch to another profile", []Step{
			{"kumo profile list --json", "find the profile id"},
			{"kumo profile use <id> --json", "make it current; a running core reloads without restarting"},
		}},
	}},
	{"proxies", []Intent{
		{"fastest", "Pick the fastest node of a group", []Step{
			{"kumo proxies --json", "list groups and their nodes"},
			{"kumo delay '<group>' --json", "measure every node of the group; -1 is a timeout"},
			{"kumo select '<group>' '<node>' --json", "use the node with the lowest delay"},
		}},
		{"select", "Use a specific node", []Step{
			{"kumo proxies --json", "find the group and node names"},
			{"kumo select '<group>' '<node>' --json", "select it; connections on the old node close"},
		}},
	}},
	{"topology", []Intent{
		{"explain", "Explain how traffic flows through my configuration", []Step{
			{"kumo topology --json", "read the graph: mode, rules, groups, nodes and active path"},
			{"kumo topology --issues --json", "list references the configuration cannot resolve"},
		}},
		{"route", "Find which proxy a domain will use", []Step{
			{"kumo rules --json", "read the rules in order; the first match wins"},
			{"kumo proxies --json", "follow the matched policy group to its selected node"},
		}},
	}},
	{"connections", []Intent{
		{"inspect", "See which apps use the proxy", []Step{
			{"kumo connections --json", "list connections with host, process, rule and chain"},
		}},
		{"close", "Close stuck connections", []Step{
			{"kumo connections --close-all --json", "close every connection, or --close <id> for one"},
		}},
	}},
	{"logs", []Intent{
		{"errors", "Summarize recent errors", []Step{
			{"kumo logs --limit 200 --level warning --json", "read recent warnings and errors"},
		}},
		{"follow", "Watch the log live", []Step{
			{"kumo logs --follow --level info", "stream new lines until interrupted"},
		}},
	}},
	{"rules", []Intent{
		{"toggle", "Temporarily turn off a rule", []Step{
			{"kumo rules --json", "find the rule's index"},
			{"kumo rules --disable <index> --json", "turn it off until Mihomo restarts; --enable turns it back on"},
		}},
	}},
	{"core", []Intent{
		{"install", "Install or update the Mihomo core", []Step{
			{"kumo core status --json", "see which core Kumo uses"},
			{"kumo core install --json", "download the latest Mihomo for this machine"},
			{"kumo restart --json", "run the new core"},
		}},
		{"ports", "Change ports or allow the LAN", []Step{
			{"kumo config show --json", "read the current settings"},
			{"kumo config set mixed-port 7890 --json", "set the mixed port; also allow-lan, log-level, ipv6"},
		}},
	}},
	{"systemProxy", []Intent{
		{"enable", "Turn the system proxy on", []Step{
			{"kumo sysproxy on --dry-run --json", "preview the change"},
			{"kumo sysproxy on --json", "apply it while Kumo runs"},
		}},
		{"contention", "Check whether another app took over the proxy", []Step{
			{"kumo sysproxy status --json", "ownership is owned, takenOver or cleared"},
			{"kumo sysproxy set --reclaim on --json", "let Kumo put its proxy back automatically"},
		}},
	}},
	{"dns", []Intent{
		{"manage", "Let Kumo manage DNS", []Step{
			{"kumo dns show --json", "read Kumo's DNS settings"},
			{"kumo dns on --json", "replace the profile's dns section with Kumo's"},
		}},
	}},
	{"tun", []Intent{
		{"enable", "Route all traffic with TUN", []Step{
			{"kumo tun status --json", "check whether the core may create a TUN interface"},
			{"kumo tun grant", "ask for the permission once (prompts for a password)"},
			{"kumo tun on --json", "enable TUN; the core restarts if needed"},
		}},
	}},
	{"sniffer", []Intent{
		{"manage", "Let Kumo manage the sniffer", []Step{
			{"kumo sniffer show --json", "read Kumo's sniffer settings"},
			{"kumo sniffer on --json", "replace the profile's sniffer section with Kumo's"},
		}},
	}},
	{"resources", []Intent{
		{"update", "Update providers and geo data", []Step{
			{"kumo providers --json", "list proxy and rule providers"},
			{"kumo providers update rule '<name>' --json", "refresh one provider"},
			{"kumo providers geo-update --json", "download fresh GeoIP and GeoSite data"},
		}},
	}},
	{"overrides", []Intent{
		{"add", "Add a rule to every profile", []Step{
			{`kumo override add --name 'Extra rules' --yaml $'+rules:\n  - DOMAIN-SUFFIX,example.com,DIRECT' --json`, "prepend rules with an override"},
			{"kumo config show --json", "check the merged configuration"},
		}},
		{"order", "Reorder or turn off overrides", []Step{
			{"kumo override list --json", "list overrides in the order they apply"},
			{"kumo override disable <id> --json", "turn one off without deleting it"},
		}},
	}},
	{"agents", []Intent{
		{"skills", "Install the Kumo skill for coding agents", []Step{
			{"kumo skills status --json", "see which agents have it"},
			{"kumo skills install --agent all --dry-run --json", "preview where it goes"},
			{"kumo skills install --agent all --json", "install it"},
		}},
	}},
	{"settings", []Intent{
		{"cli", "Install the kumo command", []Step{
			{"kumo cli-link status --json", "see whether kumo is on PATH"},
		}},
	}},
}

// Guides lists every page's intents.
func Guides() []Guide { return guides }

// GuideFor returns a page's intents.
func GuideFor(page string) (Guide, bool) {
	for _, g := range guides {
		if g.Page == page {
			return g, true
		}
	}
	return Guide{}, false
}

// Prompt is a message to paste into a coding agent: what Kumo is, the
// rules of its CLI, and the steps of one intent.
func Prompt(page, intent string) (string, error) {
	guide, ok := GuideFor(page)
	if !ok {
		return "", fmt.Errorf("no guidance for page %q", page)
	}
	chosen := guide.Intents[0]
	for _, i := range guide.Intents {
		if i.ID == intent {
			chosen = i
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Goal: %s.\n\n", chosen.Goal)
	b.WriteString("Use Kumo's command-line tool, `kumo`, which controls the Kumo app and its Mihomo core on this computer.\n")
	b.WriteString("- Add --json to every command and read `ok`, `data` and `error` from the result; exit code 0 is success.\n")
	b.WriteString("- If `kumo` is not on PATH, ask me to install it from Kumo › Settings, or run it from the app bundle.\n")
	b.WriteString("- Run `kumo <command> -h` when unsure of a command's options. Do not edit Kumo's files directly.\n\n")
	b.WriteString("Steps:\n")
	for n, step := range chosen.Steps {
		fmt.Fprintf(&b, "%d. `%s` — %s.\n", n+1, step.Command, step.Why)
	}
	b.WriteString("\nReport what you changed and the final `kumo status --json`.")
	return b.String(), nil
}
