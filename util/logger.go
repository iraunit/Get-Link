package util

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("[02-Jan-06 03:04:05 PM]"))
}

func InitLogger() *zap.SugaredLogger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = CustomTimeEncoder
	config.Level = zap.NewAtomicLevelAt(zapcore.Level(0))
	logger, _ := config.Build()
	return logger.Sugar()
}
