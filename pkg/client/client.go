package client

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"io"
	"log"
	"net"
	"queuego/internal/protocol"
	"sync"
	"time"
)

type ClientConfig struct {
	RetryMax      int
	RetryInterval time.Duration
	ConnTimeout   time.Duration
}

type Client struct {
	Address     string
	Conn        net.Conn
	Config      ClientConfig
	mu          sync.Mutex
	subscribers map[string]func(*protocol.Command) // topic -> handler
	active      bool
}

func (c *Client) Connect(address string) error {
	c.Address = address
	var err error

	for attempt := 0; attempt < c.Config.RetryMax; attempt++ {
		log.Printf("Attempting to connect to %s (try %d/%d)", address, attempt+1, c.Config.RetryMax)
		c.Conn, err = net.DialTimeout("tcp", address, c.Config.ConnTimeout)
		if err == nil {
			c.active = true
			log.Printf("Successfully connected to %s", address)
			return nil
		}
		log.Printf("Connection failed: %v. Retrying in %s...", err, c.Config.RetryInterval*(1<<attempt))
		time.Sleep(c.Config.RetryInterval * (1 << attempt))
	}
	log.Printf("Failed to connect to %s after %d attempts", address, c.Config.RetryMax)
	return err
}

func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Conn != nil {
		c.active = false
		err := c.Conn.Close()
		if err != nil {
			log.Printf("Error disconnecting from %s: %v", c.Address, err)
		} else {
			log.Printf("Disconnected from %s", c.Address)
		}
		return err
	}
	log.Printf("No active connection to disconnect from %s", c.Address)
	return nil
}
func EncodeWithLength(cmd *protocol.Command) ([]byte, error) {
	data, err := protocol.Encode(cmd)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(data))); err != nil {
		return nil, err
	}
	buf.Write(data)
	return buf.Bytes(), nil
}

func (c *Client) SendCommand(cmd *protocol.Command) error {
	data, err := protocol.Encode(cmd)
	if err != nil {
		return err
	}

	// prepend length
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}
	buf.Write(data)

	c.mu.Lock()
	defer c.mu.Unlock()

	c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, err = c.Conn.Write(buf.Bytes())
	if err != nil {
		return err
	}
	log.Printf("Sent command (%d bytes) to %s", len(buf.Bytes()), c.Address)
	return nil
}
func (c *Client) ReadResponse() (*protocol.Command, error) {
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(c.Conn, lenBuf); err != nil {
		// show reason for read length failure
		log.Printf("ReadResponse: failed to read length prefix: %v", err)
		return nil, err
	}
	msgLen := binary.BigEndian.Uint32(lenBuf)

	data := make([]byte, msgLen)
	if _, err := io.ReadFull(c.Conn, data); err != nil {
		// log both the length we expected and the underlying error
		log.Printf("ReadResponse: failed to read %d bytes of payload: %v", msgLen, err)
		return nil, err
	}

	cmd, err := protocol.Decode(data)
	if err != nil {
		// debug: dump raw bytes received
		log.Printf("ReadResponse: Failed to decode response: %v", err)
		log.Printf("ReadResponse: raw length prefix: %s", hex.EncodeToString(lenBuf))
		log.Printf("ReadResponse: raw payload (%d bytes): %s", len(data), hex.EncodeToString(data))
		return nil, err
	}
	return cmd, nil
}
