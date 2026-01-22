package ctxkey

import "context"

var CorrelationIDCtxKey = struct{}{}

func SetCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, CorrelationIDCtxKey, correlationID)
}

func CorrelationID(ctx context.Context) string {
	return ctx.Value(CorrelationIDCtxKey).(string)
}
