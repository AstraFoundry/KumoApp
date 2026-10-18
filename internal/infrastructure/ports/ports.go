// Package ports finds out whether the ports Kumo needs are free, and who
// holds them when they are not.
package ports

import (
	"net"
	"strconv"
)

// Holder is the process listening on a port.
type Holder struct {
	PID     int    `json:"pid"`
	Process string `json:"process"`
}

// Free reports whether host:port can be listened on.
func Free(host string, port int) bool {
	ln, err := net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// Accepting reports whether something answers TCP connections on host:port.
func Accepting(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, strconv.Itoa(port)), 500_000_000)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Suggest returns the first free port after port, skipping any in avoid.
func Suggest(host string, port int, avoid ...int) int {
	for candidate := port + 1; candidate < port+200 && candidate <= 65535; candidate++ {
		skip := false
		for _, a := range avoid {
			skip = skip || a == candidate
		}
		if !skip && Free(host, candidate) {
			return candidate
		}
	}
	ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
	if err != nil {
		return 0
	}
	defer ln.Close()
	return ln.Addr().(*net.TCPAddr).Port
}
