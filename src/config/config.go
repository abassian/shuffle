package config

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

var (
	// Base
	defaultLogLevel = "debug"
	DefaultDataDir  = defaultHomeDir()
)

// Config contains de configuration for an Shuffle node
type Config struct {

	// Top level options use an anonymous struct
	BaseConfig `mapstructure:",squash"`

	// Options for EVM and State
	Eth *EthConfig `mapstructure:"eth"`

	// Options for Huron consensus
	Huron *HuronConfig `mapstructure:"huron"`

	// Options for Raft consensus
	Raft *RaftConfig `mapstructure:"raft"`
}

// DefaultConfig returns the default configuration for an Shuffle node
func DefaultConfig() *Config {
	return &Config{
		BaseConfig: DefaultBaseConfig(),
		Eth:        DefaultEthConfig(),
		Huron:     DefaultHuronConfig(),
		Raft:       DefaultRaftConfig(),
	}
}

// SetDataDir updates the root data directory as well as the various lower config
// for eth and consensus
func (c *Config) SetDataDir(datadir string) {
	c.BaseConfig.DataDir = datadir
	if c.Eth != nil {
		c.Eth.SetDataDir(fmt.Sprintf("%s/eth", datadir))
	}
	if c.Huron != nil {
		c.Huron.SetDataDir(fmt.Sprintf("%s/huron", datadir))
	}
	if c.Raft != nil {
		c.Raft.SetDataDir(fmt.Sprintf("%s/raft", datadir))
	}
}

/*******************************************************************************
BASE CONFIG
*******************************************************************************/

// BaseConfig contains the top level configuration for an EVM-Huron node
type BaseConfig struct {

	// Top-level directory of evm-huron data
	DataDir string `mapstructure:"datadir"`

	// Debug, info, warn, error, fatal, panic
	LogLevel string `mapstructure:"log"`
}

// DefaultBaseConfig returns the default top-level configuration for EVM-Huron
func DefaultBaseConfig() BaseConfig {
	return BaseConfig{
		DataDir:  DefaultDataDir,
		LogLevel: defaultLogLevel,
	}
}

/*******************************************************************************
FILE HELPERS
*******************************************************************************/

func defaultHomeDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "EVMLITE")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "EVMLITE")
		} else {
			return filepath.Join(home, ".shuffle")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
