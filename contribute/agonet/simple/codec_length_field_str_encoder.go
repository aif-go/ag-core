package simple

import (
	"github.com/aif-go/ag-core/contribute/agonet/simple/utils"
)

func NewLengthFieldStrEncoder(
	lengthFieldLength int,
	lengthAdjustment int,
	lengthIncludesLengthFieldLength bool,
) Encoder {

	// utils.AssertIf(lengthFieldLength != 1 && lengthFieldLength != 2 &&
	// 	lengthFieldLength != 4 && lengthFieldLength != 8, "lengthFieldLength must be either 1, 2, 4 or 8")

	return &lengthFieldStrEncoder{
		lengthFieldLength:               lengthFieldLength,
		lengthAdjustment:                lengthAdjustment,
		lengthIncludesLengthFieldLength: lengthIncludesLengthFieldLength,
	}
}

type lengthFieldStrEncoder struct {
	lengthFieldLength               int
	lengthAdjustment                int
	lengthIncludesLengthFieldLength bool
}

func (l *lengthFieldStrEncoder) Name() string {
	return "length-field-string-encoder"
}

func (l *lengthFieldStrEncoder) HandleWrite(ctx OutboundContext, message any) {
	bodyBytes := utils.MustToBytes(message)

	length := len(bodyBytes) + l.lengthAdjustment
	if l.lengthIncludesLengthFieldLength {
		length += l.lengthFieldLength
	}

	// head buffer
	lengthBuff := packFieldLengthStr(l.lengthFieldLength, int64(length))

	// HEAD | BODY
	ctx.FireWrite(append(lengthBuff, bodyBytes...))
}
