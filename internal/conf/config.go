package conf

import (
	"bank_test/internal/enum"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var GlobalConfig *Config

// Config holds the configuration values for the API
type Config struct {
	Port       string        `mapstructure:"PORT" validate:"required"`        // Port in which the API will listen
	HealthPort string        `mapstructure:"HEALTH_PORT" validate:"required"` // Health port in which the API will listen
	LogLevel   enum.LogLevel `mapstructure:"LOG_LEVEL" validate:"required"`   // Log level for the API: debug, info
}

// NewConfig returns a new Config instance
func NewConfig() *Config {
	GlobalConfig = new(Config)
	return GlobalConfig
}

// Validate validates that all mandatory fields are correctly set
func (c *Config) Validate() error {
	if !c.LogLevel.IsValid() {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}

	v := validator.New()
	return v.Struct(c)
}
