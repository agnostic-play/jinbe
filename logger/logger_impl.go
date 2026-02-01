package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type engine struct {
	sys *zap.Logger
	tdr *zap.Logger
}

func (z engine) TDR(ctx context.Context, identifier string, fields ...zap.Field) {
	loggerCtx := GetLoggerCtx(ctx)
	fields = append(fields, zap.Any("_app_trace_id", loggerCtx.TraceID))
	fields = append(fields, zap.Any("_app_span_id", loggerCtx.SpanID))
	z.tdr.Info(fmt.Sprintf("[TDR] | %s", identifier), fields...)
}

func (z engine) Info(ctx context.Context, identifier string, fields ...zap.Field) {
	loggerCtx := GetLoggerCtx(ctx)
	fields = append(fields, zap.Any("_app_trace_id", loggerCtx.TraceID))
	fields = append(fields, zap.Any("_app_span_id", loggerCtx.SpanID))
	z.sys.Info(fmt.Sprintf("[INFO] %s", identifier), fields...)
}

func (z engine) Debug(ctx context.Context, identifier string, fields ...zap.Field) {
	loggerCtx := GetLoggerCtx(ctx)
	fields = append(fields, zap.Any("_app_trace_id", loggerCtx.TraceID))
	fields = append(fields, zap.Any("_app_span_id", loggerCtx.SpanID))
	z.sys.Debug(fmt.Sprintf("[DEBUG] %s", identifier), fields...)
}

func (z engine) Warn(ctx context.Context, identifier string, fields ...zap.Field) {
	loggerCtx := GetLoggerCtx(ctx)
	fields = append(fields, zap.Any("_app_trace_id", loggerCtx.TraceID))
	fields = append(fields, zap.Any("_app_span_id", loggerCtx.SpanID))
	z.sys.Warn(fmt.Sprintf("[WARN] %s", identifier), fields...)
}

func (z engine) Err(ctx context.Context, identifier string, fields ...zap.Field) {
	loggerCtx := GetLoggerCtx(ctx)
	fields = append(fields, zap.Any("_app_trace_id", loggerCtx.TraceID))
	fields = append(fields, zap.Any("_app_span_id", loggerCtx.SpanID))
	fields = append(fields, zap.String("_app_sys_message", identifier))
	z.sys.Error(fmt.Sprintf("[ERR] %s", identifier), fields...)
}

func (z engine) ThirdPartyLogger(ctx context.Context, identifier string, fields ...zap.Field) {
	loggerCtx := GetLoggerCtx(ctx)

	fields = append(fields, zap.Any("_app_trace_id", loggerCtx.TraceID))
	fields = append(fields, zap.Any("_app_span_id", loggerCtx.SpanID))
	fields = append(fields, zap.Any("_app_third_party_id", ctx.Value("_app_third_party_id")))
	fields = append(fields, zap.String("_app_sys_message", identifier))
	z.sys.Info(fmt.Sprintf("[ThirdPartyLog] %s", identifier), fields...)
}

func (z engine) SystemInfo(identifier string, fields ...zap.Field) {
	fields = append(fields, zap.String("_app_sys_message", identifier))
	z.sys.Info(fmt.Sprintf("[SysInfo] %s", identifier), fields...)
}

func (z engine) SystemFailure(identifier string, fields ...zap.Field) {
	fields = append(fields, zap.String("_app_sys_message", identifier))
	z.sys.Info(fmt.Sprintf("[SysErr] %s", identifier), fields...)
}
