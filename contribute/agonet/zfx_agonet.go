package agonet

import (
	"ag-core/ag/ag_conf"
	"context"

	"ag-core/ag/ag_server"

	"go.uber.org/fx"
)

var FxAgonetServerModule = fx.Module("fx_agonet_server",
	fx.Provide(
		NewServerConfig,
		NewServer,
		WarpServer,
	),
)

var FxAgonetClientModule = fx.Module("fx_agonet_client",
	fx.Provide(
		NewClientConfig,
		NewClient,
	),
)

const (
	ServerConfigKey = "agonet.server"
	ClientConfigKey = "agonet.client"
)

func NewServerConfig(binder ag_conf.IBinder) (*ServerConfig, error) {
	cfg := DefaultServerConfig()
	serverConfig := &cfg
	err := binder.Bind(serverConfig, ServerConfigKey)
	if err != nil {
		return nil, err
	}
	return serverConfig, nil
}

func NewClientConfig(binder ag_conf.IBinder) (*ClientConfig, error) {
	cfg := DefaultClientConfig()
	clientConfig := &cfg
	err := binder.Bind(clientConfig, ClientConfigKey)
	if err != nil {
		return nil, err
	}
	return clientConfig, nil
}

func WarpServer(server Server) ag_server.Server {
	return &agServer{
		rawserver: server,
	}
}

type agServer struct {
	rawserver Server
}

func (s *agServer) Start(ctx context.Context) error {
	return s.rawserver.Start()
}

func (s *agServer) Stop(ctx context.Context) error {
	return s.rawserver.Stop()
}
