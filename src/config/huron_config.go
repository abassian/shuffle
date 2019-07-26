package config

import (
	"fmt"
	"time"

	_huron "github.com/abassian/huron/src/huron"
)

var (
	defaultNodeAddr       = "127.0.0.1:1337"
	defaultHuronAPIAddr  = ":8000"
	defaultHeartbeat      = 500 * time.Millisecond
	defaultTCPTimeout     = 1000 * time.Millisecond
	defaultCacheSize      = 50000
	defaultSyncLimit      = 1000
	defaultEnableFastSync = false
	defaultMaxPool        = 2
	defaultHuronDir      = fmt.Sprintf("%s/huron", DefaultDataDir)
	defaultPeersFile      = fmt.Sprintf("%s/peers.json", defaultHuronDir)
)

// HuronConfig contains the configuration of a Huron node
type HuronConfig struct {

	// Directory containing priv_key.pem and peers.json files
	DataDir string `mapstructure:"datadir"`

	// Address of Huron node (where it talks to other Huron nodes)
	BindAddr string `mapstructure:"listen"`

	// Huron HTTP API address
	ServiceAddr string `mapstructure:"service-listen"`

	// Gossip heartbeat
	Heartbeat time.Duration `mapstructure:"heartbeat"`

	// TCP timeout
	TCPTimeout time.Duration `mapstructure:"timeout"`

	// Max number of items in caches
	CacheSize int `mapstructure:"cache-size"`

	// Max number of Event in SyncResponse
	SyncLimit int `mapstructure:"sync-limit"`

	// Allow node to FastSync
	EnableFastSync bool `mapstructure:"fast-sync"`

	// Max number of connections in net pool
	MaxPool int `mapstructure:"max-pool"`

	// Database type; badger or inmeum
	Store bool `mapstructure:"store"`

	// Bootstrap from database
	Bootstrap bool `mapstructure:"bootstrap"`
}

// DefaultHuronConfig returns the default configuration for a Huron node
func DefaultHuronConfig() *HuronConfig {
	return &HuronConfig{
		DataDir:        defaultHuronDir,
		BindAddr:       defaultNodeAddr,
		ServiceAddr:    defaultHuronAPIAddr,
		Heartbeat:      defaultHeartbeat,
		TCPTimeout:     defaultTCPTimeout,
		CacheSize:      defaultCacheSize,
		SyncLimit:      defaultSyncLimit,
		EnableFastSync: defaultEnableFastSync,
		MaxPool:        defaultMaxPool,
	}
}

// SetDataDir updates the huron configuration directories if they were set to
// to default values.
func (c *HuronConfig) SetDataDir(datadir string) {
	if c.DataDir == defaultHuronDir {
		c.DataDir = datadir
	}
}

// ToRealHuronConfig converts an shuffle/src/config.HuronConfig to a
// huron/src/huron.HuronConfig as used by Huron
func (c *HuronConfig) ToRealHuronConfig() *_huron.HuronConfig {
	huronConfig := _huron.NewDefaultConfig()
	huronConfig.DataDir = c.DataDir
	huronConfig.BindAddr = c.BindAddr
	huronConfig.ServiceAddr = c.ServiceAddr
	huronConfig.MaxPool = c.MaxPool
	huronConfig.Store = c.Store
	huronConfig.NodeConfig.HeartbeatTimeout = c.Heartbeat
	huronConfig.NodeConfig.TCPTimeout = c.TCPTimeout
	huronConfig.NodeConfig.CacheSize = c.CacheSize
	huronConfig.NodeConfig.SyncLimit = c.SyncLimit
	huronConfig.NodeConfig.EnableFastSync = c.EnableFastSync
	huronConfig.NodeConfig.Bootstrap = c.Bootstrap
	return huronConfig
}
