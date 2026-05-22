package agslog

import "context"

type (
	HandlerStartCtxKey struct{}
)

func HandlerStartNameFromContext(ctx context.Context) string {
	n := ctx.Value(HandlerStartCtxKey{})
	return parseHandlerName(n)
}

func parseHandlerName(n any) string {
	if n == nil {
		return ""
	}
	ns, ok := n.(string)
	if !ok {
		return ""
	}
	return ns
}
