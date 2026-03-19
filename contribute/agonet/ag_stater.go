package agonet

import (
	"ag-core/ag/ag_conf"
	"ag-core/ag/ag_server"
	"context"
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
