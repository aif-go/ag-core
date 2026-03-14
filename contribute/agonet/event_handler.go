package agonet

import (
	"net"
)

type (
	EventHandler interface {
		OnBoot(eng Engine) (action Action)

		OnShutdown(eng Engine)

		OnOpen(c net.Conn) (out []byte, action Action)

		OnClose(c net.Conn, err error) (action Action)

		OnTraffic(c net.Conn) (action Action)

		// OnTick() (delay time.Duration, action Action)
	}

	BuiltinEventEngine struct{}
)

func (*BuiltinEventEngine) OnBoot(_ Engine) (action Action) {
	return
}

func (*BuiltinEventEngine) OnShutdown(_ Engine) {
}

func (*BuiltinEventEngine) OnOpen(_ net.Conn) (out []byte, action Action) {
	return
}

func (*BuiltinEventEngine) OnClose(_ net.Conn, _ error) (action Action) {
	return
}

func (*BuiltinEventEngine) OnTraffic(_ net.Conn) (action Action) {
	return
}
