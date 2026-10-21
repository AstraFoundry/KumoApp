package desktop

import (
	_ "embed"
	"log"
	"sync"

	"github.com/egoist/mygo"

	"github.com/AstraFoundry/KumoApp/internal/application"
	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/locales"
)

var (
	//go:embed assets/tray-on.png
	trayOn []byte
	//go:embed assets/tray-off.png
	trayOff []byte
)

// Tray is the menu bar item: the status, start and stop, the mode, the
// system proxy and TUN, without opening the window.
type Tray struct {
	k    *application.Kumo
	open func(route string)
	tray *mygo.Tray

	mu   sync.Mutex
	last string
}

// NewTray creates the tray item. It needs the app to be ready.
func NewTray(k *application.Kumo, open func(route string)) *Tray {
	t := &Tray{k: k, open: open}
	tray, err := mygo.NewTray(mygo.TrayOptions{Icon: trayOff, IconIsTemplate: true, ToolTip: "Kumo"})
	if err != nil {
		log.Println("tray:", err)
		return t
	}
	t.tray = tray
	if status, err := k.Lifecycle.Status(); err == nil {
		t.Update(status)
	}
	return t
}

// Update rebuilds the menu for a status.
func (t *Tray) Update(s application.Status) {
	if t.tray == nil {
		return
	}
	key := statusKey(s)
	t.mu.Lock()
	if key == t.last {
		t.mu.Unlock()
		return
	}
	t.last = key
	t.mu.Unlock()

	running := s.State == core.StateRunning
	headline := locales.TrayStopped()
	switch {
	case running:
		headline = locales.TrayRunning()
	case s.State == core.StateFailed:
		headline = locales.TrayFailed()
	}
	toggle := &mygo.MenuItem{Label: locales.TrayStart(), Accelerator: "CmdOrCtrl+Shift+S", Click: func(*mygo.MenuItem, *mygo.Window) {
		go Runtime{t.k}.Start(bg(), false)
	}}
	if running {
		toggle = &mygo.MenuItem{Label: locales.TrayStop(), Click: func(*mygo.MenuItem, *mygo.Window) {
			go t.k.Lifecycle.Stop(bg())
		}}
	}
	modes := []*mygo.MenuItem{}
	for _, m := range []struct {
		mode  core.Mode
		label string
	}{{core.ModeRule, locales.ModeRule()}, {core.ModeGlobal, locales.ModeGlobal()}, {core.ModeDirect, locales.ModeDirect()}} {
		mode := m.mode
		modes = append(modes, &mygo.MenuItem{Label: m.label, Type: mygo.MenuItemRadio, Checked: s.Mode == mode, Click: func(*mygo.MenuItem, *mygo.Window) {
			go t.k.Settings.SetMode(bg(), mode)
		}})
	}
	menu := mygo.NewMenu([]*mygo.MenuItem{
		{Label: headline, Disabled: true},
		toggle,
		mygo.Separator(),
		{Label: locales.TrayMode(), Submenu: modes},
		{Label: locales.TraySystemProxy(), Type: mygo.MenuItemCheckbox, Checked: s.SystemProxy.Enabled, Click: func(*mygo.MenuItem, *mygo.Window) {
			go t.k.SystemProxy.SetEnabled(!s.SystemProxy.Enabled)
		}},
		{Label: locales.TrayTun(), Type: mygo.MenuItemCheckbox, Checked: s.TUN, Click: func(*mygo.MenuItem, *mygo.Window) {
			go t.k.Settings.SetTUN(bg(), !s.TUN)
		}},
		mygo.Separator(),
		{Label: locales.TrayProxies(), Click: func(*mygo.MenuItem, *mygo.Window) { t.open("/proxies") }},
		{Label: locales.TrayOpen(), Click: func(*mygo.MenuItem, *mygo.Window) { t.open("") }},
		mygo.Separator(),
		{Label: locales.TrayQuit(), Accelerator: "CmdOrCtrl+Q", Click: func(*mygo.MenuItem, *mygo.Window) { mygo.App.Quit() }},
	})
	icon := trayOff
	if running {
		icon = trayOn
	}
	t.tray.SetIcon(icon, true)
	t.tray.SetMenu(menu)
}

func statusKey(s application.Status) string {
	return string(s.State) + string(s.Mode) + boolKey(s.SystemProxy.Enabled) + boolKey(s.TUN) + locales.TrayStart()
}

func boolKey(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
