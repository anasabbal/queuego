package config

import (
	"errors"
	"gopkg.in/yaml.v3"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Broker  BrokerConfig  `yaml:"broker"`
	Network NetworkConfig `yaml:"network"`
	Storage StorageConfig `yaml:"storage"`
}
type ServerConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	MaxConnections int    `yaml:"maxConnections"`
}

type BrokerConfig struct {
	MaxTopics        int           `yaml:"maxTopics"`
	DefaultQueueSize int           `yaml:"defaultQueueSize"`
	MessageTTL       time.Duration `yaml:"messageTTL"`
}

type NetworkConfig struct {
	ReadTimeout       time.Duration `yaml:"readTimeout"`
	WriteTimeout      time.Duration `yaml:"writeTimeout"`
	HeartbeatInterval time.Duration `yaml:"heartbeatInterval"`
}

type StorageConfig struct {
	Type              string        `yaml:"type"`
	MaxSize           int           `yaml:"maxSize"`
	RetentionDuration time.Duration `yaml:"retentionDuration"`
}

func New() *Config {
	return &Config{
		Server: ServerConfig{
			Host:           "0.0.0.0",
			Port:           9092,
			MaxConnections: 1000,
		},
		Broker: BrokerConfig{
			MaxTopics:        1000,
			DefaultQueueSize: 1000,
			MessageTTL:       time.Hour,
		},
		Network: NetworkConfig{
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			HeartbeatInterval: 10 * time.Second,
		},
		Storage: StorageConfig{
			Type:              "memory",
			MaxSize:           100_000,
			RetentionDuration: 24 * time.Hour,
		},
	}
}
func (c *Config) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, c)
}
func (c *Config) LoadFromEnv() {
	if v := os.Getenv("QUEUEGO_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Server.Port = port
		}
	}

	if v := os.Getenv("QUEUEGO_MAX_CONNECTIONS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Server.MaxConnections = n
		}
	}

	if v := os.Getenv("QUEUEGO_MESSAGE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.Broker.MessageTTL = d
		}
	}

	if v := os.Getenv("QUEUEGO_STORAGE_TYPE"); v != "" {
		c.Storage.Type = v
	}
}
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return errors.New("invalid server port")
	}
	if c.Server.MaxConnections <= 0 {
		return errors.New("maxConnections must be > 0")
	}
	if c.Broker.DefaultQueueSize <= 0 {
		return errors.New("defaultQueueSize must be > 0")
	}
	if c.Storage.Type != "memory" && c.Storage.Type != "file" {
		return errors.New("storage.type must be memory or file")
	}
	return nil
}
