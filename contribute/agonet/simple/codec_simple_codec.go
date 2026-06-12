package simple

import (
	"github.com/aif-go/ag-core/contribute/agonet/simple/utils"
)

func NewSimpleCodec[T any, U any](name string, decode func(T) ([]any, error), encode func(U) ([]any, error)) Codec {
	return &SimpleCodec[T, U]{
		CodeName: name,
		Decode:   decode,
		Encode:   encode,
	}
}

type SimpleCodec[T any, U any] struct {
	CodeName string
	Decode   func(T) ([]any, error)

	Encode func(U) ([]any, error)
}

func (c *SimpleCodec[T, U]) Name() string {
	return c.CodeName
}

func (c *SimpleCodec[T, U]) HandleRead(ctx InboundContext, message any) {

	if c.Decode == nil {
		ctx.FireRead(message)
		return
	}

	input, ok := message.(T)
	if ok {
		out, err := c.Decode(input) // TODO out 是否池化复用
		if err != nil {
			utils.Assert(err) // TODO 直接panic吗？ 异常处理，如长度不够， 重要！！！！！
		}

		for _, item := range out {
			ctx.FireRead(item)
		}
	} else {
		ctx.FireRead(message)
	}
}

func (c *SimpleCodec[T, U]) HandleWrite(ctx OutboundContext, message any) {
	if c.Encode == nil {
		ctx.FireWrite(message)
		return
	}
	in, ok := message.(U)
	if ok {
		out, err := c.Encode(in)
		if err != nil {
			utils.Assert(err) // TODO 直接panic吗？ 异常处理，如长度不够， 重要！！！！！
		}
		for _, item := range out {
			ctx.FireWrite(item)
		}
	} else {
		ctx.FireWrite(message)
	}
}
