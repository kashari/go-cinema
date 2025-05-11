package theatre

import (
	"context"
)

// contextKey is a custom type to prevent collisions in context values
type contextKey string

// GetParam retrieves a URL parameter from the request context
func GetParam(ctx context.Context, key string) string {
	value, _ := ctx.Value(contextKey(key)).(string)
	return value
}

// contextWithParam adds a parameter to the request context
func contextWithParam(ctx context.Context, key, value string) context.Context {
	return context.WithValue(ctx, contextKey(key), value)
}
