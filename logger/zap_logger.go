package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	TDRLog = "tdr"
	SYSLog = "sys"

	filePermission = 0755
)

// loggerLevel defines the logger's log level and can be updated at runtime
// without restarting the service.
var loggerLevel zap.AtomicLevel

// Config defines logger configuration.
type Config struct {
	LogType      string
	LogLevel     zapcore.Level
	SkipCaller   int
	EnableFile   bool // Enable/disable file logging
	EnableStdout bool // Enable/disable stdout logging
}

// NewZapLogger creates a new zap.Logger with daily log rotation.
func NewZapLogger(cfg Config) (*zap.Logger, error) {
	loggerLevel = zap.NewAtomicLevelAt(cfg.LogLevel)

	if cfg.LogType != TDRLog && cfg.LogType != SYSLog {
		return nil, fmt.Errorf("invalid log type %q: must be %s or %s", cfg.LogType, TDRLog, SYSLog)
	}

	if !cfg.EnableFile && !cfg.EnableStdout {
		return nil, fmt.Errorf("at least one output (file or stdout) must be enabled")
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout(time.RFC3339),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var syncers []zapcore.WriteSyncer

	if cfg.EnableStdout {
		syncers = append(syncers, zapcore.AddSync(os.Stdout))
	}

	if cfg.EnableFile {
		logDir := filepath.Join("logs", cfg.LogType)
		writer := &RotateFileWriter{
			LogDir:  logDir,
			LogType: cfg.LogType,
		}
		syncers = append(syncers, zapcore.AddSync(writer))
	}

	writeSyncer := zapcore.NewMultiWriteSyncer(syncers...)

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		writeSyncer,
		loggerLevel,
	)

	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(cfg.SkipCaller),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	return logger, nil
}

func OverrideLogLevelTo(level zapcore.Level) {
	log.Printf("overriding log level from %s to %s", loggerLevel, level.String())
	loggerLevel.SetLevel(level)
}
