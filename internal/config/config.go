package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/NaMinhyeok/calcli/internal/domain"
)

type Config struct {
	Calendars map[string]CalendarConfig `json:"calendars"`
	Defaults  DefaultsConfig            `json:"defaults"`
}

type CalendarConfig struct {
	Path     string `json:"path"`
	Color    string `json:"color"`
	ReadOnly bool   `json:"readonly"`
}

type DefaultsConfig struct {
	DefaultCalendar string `json:"defaultCalendar"`
}

func Load(configPath string) (*Config, error) {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func GetDefaultConfigPath() string {
	if v := os.Getenv("CALCLI_CONFIG"); v != "" {
		// Allow tests or users to override the config path via environment.
		if len(v) >= 1 && v[0] == '~' {
			return expandPath(v)
		}
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		// Fallback to HOME if UserHomeDir fails
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, ".calcli", "config.json")
}

func (c *Config) GetCalendar(name string) (CalendarConfig, bool) {
	cal, exists := c.Calendars[name]
	return cal, exists
}

func (c *Config) GetDefaultCalendar() (domain.Calendar, error) {
	defaultName := c.Defaults.DefaultCalendar
	if defaultName == "" {
		defaultName = "home"
	}

	calConfig, exists := c.GetCalendar(defaultName)
	if !exists {
		return domain.Calendar{}, os.ErrNotExist
	}

	return domain.Calendar{
		Name:     defaultName,
		Path:     expandPath(calConfig.Path),
		Color:    calConfig.Color,
		ReadOnly: calConfig.ReadOnly,
	}, nil
}

// defaultConfig returns a sensible default configuration
func defaultConfig() *Config {
	home := os.Getenv("HOME")
	return &Config{
		Calendars: map[string]CalendarConfig{
			"home": {
				Path:     filepath.Join(home, ".calcli", "home"),
				Color:    "blue",
				ReadOnly: false,
			},
		},
		Defaults: DefaultsConfig{
			DefaultCalendar: "home",
		},
	}
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if len(path) >= 2 && path[0] == '~' && path[1] == '/' {
		home := os.Getenv("HOME")
		return filepath.Join(home, path[2:])
	}
	return path
}
