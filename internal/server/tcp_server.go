package server

import (
	"log"
	"net"
	"queuego/internal/broker"
	"sync"
)

type Server struct {
	Listener    net.Listener
	Broker      *broker.Broker
	Handler     *Handler
	Connections map[string]*Connection
	mu          sync.Mutex
}

func NewServer(addr string, broker *broker.Broker) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := &Server{
		Listener:    ln,
		Broker:      broker,
		Handler:     &Handler{Broker: broker},
		Connections: make(map[string]*Connection),
	}
	return s, nil
}

// Start accepts incoming connections
func (s *Server) Start() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			continue
		}
		client := NewConnection(conn.RemoteAddr().String(), conn)
		client.Handler = s.Handler
		log.Printf("New client connected: %s", client.ID)

		s.mu.Lock()
		s.Connections[client.ID] = client
		s.mu.Unlock()
	}
}

// Stop closes all connections and listener
func (s *Server) Stop() {
	s.Listener.Close()
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.Connections {
		c.Close()
	}
}
