package agslog

import "context"

type (
	// HandlerNameCtxKey  struct{}
	HandlerStartCtxKey struct{}
	// HandlerEndCtxKey   struct{}
)

// func HandlerNameFromContext(ctx context.Context) string {
// 	n := ctx.Value(HandlerNameCtxKey{})
// 	return parseHandlerName(n)
// }

func HandlerStartNameFromContext(ctx context.Context) string {
	n := ctx.Value(HandlerStartCtxKey{})
	return parseHandlerName(n)
}

// func HandlerEndNameFromContext(ctx context.Context) string {
// 	n := ctx.Value(HandlerEndCtxKey{})
// 	return parseHandlerName(n)
// }

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
