package backend

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	gormlogger "gorm.io/gorm/logger"
)

var (
	Log      *zap.SugaredLogger
	logLevel zap.AtomicLevel
	devMode  bool
)

func init() {
	logLevel = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	Log = buildLogger().Sugar()
}

func SetMode(mode string) {
	devMode = mode == "dev"
	Log = buildLogger().Sugar()
	if db != nil {
		db.Logger = newGORMLogger()
	}
}

func buildLogger() *zap.Logger {
	if devMode {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = logLevel
		l, _ := cfg.Build()
		return l
	}
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.DateTime)
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.Level = logLevel
	l, _ := cfg.Build()
	return l
}

func zapLevelFromString(level string) zapcore.Level {
	switch level {
	case "silent":
		return zapcore.Level(127)
	case "panic":
		return zapcore.PanicLevel
	case "error":
		return zapcore.ErrorLevel
	case "info":
		return zapcore.InfoLevel
	case "debug":
		return zapcore.DebugLevel
	default:
		return zapcore.WarnLevel
	}
}

func SetLogLevel(level string) {
	logLevel.SetLevel(zapLevelFromString(level))
	if db != nil {
		db.Logger = newGORMLogger()
	}
}

// zapGORMWriter routes GORM query logs to zap at debug level.
type zapGORMWriter struct{}

func (zapGORMWriter) Printf(format string, args ...interface{}) {
	Log.Debugf(fmt.Sprintf(format, args...))
}

func newGORMLogger() gormlogger.Interface {
	return gormlogger.New(
		zapGORMWriter{},
		gormlogger.Config{
			SlowThreshold:             100 * time.Millisecond,
			LogLevel:                  gormlogger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  devMode,
		},
	)
}
