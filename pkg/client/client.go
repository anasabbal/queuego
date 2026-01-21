package client

import (
	"errors"
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
func (c *Client) SendCommand(cmd *protocol.Command) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.active || c.Conn == nil {
		log.Printf("Cannot send command: connection not active")
		return errors.New("connection not active")
	}
	data, err := protocol.Encode(cmd)
	if err != nil {
		log.Printf("Failed to encode command: %v", err)
		return err
	}
	err = c.Conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return err
	}
	n, err := c.Conn.Write(data)
	if err != nil {
		log.Printf("Failed to send command (%d bytes): %v", n, err)
		return err
	}
	log.Printf("Sent command (%d bytes) to %s", n, c.Address)
	return nil
}
func (c *Client) ReadResponse() (*protocol.Command, error) {
	buf := make([]byte, 4096)
	err := c.Conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	if err != nil {
		return nil, err
	}
	n, err := c.Conn.Read(buf)
	if err != nil {
		log.Printf("Failed to read response: %v", err)
		return nil, err
	}
	cmd, err := protocol.Decode(buf[:n])
	if err != nil {
		log.Printf("Failed to decode response: %v", err)
		return nil, err
	}
	log.Printf("Received response (%d bytes) from %s", n, c.Address)
	return cmd, nil
}
