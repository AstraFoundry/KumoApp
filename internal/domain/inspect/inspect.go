// Package inspect models what a running core reports about its work:
// connections, rules, logs, traffic and providers.
package inspect

import "time"

// Connection is one open connection through the core.
type Connection struct {
	ID          string    `json:"id"`
	Host        string    `json:"host"`
	Destination string    `json:"destination"`
	Network     string    `json:"network"`
	Process     string    `json:"process,omitempty"`
	ProcessPath string    `json:"processPath,omitempty"`
	Rule        string    `json:"rule,omitempty"`
	RulePayload string    `json:"rulePayload,omitempty"`
	Chain       []string  `json:"chain"`
	Upload      int64     `json:"upload"`
	Download    int64     `json:"download"`
	UpSpeed     int64     `json:"uploadSpeed"`
	DownSpeed   int64     `json:"downloadSpeed"`
	Start       time.Time `json:"start"`
}

// Connections is a snapshot of every open connection with the totals.
type Connections struct {
	Connections   []Connection `json:"connections"`
	UploadTotal   int64        `json:"uploadTotal"`
	DownloadTotal int64        `json:"downloadTotal"`
}

// Rule is one routing rule of the running configuration.
type Rule struct {
	Index    int    `json:"index"`
	Type     string `json:"type"`
	Payload  string `json:"payload"`
	Proxy    string `json:"proxy"`
	Size     int    `json:"size"`
	Enabled  bool   `json:"enabled"`
	HitCount int64  `json:"hitCount"`
	HitAt    string `json:"hitAt,omitempty"`
}

// LogEntry is one line of the core's log.
type LogEntry struct {
	Level   string    `json:"level"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

// Traffic is the throughput at one moment, in bytes per second, with the
// totals since the core started.
type Traffic struct {
	Up        int64 `json:"up"`
	Down      int64 `json:"down"`
	UpTotal   int64 `json:"upTotal"`
	DownTotal int64 `json:"downTotal"`
}

// Memory is the core's memory use in bytes.
type Memory struct {
	InUse   int64 `json:"inuse"`
	OSLimit int64 `json:"oslimit"`
}

// ProxyProvider is a provider of proxies, usually a subscription.
type ProxyProvider struct {
	Name        string    `json:"name"`
	VehicleType string    `json:"vehicleType"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Count       int       `json:"count"`
	Usage       *Usage    `json:"usage,omitempty"`
}

// Usage is a provider's reported traffic.
type Usage struct {
	Upload   int64 `json:"upload"`
	Download int64 `json:"download"`
	Total    int64 `json:"total"`
	Expire   int64 `json:"expire,omitempty"`
}

// RuleProvider is a provider of rules.
type RuleProvider struct {
	Name        string    `json:"name"`
	VehicleType string    `json:"vehicleType"`
	Behavior    string    `json:"behavior"`
	Format      string    `json:"format"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Count       int       `json:"count"`
}

// Version is the running core's version.
type Version struct {
	Version string `json:"version"`
	Meta    bool   `json:"meta"`
}
