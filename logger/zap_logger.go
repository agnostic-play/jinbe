package logger

import (
	"fmt"
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

// Config defines logger configuration.
type Config struct {
	LogType      string
	LogLevel     zapcore.Level
	SkipCaller   int
	EnableFile   bool
	EnableStdout bool
}

// NewZapLogger creates a new zap.Logger with daily log rotation.
// It returns both the logger and its AtomicLevel so the caller can adjust the level at runtime.
func NewZapLogger(cfg Config) (*zap.Logger, zap.AtomicLevel, error) {
	level := zap.NewAtomicLevelAt(cfg.LogLevel)

	if cfg.LogType != TDRLog && cfg.LogType != SYSLog {
		return nil, level, fmt.Errorf("invalid log type %q: must be %s or %s", cfg.LogType, TDRLog, SYSLog)
	}

	if !cfg.EnableFile && !cfg.EnableStdout {
		return nil, level, fmt.Errorf("at least one output (file or stdout) must be enabled")
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
		level,
	)

	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(cfg.SkipCaller),
		zap.AddStacktrace(zap.ErrorLevel),
	)

	return logger, level, nil
}
