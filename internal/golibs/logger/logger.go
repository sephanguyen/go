package logger

import (
	"fmt"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger returns a zap logger whose log level is fixed (DEBUG by default)
func NewZapLogger(logLevel string, isLocalEnv bool, opts ...zap.Option) *zap.Logger {
	var (
		zapLogger *zap.Logger
		zapLogLvl zapcore.Level
	)

	err := zapLogLvl.Set(logLevel)
	if err != nil {
		log.Println("cannot parse logLevel, err:", err.Error())
		zapLogLvl = zap.DebugLevel
	}

	if isLocalEnv {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.Level = zap.NewAtomicLevelAt(zapLogLvl)
		zapLogger, err = config.Build(opts...)
		if err != nil {
			log.Println("cannot build logger, err:", err.Error())
		}
		return zapLogger
	}

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapLogLvl && lvl < zapcore.ErrorLevel
	})
	consoleInfos := zapcore.Lock(os.Stdout)
	consoleErrors := zapcore.Lock(os.Stderr)

	// Configure console output.
	consoleEncoder := newJSONEncoder()
	// Join the outputs, encoders, and level-handling functions into
	// zapcore.
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleErrors, highPriority),
		zapcore.NewCore(consoleEncoder, consoleInfos, lowPriority),
	)

	// From a zapcore.Core, it's easy to construct a Logger.
	opts2 := []zap.Option{zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)}
	opts2 = append(opts2, opts...)
	zapLogger = zap.New(core, opts2...)
	zap.RedirectStdLog(zapLogger)

	return zapLogger
}

// Create a new JSON log encoder with the correct settings.
func newJSONEncoder() zapcore.Encoder {
	return zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "severity",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    appendLogLevel,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})
}

func appendLogLevel(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString("debug")
	case zapcore.InfoLevel:
		enc.AppendString("info")
	case zapcore.WarnLevel:
		enc.AppendString("warning")
	case zapcore.ErrorLevel:
		enc.AppendString("error")
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		enc.AppendString("critical")
	default:
		enc.AppendString(fmt.Sprintf("Level(%d)", l))
	}
}
