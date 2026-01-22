package server

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
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
	for c.IsAlive() {
		lenBuf := make([]byte, 4)
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		if _, err := io.ReadFull(c.Conn, lenBuf); err != nil {
			// Treat EOF/unexpected EOF as normal client close (avoid noisy log)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Printf("[%s] client closed connection", c.ID)
			} else {
				log.Printf("[%s] read length error: %v", c.ID, err)
			}
			c.Close()
			return
		}
		msgLen := binary.BigEndian.Uint32(lenBuf)

		if msgLen > 10*1024*1024 { // 10 MB sanity check
			log.Printf("[%s] message too large: %d bytes", c.ID, msgLen)
			c.Close()
			return
		}

		data := make([]byte, msgLen)
		if _, err := io.ReadFull(c.Conn, data); err != nil {
			// Treat EOF/unexpected EOF as normal client close
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				log.Printf("[%s] client closed connection while reading payload", c.ID)
			} else {
				log.Printf("[%s] read message error: %v", c.ID, err)
			}
			c.Close()
			return
		}

		cmd, err := protocol.Decode(data)
		if err != nil {
			log.Printf("[%s] decode error: %v", c.ID, err)
			continue
		}

		if c.Handler != nil {
			c.Handler.HandleCommand(c, cmd)
		}
	}
}
func (c *Connection) writer() {
	for c.IsAlive() {
		select {
		case cmd, ok := <-c.SendChan:
			if !ok || cmd == nil {
				continue
			}

			data, err := protocol.Encode(cmd)
			if err != nil {
				log.Printf("[%s] encode failed: %v", c.ID, err)
				continue
			}

			var buf bytes.Buffer
			binary.Write(&buf, binary.BigEndian, uint32(len(data)))
			buf.Write(data)

			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			n, err := c.Conn.Write(buf.Bytes())
			if err != nil {
				log.Printf("[%s] write failed (%d bytes): %v", c.ID, n, err)
				c.Close()
				return
			}

			log.Printf("[%s] sent command %s (%d bytes)", c.ID, cmd.Type, n)
		}
	}
}

// Send pushes a command to SendChan safely
func (c *Connection) Send(cmd *protocol.Command) {
	if cmd == nil {
		log.Printf("[%s] attempted to send nil command, skipping", c.ID)
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.Active {
		log.Printf("[%s] cannot send, connection inactive", c.ID)
		return
	}

	select {
	case c.SendChan <- cmd:
		log.Printf("[%s] queued command %s for sending", c.ID, cmd.Type)
	default:
		log.Printf("[%s] send channel full, dropping command %s", c.ID, cmd.Type)
	}
}

// Close safely closes the connection
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Active {
		c.Active = false
		c.Conn.Close()
		close(c.SendChan)
		log.Printf("[%s] connection closed", c.ID)
	}
}

// IsAlive returns true if connection is active
func (c *Connection) IsAlive() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.Active
}
