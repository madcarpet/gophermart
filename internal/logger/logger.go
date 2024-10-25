package logger

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger = zap.NewNop()
var once sync.Once

// initialize logger with given level
func LoggerInit(cfgLevel string) {
	once.Do(func() {
		var level zapcore.Level
		switch cfgLevel {
		case "debug":
			level = zap.DebugLevel
		case "info":
			level = zap.InfoLevel
		case "warn":
			level = zap.WarnLevel
		case "error":
			level = zap.ErrorLevel
		default:
			level = zap.InfoLevel
		}
		encoderCfg := zap.NewProductionEncoderConfig()
		encoderCfg.LevelKey = "lvl"
		encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
		loggerCfg := zap.NewProductionConfig()
		loggerCfg.Level = zap.NewAtomicLevelAt(level)
		loggerCfg.OutputPaths = []string{"stdout"}
		loggerCfg.DisableCaller = true
		loggerCfg.EncoderConfig = encoderCfg
		zl := zap.Must(loggerCfg.Build())
		Log = zl
	})
}
