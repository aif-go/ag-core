package server

import (
	"errors"
	"log/slog"

	"github.com/cloudwego/kitex/pkg/serviceinfo"
	"github.com/cloudwego/kitex/server"
)

var (
	// ErrServiceInfoIsNil is the error when service info is nil.
	ErrServiceInfoIsNil = errors.New("service info is nil")

	// ErrHandlerIsNil is the error when handler is nil.
	ErrHandlerIsNil = errors.New("handler is nil")
)

type (
	// AgKitexServiceRegistry is the service registry for kitex.
	AgKitexServiceRegistry struct {
		// ServiceName string
		ServiceInfo *serviceinfo.ServiceInfo
		Handler     interface{}
		Opts        []server.RegisterOption
	}

	// AgKitexServiceRegistryHolder is the holder for AgKitexServiceRegistry.
	AgKitexServiceRegistryHolder struct {
		Registries []*AgKitexServiceRegistry
	}
)

// NewAgKitexServiceRegistry creates a new AgKitexServiceRegistry.
func NewAgKitexServiceRegistry(
	// serviceName string,
	serviceInfo *serviceinfo.ServiceInfo,
	handler interface{},
	opts ...server.RegisterOption,
) *AgKitexServiceRegistry {

	// if serviceInfo == nil {
	// 	return nil, ErrServiceInfoIsNil
	// }

	// if handler == nil {
	// 	return nil, ErrHandlerIsNil
	// }

	reg := &AgKitexServiceRegistry{
		// ServiceName: serviceInfo.ServiceName,
		ServiceInfo: serviceInfo,
		Handler:     handler,
		Opts:        opts,
	}

	// return reg, nil
	return reg
}

func (r *AgKitexServiceRegistry) GetServiceName() string {
	if r.ServiceInfo == nil {
		return ""
	}
	return r.ServiceInfo.ServiceName
}

// Register registers the service to the server.
func (r *AgKitexServiceRegistry) Register(svr server.Server, options ...server.RegisterOption) error {
	if r.ServiceInfo == nil {
		return ErrServiceInfoIsNil
	}
	if r.Handler == nil {
		return ErrHandlerIsNil
	}
	slog.Info("register kitex service", "serviceName", r.ServiceInfo.ServiceName)
	return svr.RegisterService(r.ServiceInfo, r.Handler, append(r.Opts, options...)...)
}

// Register registers the services to the server.
func (h *AgKitexServiceRegistryHolder) RegisterService(svr server.Server, options ...server.RegisterOption) error {
	slog.Info("will register kitex services", "count", len(h.Registries))
	for _, reg := range h.Registries {
		if err := reg.Register(svr, options...); err != nil {
			return err
		}
	}
	return nil
}
