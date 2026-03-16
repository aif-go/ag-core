package agonet

import (
	"ag-core/contribute/agonet/pkg/aerrors"
	"context"
	"log/slog"
	"net"
	"os"
	"sync"
)

type listener struct {
	openOnce, closeOnce sync.Once
	network             string
	address             string
	lc                  *net.ListenConfig
	ln                  net.Listener
	// pc                  net.PacketConn // udp
	addr net.Addr
}

func (l *listener) open() (err error) {
	l.openOnce.Do(func() {
		switch l.network {
		// case "udp", "udp4", "udp6":
		// 	if l.pc, err = l.lc.ListenPacket(context.Background(), l.network, l.address); err == nil {
		// 		l.addr = l.pc.LocalAddr()
		// 	}
		case "unix":
			_ = os.Remove(l.address)
			fallthrough
		case "tcp", "tcp4", "tcp6":
			if l.ln, err = l.lc.Listen(context.Background(), l.network, l.address); err == nil {
				l.addr = l.ln.Addr()
			}
		default:
			err = aerrors.ErrUnsupportedProtocol
		}
	})
	return
}

func (l *listener) close() {
	l.closeOnce.Do(func() {
		err := l.ln.Close()

		if err != nil {
			slog.Error("close listener failed", "err", err)
		}
	})
}

func createListener(network, addr string, options *Options) (*listener, error) {

	lc := net.ListenConfig{}

	if options.KeepAlive.Enable && options.KeepAlive.Idle > 0 {
		keepAlive := buildKeepAliveWithConfig(options.KeepAlive)
		if keepAlive != nil {
			lc.KeepAliveConfig = *keepAlive
		}
	}

	l := listener{network: network, address: addr, lc: &lc}

	err := l.open()
	if err != nil {
		return &l, err
	}
	return &l, err
}
