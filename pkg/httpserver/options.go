package httpserver

import (
	"net"
	"time"
)

// Option - server option type.
type Option func(*Server)

// Port sets server port.
func Port(port string) Option {
	return func(s *Server) {
		if port == "" {
			port = "8080"
		}
		s.server.Addr = net.JoinHostPort("", port)
	}
}

// ReadTimeout sets read timeout.
func ReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.ReadTimeout = timeout
	}
}

// WriteTimeout sets write timeout.
func WriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.server.WriteTimeout = timeout
	}
}

// ShutdownTimeout sets shutdown timeout.
func ShutdownTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.shutdownTimeout = timeout
	}
}

// Addr sets server address (host:port).
func Addr(addr string) Option {
	return func(s *Server) {
		s.server.Addr = addr
	}
}