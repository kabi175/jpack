package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level      string
	Format     string // "json" or "console"
	TimeFormat string
	Caller     bool
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:      "info",
		Format:     "console",
		TimeFormat: time.RFC3339,
		Caller:     true,
	}
}

// Configure sets up the global logger with the given configuration
func Configure(config *LoggerConfig) {
	// Set log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure time format
	zerolog.TimeFieldFormat = config.TimeFormat

	// Configure output format
	if config.Format == "json" {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		// Console format with colors
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: config.TimeFormat,
		}
		log.Logger = zerolog.New(output).With().Timestamp().Logger()
	}

	// Add caller information if requested
	if config.Caller {
		log.Logger = log.Logger.With().Caller().Logger()
	}
}

// GetLogger returns a logger instance for a specific component
func GetLogger(component string) zerolog.Logger {
	return log.Logger.With().Str("component", component).Logger()
}

// GetLoggerWithFields returns a logger instance with additional fields
func GetLoggerWithFields(component string, fields map[string]interface{}) zerolog.Logger {
	logger := log.Logger.With().Str("component", component)

	for key, value := range fields {
		logger = logger.Interface(key, value)
	}

	return logger.Logger()
}

// Initialize sets up the logger with default configuration
func Initialize() {
	Configure(DefaultConfig())
}

// InitializeWithConfig sets up the logger with custom configuration
func InitializeWithConfig(config *LoggerConfig) {
	Configure(config)
}

// SetLevel changes the global log level
func SetLevel(level string) {
	if parsedLevel, err := zerolog.ParseLevel(level); err == nil {
		zerolog.SetGlobalLevel(parsedLevel)
	}
}

// SetDebugLevel sets the log level to debug
func SetDebugLevel() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

// SetInfoLevel sets the log level to info
func SetInfoLevel() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// SetWarnLevel sets the log level to warn
func SetWarnLevel() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
}

// SetErrorLevel sets the log level to error
func SetErrorLevel() {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
}

// Global logger instances for different components
var (
	Schema     = GetLogger("schema")
	Repository = GetLogger("repository")
	Converter  = GetLogger("converter")
	Validation = GetLogger("validation")
	Hooks      = GetLogger("hooks")
	Mongo      = GetLogger("mongo")
	Projection = GetLogger("projection")
	Examples   = GetLogger("examples")
	App        = GetLogger("app") // General application logger
)
