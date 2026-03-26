package simple

import (
	"ag-core/contribute/agonet"
	"errors"
	"sync/atomic"
)

type (
	Option func(options *eventHandlerOptions)

	// ChannelInitializer to init the pipeline of channel
	ChannelInitializer func(Channel) error

	// ChannelFactory to create channel
	ChannelFactory func(conn agonet.Conn, pipeline Pipeline) Channel

	// PipelineFactory to create pipeline
	PipelineFactory func() Pipeline
	// ChannelIDFactory to create channel id
	ChannelIDFactory func() int64
)

type eventHandlerOptions struct {
	channelIDFactory   ChannelIDFactory
	channelFactory     ChannelFactory
	channelInitializer ChannelInitializer
	// clientChInitializer ChannelInitializer
	pipelineFactory func() Pipeline
}

func WithChannelInitializer(initializer ChannelInitializer) func(*eventHandlerOptions) {
	return func(o *eventHandlerOptions) {
		o.channelInitializer = initializer
	}
}

// func WithClientChannelInitializer(initializer ChannelInitializer) func(*eventHandlerOptions) {
// 	return func(o *eventHandlerOptions) {
// 		o.clientChInitializer = initializer
// 	}
// }

// SequenceID to generate a sequence id
func SequenceID() ChannelIDFactory {
	var id int64
	return func() int64 {
		return atomic.AddInt64(&id, 1)
	}
}

func (opt *eventHandlerOptions) check() error {
	if opt.channelIDFactory == nil {
		return errors.New("channelIDFactory is nil")
	}
	if opt.channelFactory == nil {
		return errors.New("channelFactory is nil")
	}

	// if opt.channelInitializer == nil && opt.clientChInitializer == nil {
	if opt.channelInitializer == nil {
		return errors.New("channelInitializer is nil")
	}

	if opt.pipelineFactory == nil {
		return errors.New("pipelineFactory is nil")
	}
	return nil
}
