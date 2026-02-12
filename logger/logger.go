package logger

import "go.uber.org/zap"

var Logger *zap.Logger

func Init(level string) error {
	var cfg zap.Config

	if level == "debug" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	var err error
	Logger, err = cfg.Build()
	return err
}

// Info 日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Error 日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Debug 日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}
