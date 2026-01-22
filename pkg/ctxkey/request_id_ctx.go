package ctxkey

import "context"

var RequestIDCtxKey = struct{}{}

func SetRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDCtxKey, requestID)
}

func RequestID(ctx context.Context) string {
	return ctx.Value(RequestIDCtxKey).(string)
}
