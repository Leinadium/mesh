package logger

import "go.uber.org/zap"

var instance *zap.Logger

func Logger() *zap.Logger {
	if instance == nil {
		instance, _ = zap.NewProduction()
	}
	return instance
}

func Info(msg string, fields ...zap.Field) {
	Logger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Logger().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Logger().Error(msg, fields...)
}
