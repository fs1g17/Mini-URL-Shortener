package user_context

import "context"

type contextKey struct{}

type ContextValue struct {
	Username string
	UserID   int
}

func NewContext(ctx context.Context, val *ContextValue) context.Context {
	return context.WithValue(ctx, contextKey{}, val)
}

func FromContext(ctx context.Context) *ContextValue {
	v, ok := ctx.Value(contextKey{}).(*ContextValue)
	if !ok {
		return nil
	}
	return v
}
