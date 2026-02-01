package logger

import (
	"context"
	"log"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	registry atomic.Value
)

type wrapper struct {
	engine Logger
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
	tdrLog, err := NewZapLogger(Config{
		LogType:      TDRLog,
		LogLevel:     logLevel,
		SkipCaller:   2,
		EnableFile:   true,
		EnableStdout: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	sysLog, err := NewZapLogger(Config{
		LogType:      SYSLog,
		LogLevel:     logLevel,
		SkipCaller:   2,
		EnableFile:   true,
		EnableStdout: true,
	})
	if err != nil {
		log.Fatalf("sysLog failed to initialize zap logger TDR: %v", err)
	}

	registry.Store(wrapper{&engine{
		tdr: tdrLog,
		sys: sysLog,
	}})
}

func caller() Logger {
	wrapper, valid := registry.Load().(wrapper)
	if !valid {
		log.Fatal("invalid log")
	}

	return wrapper.engine
}
