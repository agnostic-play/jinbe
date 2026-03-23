package logger

import (
	`context`

	"github.com/google/uuid"
)

const (
	AppLoggerCtx  = "_app_logger_ctx"
	SystemTraceID = "X-TRACE-ID"
)

type Ctx struct {
	SpanID       string `json:"span_id"`
	ThirdPartyID string `json:"third_party_id"`
	TraceID      string `json:"trace_id"`
}

func NewLoggerCtx(ctx context.Context, traceID string) (context.Context, Ctx) {
	loggerCtx := Ctx{
		SpanID:  uuid.New().String(),
		TraceID: traceID,
	}

	ctx = context.WithValue(ctx, AppLoggerCtx, loggerCtx)

	return ctx, loggerCtx
}

func GetLoggerCtx(ctx context.Context) *Ctx {
	if val := ctx.Value(AppLoggerCtx); val != nil {
		if val, ok := val.(*Ctx); ok {
			if val != nil {
				return val
			}
		}
	}

	return &Ctx{}
}
