package simple

// Codec defines an CodecHandler alias
type (
	Codec   = CodecHandler
	Decoder = DecoderHandler
	Encoder = EncoderHandler
)

// Combine to wrap InboundHandler and OutboundHandler into Codec.
func Combine(name string, inbound InboundHandler, outbound OutboundHandler) Codec {
	return &combineCodec{name: name, InboundHandler: inbound, OutboundHandler: outbound}
}

type combineCodec struct {
	InboundHandler
	OutboundHandler
	name string
}

func (c *combineCodec) Name() string {
	return c.name
}
