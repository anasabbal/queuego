package server

import (
	"net"
	"queuego/internal/protocol"
	"sync"
	"time"
)

type Connection struct {
	ID            string
	Conn          net.Conn
	ClientID      string
	Subscriptions map[string]bool
	SendChan      chan *protocol.Command
	Active        bool
	Handler       *Handler
	mu            sync.Mutex
}

func NewConnection(id string, conn net.Conn) *Connection {
	c := &Connection{
		ID:            id,
		Conn:          conn,
		Subscriptions: make(map[string]bool),
		SendChan:      make(chan *protocol.Command, 100),
		Active:        true,
		Handler:       nil,
	}
	go c.reader()
	go c.writer()
	return c
}

func (c *Connection) reader() {
	for c.Active {
		buf := make([]byte, 4096)
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		n, err := c.Conn.Read(buf)
		if err != nil {
			c.Close()
			return
		}

		cmd, err := protocol.Decode(buf[:n])
		if err != nil {
			continue // ignore malformed commands
		}

		if c.Handler != nil {
			c.Handler.HandleCommand(c, cmd)
		}
	}
}

func (c *Connection) writer() {
	for c.Active {
		select {
		case cmd := <-c.SendChan:
			data, err := protocol.Encode(cmd)
			if err != nil {
				continue
			}
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			c.Conn.Write(data)
		}
	}
}

// Send pushes a command to the SendChan
func (c *Connection) Send(cmd *protocol.Command) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Active {
		c.SendChan <- cmd
	}
}

// Close terminates the connection
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.Active {
		c.Active = false
		c.Conn.Close()
		close(c.SendChan)
	}
}

// IsAlive returns true if connection is active
func (c *Connection) IsAlive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Active
}
