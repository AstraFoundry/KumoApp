// Package pac serves a proxy auto-config script on loopback.
package pac

import (
	"net"
	"net/http"
	"strconv"
	"sync"
)

// Server serves one script at /proxy.pac.
type Server struct {
	mu     sync.Mutex
	script string
	ln     net.Listener
	srv    *http.Server
}

// Serve starts (or updates) the server and returns the script's URL. The
// port stays the same across updates, so the URL the system holds keeps
// working.
func (s *Server) Serve(script string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.script = script
	if s.ln == nil {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return "", err
		}
		s.ln = ln
		s.srv = &http.Server{Handler: http.HandlerFunc(s.handle)}
		go s.srv.Serve(ln)
	}
	return "http://127.0.0.1:" + strconv.Itoa(s.ln.Addr().(*net.TCPAddr).Port) + "/proxy.pac", nil
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	script := s.script
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/x-ns-proxy-autoconfig")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write([]byte(script))
}

// Stop shuts the server down.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.srv != nil {
		s.srv.Close()
		s.srv, s.ln = nil, nil
	}
}
