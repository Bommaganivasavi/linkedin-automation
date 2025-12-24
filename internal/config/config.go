package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all configuration for the LinkedIn bot
type Config struct {
	Automation AutomationConfig `yaml:"automation"`
	Timing     TimingConfig     `yaml:"timing"`
	Stealth    StealthConfig    `yaml:"stealth"`
	Browser    BrowserConfig    `yaml:"browser"`
}

// AutomationConfig holds settings related to automation limits
type AutomationConfig struct {
	DailyConnectionLimit    int `yaml:"daily_connection_limit"`
	DailyMessageLimit       int `yaml:"daily_message_limit"`
	ConnectionNoteMaxLength int `yaml:"connection_note_max_length"`
}

// TimingConfig holds timing-related settings
type TimingConfig struct {
	MinDelaySeconds   int `yaml:"min_delay_seconds"`
	MaxDelaySeconds   int `yaml:"max_delay_seconds"`
	TypingSpeedMinMs  int `yaml:"typing_speed_min_ms"`
	TypingSpeedMaxMs  int `yaml:"typing_speed_max_ms"`
	PageLoadTimeout   int `yaml:"page_load_timeout"`
	NavigationTimeout int `yaml:"navigation_timeout"`
}

// StealthConfig holds settings to make the bot appear more human-like
type StealthConfig struct {
	RandomScrolling   bool `yaml:"random_scrolling"`
	HumanTyping       bool `yaml:"human_typing"`
	MouseMovement     bool `yaml:"mouse_movement"`
	BusinessHoursOnly bool `yaml:"business_hours_only"`
}

// BrowserConfig holds browser-specific settings
type BrowserConfig struct {
	Headless  bool   `yaml:"headless"`
	UserAgent string `yaml:"user_agent"`
	Viewport  struct {
		Width  int `yaml:"width"`
		Height int `yaml:"height"`
	} `yaml:"viewport"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	cfg := &Config{
		Automation: AutomationConfig{
			DailyConnectionLimit:    30,
			DailyMessageLimit:       50,
			ConnectionNoteMaxLength: 300,
		},
		Timing: TimingConfig{
			MinDelaySeconds:   2,
			MaxDelaySeconds:   5,
			TypingSpeedMinMs:  50,
			TypingSpeedMaxMs:  150,
			PageLoadTimeout:   30,
			NavigationTimeout: 60,
		},
		Stealth: StealthConfig{
			RandomScrolling:   true,
			HumanTyping:       true,
			MouseMovement:     true,
			BusinessHoursOnly: true,
		},
	}

	// Set browser defaults
	cfg.Browser.Headless = false
	cfg.Browser.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	cfg.Browser.Viewport.Width = 1920
	cfg.Browser.Viewport.Height = 1080

	return cfg
}

// LoadConfig loads configuration from a YAML file and merges it with defaults
func LoadConfig(path string) (*Config, error) {
	// Start with default config
	cfg := DefaultConfig()

	// If no config file is specified, return defaults
	if path == "" {
		return cfg, nil
	}

	// Read the config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal the YAML into our config
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate automation settings
	if c.Automation.DailyConnectionLimit <= 0 {
		return fmt.Errorf("daily_connection_limit must be greater than 0")
	}
	if c.Automation.DailyMessageLimit <= 0 {
		return fmt.Errorf("daily_message_limit must be greater than 0")
	}
	if c.Automation.ConnectionNoteMaxLength <= 0 || c.Automation.ConnectionNoteMaxLength > 300 {
		return fmt.Errorf("connection_note_max_length must be between 1 and 300")
	}

	// Validate timing settings
	if c.Timing.MinDelaySeconds < 0 {
		return fmt.Errorf("min_delay_seconds must be 0 or greater")
	}
	if c.Timing.MaxDelaySeconds < c.Timing.MinDelaySeconds {
		return fmt.Errorf("max_delay_seconds must be greater than or equal to min_delay_seconds")
	}
	if c.Timing.TypingSpeedMinMs <= 0 || c.Timing.TypingSpeedMaxMs <= c.Timing.TypingSpeedMinMs {
		return fmt.Errorf("typing_speed_max_ms must be greater than typing_speed_min_ms")
	}
	if c.Timing.PageLoadTimeout <= 0 || c.Timing.NavigationTimeout <= 0 {
		return fmt.Errorf("timeout values must be greater than 0")
	}

	// Validate browser settings
	if c.Browser.Viewport.Width <= 0 || c.Browser.Viewport.Height <= 0 {
		return fmt.Errorf("viewport dimensions must be greater than 0")
	}

	return nil
}

// IsBusinessHours checks if the current time is within business hours
func (c *Config) IsBusinessHours() bool {
	if !c.Stealth.BusinessHoursOnly {
		return true
	}

	now := time.Now()
	hour := now.Hour()
	weekday := now.Weekday()

	// Business hours: Monday to Friday, 9 AM to 5 PM
	return weekday >= time.Monday &&
		weekday <= time.Friday &&
		hour >= 9 && hour < 17
}

// WaitForBusinessHours blocks until business hours if configured
func (c *Config) WaitForBusinessHours() {
	if !c.Stealth.BusinessHoursOnly {
		return
	}

	for !c.IsBusinessHours() {
		nextHour := time.Now().Add(time.Hour).Truncate(time.Hour)
		time.Sleep(time.Until(nextHour))
	}
}
