package metadata

import (
	"github.com/aif-go/ag-core/ag/ag_common/agmetadata"
	"context"
	"log/slog"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/protocol"
)

// HertzServerHeadMiddleware Hertz 服务端头信息中间件
func HertzServerHeadMiddleware(ctx context.Context, r *app.RequestContext) {
	rctx, err := agmetadata.ParseMdToContext(ctx, func(key string) ([]string, bool) {
		tmph := r.GetHeader(key)
		if len(tmph) > 0 {
			return []string{string(tmph)}, true
		} else {
			return nil, false
		}
	})
	if err != nil {
		slog.Error("parse md to context error", "err", err)
		return
	}

	r.Next(rctx)
}

// HertzClientHeadMiddleware  Hertz 客户端头信息中间件
func HertzClientHeadMiddleware(endpoint client.Endpoint) client.Endpoint {
	return func(ctx context.Context, req *protocol.Request, resp *protocol.Response) (err error) {

		err = agmetadata.HandlerMdFromContext(ctx, func(k, v string) {
			req.Header.Set(k, v)
		})
		if err != nil {
			return err
		}

		return endpoint(ctx, req, resp)
	}
}

func NewAgHertzServerAgMetadataMiddleware() app.HandlerFunc {
	return HertzServerHeadMiddleware
}

func NewAgHertzClientAgMetadataMiddleware() client.Middleware {
	return HertzClientHeadMiddleware
}
