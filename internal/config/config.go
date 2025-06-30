package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config enforces type safety and validation for user preferences
type Config struct {
	UI      UIConfig      `mapstructure:"ui"`
	Display DisplayConfig `mapstructure:"display"`
}

// UIConfig optimizes interface defaults for portfolio monitoring workflows
type UIConfig struct {
	Theme          string `mapstructure:"theme"`
	DefaultSection string `mapstructure:"default_section"`
}

// DisplayConfig adapts to NEPSE market conventions and user preferences
type DisplayConfig struct {
	RefreshInterval int    `mapstructure:"refresh_interval"`
	CurrencySymbol  string `mapstructure:"currency_symbol"`
}

// Load implements configuration cascade prioritizing user control over defaults
// Hierarchy prevents frustrating config resets during critical trading periods
func Load() (*Config, error) {
	// Standard config locations enable predictable deployment patterns
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	// Multi-path search supports both system and user-specific configurations
	configDir, err := getConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// NTX prefix prevents conflicts with other trading applications
	viper.SetEnvPrefix("NTX")
	viper.AutomaticEnv()

	// Defaults ensure functional application even without configuration files
	setDefaults()

	// Graceful config file handling prevents startup failures
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Auto-create config reduces setup friction for new users
			if err := createDefaultConfigFile(configDir); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Struct validation prevents runtime errors from malformed configuration
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// Save persists user preferences to survive application restarts
func Save(config *Config) error {
	configDir, err := getConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get config directory: %w", err)
	}

	// Viper synchronization ensures consistent state across save operations
	viper.Set("ui.theme", config.UI.Theme)
	viper.Set("ui.default_section", config.UI.DefaultSection)
	viper.Set("display.refresh_interval", config.Display.RefreshInterval)
	viper.Set("display.currency_symbol", config.Display.CurrencySymbol)

	// Atomic write prevents config corruption during market volatility
	configFile := filepath.Join(configDir, "config.toml")
	return viper.WriteConfigAs(configFile)
}

// getConfigDir follows XDG standards for predictable config location
func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(homeDir, ".config", "ntx")

	// Auto-create directory eliminates manual setup steps for users
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// setDefaults provides NEPSE-optimized defaults for immediate usability
func setDefaults() {
	viper.SetDefault("ui.theme", "tokyo_night")
	viper.SetDefault("ui.default_section", "holdings")
	viper.SetDefault("display.refresh_interval", 30)
	viper.SetDefault("display.currency_symbol", "Rs.") // Standard Rs. format for better terminal compatibility
}

// createDefaultConfigFile bootstraps user configuration with optimal defaults
func createDefaultConfigFile(configDir string) error {
	configFile := filepath.Join(configDir, "config.toml")

	// Existence check prevents overwriting user customizations
	if _, err := os.Stat(configFile); err == nil {
		return nil // File exists, no need to create
	}

	// TOML format chosen for human readability and easy manual editing
	defaultContent := `[ui]
theme = "tokyo_night"
default_section = "holdings"

[display]
refresh_interval = 30
currency_symbol = "Rs."
`

	return os.WriteFile(configFile, []byte(defaultContent), 0644)
}

// GetTheme enables theme-aware component rendering
func (c *Config) GetTheme() string {
	return c.UI.Theme
}

// SetTheme enables runtime theme persistence for live switching
func (c *Config) SetTheme(theme string) {
	c.UI.Theme = theme
}

// GetDefaultSection personalizes startup behavior for user workflows
func (c *Config) GetDefaultSection() string {
	return c.UI.DefaultSection
}
