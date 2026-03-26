package codec

import "ag-core/contribute/agonet/simple"

// Codec defines an CodecHandler alias
type (
	Codec   = simple.CodecHandler
	Decoder = simple.DecoderHandler
	Encoder = simple.EncoderHandler
)

// Combine to wrap InboundHandler and OutboundHandler into Codec.
func Combine(name string, inbound simple.InboundHandler, outbound simple.OutboundHandler) Codec {
	return &combineCodec{name: name, InboundHandler: inbound, OutboundHandler: outbound}
}

type combineCodec struct {
	simple.InboundHandler
	simple.OutboundHandler
	name string
}

func (c *combineCodec) Name() string {
	return c.name
}
