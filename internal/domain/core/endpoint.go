package core

import (
	"fmt"
	"net"
	"strconv"
)

// Endpoint is where Mihomo's external controller listens.
type Endpoint struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Secret string `json:"secret"`
}

// DefaultEndpoint is the controller address Kumo uses unless told otherwise.
func DefaultEndpoint() Endpoint { return Endpoint{Host: "127.0.0.1", Port: 9097} }

// Address is host:port.
func (e Endpoint) Address() string { return net.JoinHostPort(e.Host, strconv.Itoa(e.Port)) }

// BaseURL is the controller's HTTP root.
func (e Endpoint) BaseURL() string { return "http://" + e.Address() }

// Validate rejects an endpoint the core cannot listen on.
func (e Endpoint) Validate() error {
	if e.Host == "" {
		return fmt.Errorf("controller host is empty")
	}
	return ValidatePort(e.Port)
}

// ValidatePort rejects ports outside 1–65535.
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port %d is outside 1–65535", port)
	}
	return nil
}
