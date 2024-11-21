package conf

import (
	"fmt"

	"github.com/spf13/viper"
)

// setupConfig is a function that sets up the configuration for the application by reading from environment variables
// and validating the configuration. The configuration is stored in the conf package.
func SetupConfig() error {
	setDefaults()

	// Read from environment variables
	viper.AutomaticEnv()

	cfg := NewConfig()
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("bootstrap: config: failed to unmarshal configuration: %v", err)
	}

	return cfg.Validate()
}

// setDefaults is a function that sets the default values for the API configuration.
func setDefaults() {
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("HEALTH_PORT", "8081")
	viper.SetDefault("PORT", "8080")
}
