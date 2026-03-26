package simple

import (
	"ag-core/contribute/agonet"
	"errors"
)

func ChannelForConn(conn agonet.Conn) (Channel, error) {
	tmpCtx := conn.Context()
	if tmpCtx == nil {
		return nil, errors.New("conn.Context is nil")
	}
	channel, ok := tmpCtx.(Channel)
	if !ok {
		return nil, errors.New("tmpCtx is not simple.Channel")
	}
	return channel, nil
}
