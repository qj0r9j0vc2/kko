package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	RedirectURI     string            `mapstructure:"redirect_uri"`
	DefaultLocation Location          `mapstructure:"default_location"`
	Output          OutputConfig      `mapstructure:"output"`
	Aliases         map[string]string `mapstructure:"aliases"`
	Search          SearchConfig      `mapstructure:"search"`
}

type Location struct {
	Name string  `mapstructure:"name"`
	Lat  float64 `mapstructure:"lat"`
	Lng  float64 `mapstructure:"lng"`
}

type OutputConfig struct {
	Format string `mapstructure:"format"`
	Color  bool   `mapstructure:"color"`
	Lang   string `mapstructure:"lang"`
}

type SearchConfig struct {
	DefaultEngine string `mapstructure:"default_engine"`
	MaxResults    int    `mapstructure:"max_results"`
}

func DefaultConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "kko")
}

func DefaultConfigPath() string {
	return filepath.Join(DefaultConfigDir(), "config.yaml")
}

func Load(path string) (*Config, error) {
	if path != "" {
		viper.SetConfigFile(path)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(DefaultConfigDir())
	}

	viper.SetDefault("redirect_uri", "http://localhost:9876/callback")
	viper.SetDefault("output.format", "table")
	viper.SetDefault("output.color", true)
	viper.SetDefault("output.lang", "ko")
	viper.SetDefault("search.default_engine", "web")
	viper.SetDefault("search.max_results", 5)
	viper.SetDefault("aliases", map[string]string{})

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}
	return &cfg, nil
}

func Save() error {
	dir := DefaultConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return viper.WriteConfigAs(DefaultConfigPath())
}

func Set(key, value string) error {
	viper.Set(key, value)
	return Save()
}

func Get(key string) interface{} {
	return viper.Get(key)
}
