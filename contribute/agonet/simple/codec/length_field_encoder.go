package codec

import (
	"ag-core/contribute/agonet/simple"
	"ag-core/contribute/agonet/simple/utils"
	"encoding/binary"
)

func NewLengthFieldEncoder(
	byteOrder binary.ByteOrder,
	lengthFieldLength int,
	lengthAdjustment int,
	lengthIncludesLengthFieldLength bool,
) Encoder {

	if byteOrder == nil {
		byteOrder = binary.BigEndian
	}

	utils.AssertIf(lengthFieldLength != 1 && lengthFieldLength != 2 &&
		lengthFieldLength != 4 && lengthFieldLength != 8, "lengthFieldLength must be either 1, 2, 3, 4, or 8")

	utils.AssertIf(lengthFieldLength != 1 && lengthFieldLength != 2 &&
		lengthFieldLength != 4 && lengthFieldLength != 8, "lengthFieldLength must be either 1, 2, 3, 4, or 8")

	return &lengthFieldEncoder{
		byteOrder:                       byteOrder,
		lengthFieldLength:               lengthFieldLength,
		lengthAdjustment:                lengthAdjustment,
		lengthIncludesLengthFieldLength: lengthIncludesLengthFieldLength,
	}
}

type lengthFieldEncoder struct {
	byteOrder                       binary.ByteOrder
	lengthFieldLength               int
	lengthAdjustment                int
	lengthIncludesLengthFieldLength bool
}

func (l *lengthFieldEncoder) Name() string {
	return "length-field-encoder"
}

func (l *lengthFieldEncoder) HandleWrite(ctx simple.OutboundContext, message simple.Message) {
	bodyBytes := utils.MustToBytes(message)

	length := len(bodyBytes) + l.lengthAdjustment
	if l.lengthIncludesLengthFieldLength {
		length += l.lengthFieldLength
	}

	// head buffer
	lengthBuff := packFieldLength(l.byteOrder, l.lengthFieldLength, int64(length))

	// HEAD | BODY
	ctx.FireWrite(append(lengthBuff, bodyBytes...))
}
