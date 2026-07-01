package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		checkFn  func(*testing.T, *Config)
	}{
		{
			name: "loads defaults when no env vars set",
			env:  map[string]string{},
			checkFn: func(t *testing.T, c *Config) {
				if c.Port != "8080" {
					t.Errorf("Port = %q; want %q", c.Port, "8080")
				}
				if c.MaxRequestsPerMonth != 100 {
					t.Errorf("MaxRequestsPerMonth = %d; want %d", c.MaxRequestsPerMonth, 100)
				}
				if c.ProjectID != "" {
					t.Errorf("ProjectID = %q; want empty", c.ProjectID)
				}
			},
		},
		{
			name: "loads from environment",
			env: map[string]string{
				"PORT":              "9090",
				"GCP_PROJECT_ID":    "my-project",
				"SKYSCANNER_MAX_REQ": "50",
				"LOG_LEVEL":         "debug",
			},
			checkFn: func(t *testing.T, c *Config) {
				if c.Port != "9090" {
					t.Errorf("Port = %q; want %q", c.Port, "9090")
				}
				if c.ProjectID != "my-project" {
					t.Errorf("ProjectID = %q; want %q", c.ProjectID, "my-project")
				}
				if c.MaxRequestsPerMonth != 50 {
					t.Errorf("MaxRequestsPerMonth = %d; want %d", c.MaxRequestsPerMonth, 50)
				}
				if c.LogLevel != "debug" {
					t.Errorf("LogLevel = %q; want %q", c.LogLevel, "debug")
				}
			},
		},
		{
			name: "ignores negative max requests",
			env: map[string]string{
				"SKYSCANNER_MAX_REQ": "-1",
			},
			checkFn: func(t *testing.T, c *Config) {
				if c.MaxRequestsPerMonth != 100 {
					t.Errorf("MaxRequestsPerMonth = %d; want %d", c.MaxRequestsPerMonth, 100)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				os.Setenv(k, v)
			}
			t.Cleanup(func() {
				for k := range tt.env {
					os.Unsetenv(k)
				}
			})

			cfg := Load()
			tt.checkFn(t, cfg)
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Port:                "8080",
				ProjectID:           "my-project",
				SkyscannerAPIKey:    "key-123",
				MaxRequestsPerMonth: 100,
			},
			wantErr: false,
		},
		{
			name: "missing project id",
			cfg: &Config{
				Port:                "8080",
				MaxRequestsPerMonth: 100,
			},
			wantErr: true,
		},
		{
			name: "missing api key",
			cfg: &Config{
				Port:                "8080",
				ProjectID:           "my-project",
				MaxRequestsPerMonth: 100,
			},
			wantErr: true,
		},
		{
			name: "missing port",
			cfg: &Config{
				ProjectID:           "my-project",
				SkyscannerAPIKey:    "key-123",
				MaxRequestsPerMonth: 100,
			},
			wantErr: true,
		},
		{
			name: "zero max requests",
			cfg: &Config{
				Port:                "8080",
				ProjectID:           "my-project",
				SkyscannerAPIKey:    "key-123",
				MaxRequestsPerMonth: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
