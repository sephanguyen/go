package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	defaultLogger      = zap.NewNop()
	defaultSugarLogger = defaultLogger.Sugar()
)

// UseDevelopmentLogger initializes the underlying logger intended for
// local development and CI.
// It panics when failing to build the logger.
func UseDevelopmentLogger(lvl zapcore.Level) {
	lcfg := zap.NewDevelopmentConfig()
	lcfg.Level = zap.NewAtomicLevelAt(lvl)
	lcfg.OutputPaths = []string{"stdout"}
	lcfg.DisableStacktrace = true
	lcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	lcfg.EncoderConfig.EncodeTime = func(time.Time, zapcore.PrimitiveArrayEncoder) {}
	zaplogger, err := lcfg.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build logger: %s", err))
	}
	// add caller skip, otherwise it will always have this pretty.go as the caller
	zaplogger = zaplogger.WithOptions(zap.AddCallerSkip(1))
	defaultLogger = zaplogger
	defaultSugarLogger = defaultLogger.Sugar()
}

// UseDevelopmentLoggerString is similar to UseDevelopmentLogger, but accepts
// a string log level.
func UseDevelopmentLoggerString(lvl string) {
	logLevel, err := zapcore.ParseLevel(lvl)
	if err != nil {
		panic(fmt.Errorf("failed to parse log level: %w", err))
	}
	UseDevelopmentLogger(logLevel)
}

func Infof(msg string, args ...interface{}) {
	defaultSugarLogger.Infof(msg, args...)
}

func Debugf(msg string, args ...interface{}) {
	defaultSugarLogger.Debugf(msg, args...)
}

func Warnf(msg string, args ...interface{}) {
	defaultSugarLogger.Warnf(msg, args...)
}

func Errorf(msg string, args ...interface{}) {
	defaultSugarLogger.Errorf(msg, args...)
}
