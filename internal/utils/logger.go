package utils

import (
	"go.uber.org/zap"
)

/*
Global logger instance: we use Zap (production config) for structured logging.
Can be replaced or reconfigured with different settings.
*/
var Logger *zap.Logger

/*
InitLogger sets up the global Logger with production defaults.
You can switch to zap.NewDevelopment() for more human-readable output or advanced config.
*/
func InitLogger() {
	// Must() will panic if creation fails, effectively stopping the program
	Logger = zap.Must(zap.NewProduction())
}
