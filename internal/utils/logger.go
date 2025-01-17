package utils

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitLogger() {
	Logger = zap.Must(zap.NewProduction())
}
