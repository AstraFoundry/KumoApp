package desktop

import (
	"context"

	"github.com/egoist/mygo"

	"github.com/AstraFoundry/KumoApp/internal/application"
	"github.com/AstraFoundry/KumoApp/internal/domain/core"
	"github.com/AstraFoundry/KumoApp/internal/domain/inspect"
	"github.com/AstraFoundry/KumoApp/internal/domain/preferences"
)

// Events the pages listen to.
var (
	// StatusEvent carries the status whenever it changes.
	StatusEvent = mygo.NewEvent[application.Status]("status")
	// TrafficEvent carries throughput once a second while the core runs.
	TrafficEvent = mygo.NewEvent[inspect.Traffic]("traffic")
	// ContentionEvent reports another app changing the system proxy.
	ContentionEvent = mygo.NewEvent[application.SystemProxyContended]("contention")
	// CoreExitedEvent reports a core that ended on its own.
	CoreExitedEvent = mygo.NewEvent[core.Exit]("coreExited")
	// ProfilesEvent reports that profiles changed.
	ProfilesEvent = mygo.NewEvent[bool]("profilesChanged")
	// NavigateEvent asks the page to show a route, from the tray or a menu.
	NavigateEvent = mygo.NewEvent[string]("navigate")
	// QuittingEvent tells the page that Kumo is shutting down.
	QuittingEvent = mygo.NewEvent[bool]("quitting")
	// UpdateProgress carries an update's download progress, 0 to 1.
	UpdateProgress = mygo.NewEvent[float64]("updateProgress")
)

// Logs streams the core's log to a page.
type Logs struct{ k *application.Kumo }

// Follow sends log lines at level and above until the page stops
// listening.
func (l Logs) Follow(ctx context.Context, level string, lines *mygo.Channel[inspect.LogEntry]) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	unsubscribe := l.k.Events.Subscribe(func(e application.Event) {
		if entry, ok := e.(application.LogReceived); ok {
			if err := lines.Send(entry.Entry); err != nil {
				cancel()
			}
		}
	})
	defer unsubscribe()
	l.k.Inspect.StreamLogs(ctx, level)
	return nil
}

// Bind registers every service with MyGo and forwards application events
// to the pages. onPrefs runs when the page saves preferences.
func Bind(k *application.Kumo, onPrefs func(preferences.Preferences)) {
	mygo.Bind(Runtime{k})
	mygo.Bind(SystemProxy{k})
	mygo.Bind(Profiles{k})
	mygo.Bind(Overrides{k})
	mygo.Bind(Proxies{k})
	mygo.Bind(&Inspect{k: k})
	mygo.Bind(Logs{k})
	mygo.Bind(Settings{k})
	mygo.Bind(Topology{k})
	mygo.Bind(Preferences{k: k, onSave: onPrefs})
	mygo.Bind(&App{k: k})

	k.Events.Subscribe(func(e application.Event) {
		switch e := e.(type) {
		case application.StatusChanged:
			StatusEvent.Broadcast(e.Status)
		case application.TrafficSampled:
			TrafficEvent.Broadcast(e.Traffic)
		case application.SystemProxyContended:
			ContentionEvent.Broadcast(e)
		case application.CoreExited:
			CoreExitedEvent.Broadcast(e.Exit)
		case application.ProfilesChanged:
			ProfilesEvent.Broadcast(true)
		}
	})
}
