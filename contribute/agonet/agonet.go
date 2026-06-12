package agonet

import (
	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
	"math"
	"net/url"
	"path"
	"runtime"
	"strings"
)

// Action is an action that occurs after the completion of an event.
type Action int

const (
	// None indicates that no action should occur following an event.
	None Action = iota

	// Close closes the connection.
	Close

	// Shutdown shutdowns the engine.
	Shutdown
)

const (
	EventLoopIndexMax = math.MaxUint8 + 1
)

func parseProtoAddr(protoAddr string) (string, string, error) {
	// Percent-encode "%" in the address to avoid url.Parse error.
	// This is for cases like this: udp://[ff02::3%lo0]:9991
	protoAddr = strings.ReplaceAll(protoAddr, "%", "%25")

	if runtime.GOOS == "windows" {
		if strings.HasPrefix(protoAddr, "unix://") {
			parts := strings.SplitN(protoAddr, "://", 2)
			if parts[1] == "" {
				return "", "", aerrors.ErrInvalidNetworkAddress
			}
			return parts[0], parts[1], nil
		}
	}

	u, err := url.Parse(protoAddr)
	if err != nil {
		return "", "", err
	}

	switch u.Scheme {
	case "":
		return "", "", aerrors.ErrInvalidNetworkAddress
	case "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6":
		if u.Host == "" || u.Path != "" {
			return "", "", aerrors.ErrInvalidNetworkAddress
		}
		return u.Scheme, u.Host, nil
	case "unix":
		hostPath := path.Join(u.Host, u.Path)
		if hostPath == "" {
			return "", "", aerrors.ErrInvalidNetworkAddress
		}
		return u.Scheme, hostPath, nil
	default:
		return "", "", aerrors.ErrUnsupportedProtocol
	}
}

func determineEventLoops(opts *Options) int {
	numEventLoop := 1
	if opts.Multicore {
		numEventLoop = runtime.NumCPU()
	}
	if opts.NumEventLoop > 0 {
		numEventLoop = opts.NumEventLoop
	}
	if numEventLoop > EventLoopIndexMax {
		numEventLoop = EventLoopIndexMax
	}
	return numEventLoop
}

func createListeners(addrs []string, opts *Options) ([]*listener, error) {
	listeners := make([]*listener, len(addrs))
	for i, a := range addrs {
		proto, addr, err := parseProtoAddr(a)
		if err != nil {
			return nil, err
		}
		ln, err := createListener(proto, addr, opts)
		if err != nil {
			return nil, err
		}

		listeners[i] = ln
	}

	return listeners, nil
}
