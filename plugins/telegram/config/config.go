package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

// Config is the configuration of the program.
type Config struct {
	TelegramBot struct {
		Token    string `yaml:"token" envconfig:"BOT_TOKEN"`
		AdminIDs string `yaml:"admin_ids" envconfig:"BOT_ADMINIDS"`
	} `yaml:"telegram-bot"`
	TBCPusher struct {
		URL string `yaml:"url" envconfig:"TBCPUSHER_URL"`
	} `yaml:"tbcpusher"`
	Callback struct {
		Address string `yaml:"address" envconfig:"CALLBACK_ADDRESS"`
		Prefix  string `yaml:"prefix" envconfig:"CALLBACK_PREFIX"`
		URLBase string `yaml:"urlbase" envconfig:"CALLBACK_URLBASE"`
	} `yaml:"callback"`
}

func New() Config {
	var cfg Config
	cfg.TBCPusher.URL = "http://127.0.0.1:8000/"
	cfg.Callback.Address = ":8001"
	return cfg
}

// ReadFile reads the config file.
func (cfg *Config) ReadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	return decoder.Decode(cfg)
}

// ReadEnv reads the environment variables.
func (cfg *Config) ReadEnv() error {
	return envconfig.Process("", cfg)
}

// WriteFile writes the config file.
func (cfg *Config) WriteFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	return encoder.Encode(cfg)
}

// ReadAll reads the config file and the environment variables.
func (cfg *Config) ReadAll(path string) error {
	e1 := cfg.ReadFile(path)
	e2 := cfg.ReadEnv()
	if e1 != nil && e2 != nil {
		return fmt.Errorf("readfile %v readenv %v", e1, e2)
	} else if e1 != nil {
		return fmt.Errorf("readfile %v", e1)
	} else if e2 != nil {
		return fmt.Errorf("readenv %v", e2)
	}
	return nil
}

// Default path: home directory
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(home, ".config", "tbcpusher-telegram-plugin", "config.yml")
}
