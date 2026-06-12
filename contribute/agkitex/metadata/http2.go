package metadata

import (
	"github.com/aif-go/ag-core/ag/ag_common/agmetadata"
	"context"
	"log/slog"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/trans/nphttp2/metadata"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/cloudwego/kitex/transport"
)

/* = AgKitexClientAgMetadataHTTP2Handler = */
var AgKitexClientAgMetadataHTTP2Handler = &agKitexClientAgMetadataHTTP2Handler{}

type agKitexClientAgMetadataHTTP2Handler struct{}

var (
	_ remote.MetaHandler          = AgKitexClientAgMetadataHTTP2Handler
	_ remote.StreamingMetaHandler = AgKitexClientAgMetadataHTTP2Handler
)

// MetaHandler reads or writes metadata through certain protocol.
func (*agKitexClientAgMetadataHTTP2Handler) WriteMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	return ctx, nil
}
func (*agKitexClientAgMetadataHTTP2Handler) ReadMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	return ctx, nil
}

// StreamingMetaHandler reads or writes metadata through streaming header(http2 header)
// writes metadata before create a stream
func (*agKitexClientAgMetadataHTTP2Handler) OnConnectStream(ctx context.Context) (context.Context, error) {
	ri := rpcinfo.GetRPCInfo(ctx)
	if !isGRPC(ri) {
		return ctx, nil
	}
	rctx := ctx
	var kmd []string
	err := agmetadata.HandlerMdFromContext(ctx, func(k, v string) {
		kmd = append(kmd, k, v)
	})
	if err != nil {
		return ctx, err
	}

	if len(kmd) > 0 {
		rctx = metadata.AppendToOutgoingContext(ctx, kmd...)
	}

	return rctx, nil
}

func (*agKitexClientAgMetadataHTTP2Handler) OnReadStream(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

/* = AgKitexServerAgMetadataHTTP2Handler = */
var AgKitexServerAgMetadataHTTP2Handler = &agKitexServerAgMetadataHTTP2Handler{}

type agKitexServerAgMetadataHTTP2Handler struct{}

var (
	_ remote.MetaHandler          = AgKitexServerAgMetadataHTTP2Handler
	_ remote.StreamingMetaHandler = AgKitexServerAgMetadataHTTP2Handler
)

func (*agKitexServerAgMetadataHTTP2Handler) WriteMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	return ctx, nil
}
func (*agKitexServerAgMetadataHTTP2Handler) ReadMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	return ctx, nil
}

func (*agKitexServerAgMetadataHTTP2Handler) OnConnectStream(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (*agKitexServerAgMetadataHTTP2Handler) OnReadStream(ctx context.Context) (context.Context, error) {
	ri := rpcinfo.GetRPCInfo(ctx)
	if !isGRPC(ri) {
		return ctx, nil
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, nil
	}

	rctx, err := agmetadata.ParseMdToContext(ctx, func(key string) ([]string, bool) {
		tmph := md.Get(key)
		if len(tmph) == 1 {
			return []string{tmph[0]}, true
		} else if len(tmph) > 1 {
			slog.Warn("kitex head key has multiple values, metadata will use the first value", "key", key)
			return []string{tmph[0]}, true
		} else {
			return nil, false
		}
	})

	if err != nil {
		return ctx, err
	}

	return rctx, nil
}

func isGRPC(ri rpcinfo.RPCInfo) bool {
	return ri.Config().TransportProtocol()&transport.GRPC == transport.GRPC
}

func NewAgKitexClientAgMetadataHTTP2HandlerOption() *client.Option {
	opt := client.WithMetaHandler(AgKitexClientAgMetadataHTTP2Handler)
	return &opt
}

func NewAgKitexServerAgMetadataHTTP2HandlerOption() *server.Option {
	opt := server.WithMetaHandler(AgKitexServerAgMetadataHTTP2Handler)
	return &opt
}
