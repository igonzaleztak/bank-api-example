package enum

// LogLevel is a type for log levels
type LogLevel string

// Log levels
const (
	Debug     LogLevel = "debug"
	InfoLevel LogLevel = "info"
)

// String returns the string representation of the log level
func (e LogLevel) String() string {
	return string(e)
}

// IsValid checks if the log level is valid
func (e LogLevel) IsValid() bool {
	switch e {
	case Debug, InfoLevel:
		return true
	default:
		return false
	}
}
