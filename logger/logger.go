package logger

import (
	"context"
	"log"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var registry atomic.Value

type wrapper struct {
	engine   Logger
	tdrLevel zap.AtomicLevel
	sysLevel zap.AtomicLevel
}

type Logger interface {
	TDR(ctx context.Context, identifier string, fields ...zap.Field)

	Info(ctx context.Context, identifier string, fields ...zap.Field)
	Warn(ctx context.Context, identifier string, fields ...zap.Field)
	Err(ctx context.Context, identifier string, fields ...zap.Field)
	Debug(ctx context.Context, identifier string, fields ...zap.Field)
	ThirdPartyLogger(ctx context.Context, identifier string, fields ...zap.Field)
	SystemFailure(identifier string, fields ...zap.Field)
	SystemInfo(identifier string, fields ...zap.Field)
}

type PublicLoggerFn func(ctx context.Context, identifier string, objects ...zap.Field)
type PublicLoggerWithoutParamsFn func(ctx context.Context, identifier string)

func NewLogger(logLevel zapcore.Level) {
	tdrLog, tdrLevel, err := NewZapLogger(Config{
		LogType:      TDRLog,
		LogLevel:     logLevel,
		SkipCaller:   2,
		EnableFile:   true,
		EnableStdout: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	sysLog, sysLevel, err := NewZapLogger(Config{
		LogType:      SYSLog,
		LogLevel:     logLevel,
		SkipCaller:   2,
		EnableFile:   true,
		EnableStdout: true,
	})
	if err != nil {
		log.Fatalf("sysLog failed to initialize zap logger: %v", err)
	}

	registry.Store(wrapper{
		engine:   &engine{tdr: tdrLog, sys: sysLog},
		tdrLevel: tdrLevel,
		sysLevel: sysLevel,
	})
}

// OverrideLogLevelTo changes the log level for both TDR and SYS loggers at runtime
// without restarting the service.
func OverrideLogLevelTo(level zapcore.Level) {
	w, valid := registry.Load().(wrapper)
	if !valid {
		return
	}
	log.Printf("overriding log level to %s", level.String())
	w.tdrLevel.SetLevel(level)
	w.sysLevel.SetLevel(level)
}

// caller returns the active Logger. If NewLogger has not been called yet,
// it returns a no-op logger rather than crashing the process.
func caller() Logger {
	w, valid := registry.Load().(wrapper)
	if !valid {
		return noopLogger{}
	}
	return w.engine
}

// noopLogger silently drops all log calls. It is returned when the registry
// is not initialized, so library consumers don't crash on early log calls.
type noopLogger struct{}

func (noopLogger) TDR(_ context.Context, _ string, _ ...zap.Field)             {}
func (noopLogger) Info(_ context.Context, _ string, _ ...zap.Field)            {}
func (noopLogger) Warn(_ context.Context, _ string, _ ...zap.Field)            {}
func (noopLogger) Err(_ context.Context, _ string, _ ...zap.Field)             {}
func (noopLogger) Debug(_ context.Context, _ string, _ ...zap.Field)           {}
func (noopLogger) ThirdPartyLogger(_ context.Context, _ string, _ ...zap.Field) {}
func (noopLogger) SystemFailure(_ string, _ ...zap.Field)                       {}
func (noopLogger) SystemInfo(_ string, _ ...zap.Field)                          {}
