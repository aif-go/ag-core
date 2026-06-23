package ag_service

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
)

func recoveryMiddleware() MiddlewareFunc {
	return func(next Endpoint) Endpoint {
		return func(ctx context.Context, req interface{}) (_ interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					stack := make([]byte, 4096)
					n := runtime.Stack(stack, false)
					slog.ErrorContext(ctx, "recovered from panic",
						"err", r,
						"stack", string(stack[:n]))
					err = fmt.Errorf("recovered from panic: %v", r)
				}
			}()
			return next(ctx, req)
		}
	}
}
