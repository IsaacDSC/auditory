package ctxkey

import "context"

var ClientIDCtxKey = struct{}{}

func SetClientID(ctx context.Context, clientID string) context.Context {
	return context.WithValue(ctx, ClientIDCtxKey, clientID)
}

func ClientID(ctx context.Context) string {
	return ctx.Value(ClientIDCtxKey).(string)
}
