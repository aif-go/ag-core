package simple

import (
	"github.com/aif-go/ag-core/contribute/agonet"
	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
	"context"
)

// ChannelFromConn 从连接中获取通道
func ChannelFromConn(conn agonet.Conn) (Channel, error) {
	var ch Channel
	var rerr error

	wait := make(chan error)
	defer close(wait)
	times := 0
	for times < 3 {
		err2 := conn.EventLoop().Execute(context.Background(), agonet.RunnableFunc(func(_ context.Context) error {
			defer func() {
				wait <- rerr
			}()
			ch, rerr = getChannelFromConn(conn)
			return nil
		}))

		if nil != err2 {
			return nil, err2
		}

		<-wait
		if rerr != nil {
			switch rerr {
			case aerrors.ErrConnContextIsNil:
				times++
				continue
			}
			return nil, rerr
		} else {
			return ch, nil
		}
	}

	return nil, aerrors.ErrChannelNotFound
}
