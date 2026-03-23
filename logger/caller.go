package logger

import (
	"context"

	"go.uber.org/zap"
)

func TDR(ctx context.Context, identifier string, fields ...zap.Field) {
	caller().TDR(ctx, identifier, fields...)
}

func Info(ctx context.Context, identifier string, fields ...zap.Field) {
	caller().Info(ctx, identifier, fields...)
}

func Err(ctx context.Context, identifier string, fields ...zap.Field) {
	caller().Err(ctx, identifier, fields...)
}

func Warn(ctx context.Context, identifier string, fields ...zap.Field) {
	caller().Warn(ctx, identifier, fields...)
}

func Debug(ctx context.Context, identifier string, fields ...zap.Field) {
	caller().Debug(ctx, identifier, fields...)
}

func SystemFailure(identifier string, fields ...zap.Field) {
	caller().SystemFailure(identifier, fields...)
}

func SystemInfo(identifier string, fields ...zap.Field) {
	caller().SystemInfo(identifier, fields...)
}

func ThirdPartyLogger(ctx context.Context, identifier string, fields ...zap.Field) {
	caller().ThirdPartyLogger(ctx, identifier, fields...)
}
