package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		configData  string
		wantErr     bool
		checkConfig func(t *testing.T, config *Config)
	}{
		{
			name: "valid config",
			configData: `{
				"calendars": {
					"home": {
						"path": "~/.calcli/home",
						"color": "blue",
						"readonly": false
					},
					"work": {
						"path": "/work/calendar",
						"color": "red", 
						"readonly": true
					}
				},
				"defaults": {
					"defaultCalendar": "home"
				}
			}`,
			wantErr: false,
			checkConfig: func(t *testing.T, config *Config) {
				if len(config.Calendars) != 2 {
					t.Errorf("expected 2 calendars, got %d", len(config.Calendars))
				}
				if config.Defaults.DefaultCalendar != "home" {
					t.Errorf("expected default calendar 'home', got %s", config.Defaults.DefaultCalendar)
				}
			},
		},
		{
			name:       "invalid json",
			configData: "invalid json",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary config file
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.json")

			if tt.configData != "" {
				err := os.WriteFile(configPath, []byte(tt.configData), 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			config, err := Load(configPath)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.checkConfig != nil && config != nil {
				tt.checkConfig(t, config)
			}
		})
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	// Test with non-existent config file
	config, err := Load("/nonexistent/config.json")

	if err != nil {
		t.Errorf("expected no error for missing config, got %v", err)
	}

	// Should return default config
	if len(config.Calendars) == 0 {
		t.Error("expected default config to have calendars")
	}

	if config.Defaults.DefaultCalendar != "home" {
		t.Errorf("expected default calendar 'home', got %s", config.Defaults.DefaultCalendar)
	}
}

func TestGetDefaultCalendar(t *testing.T) {
	config := defaultConfig()

	calendar, err := config.GetDefaultCalendar()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if calendar.Name != "home" {
		t.Errorf("expected calendar name 'home', got %s", calendar.Name)
	}

	if calendar.ReadOnly {
		t.Error("expected home calendar to be writable")
	}
}
