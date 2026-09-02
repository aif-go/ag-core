package agonet

import (
	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
	"context"
	"log/slog"
	"net"
	"os"
	"sync"

	"gitee.com/Trisia/gotlcp/pa"
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

	l := &listener{network: network, address: addr, lc: &lc}

	err := l.open()
	if err != nil {
		return l, err
	}

	err = tlsIfNeed(l, options)
	if err != nil {
		return l, err
	}

	return l, err
}

func tlsIfNeed(l *listener, opts *Options) error {
	if l.ln == nil {
		return nil
	}

	// 未声明安全类型：不包装，明文监听（合法默认）。
	// 配置合法性（Type 与 config 对应）已由 Options.ValidateServer 在构造入口保证，
	// 此处仅做包装决策 + 防御性兜底。
	if opts.TLSType == TLSType_NONE || opts.TLSType == TLSType_UNSET {
		return nil
	}

	tlscfg := opts.TLSConfig
	tlcpcfg := opts.TLCPConfig

	if tlscfg == nil && tlcpcfg == nil {
		// 防御：正常路径不会到达（ValidateServer 已拦截），仅防绕过构造器直接调用
		return aerrors.ErrTLSConfigIsNil
	}

	l.ln = pa.NewListener(l.ln, tlcpcfg, tlscfg)

	return nil
}
