package simple

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/pkg/aerrors"
	"ag-core/contribute/agonet/simple/utils"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

var (
	// _ CodecHandler = (*lengthFieldDecoder)(nil)
	// _ Codec               = (*lengthFieldDecoder)(nil)
	_ InboundHandler = (*lengthFieldDecoder)(nil)
	_ Decoder        = (*lengthFieldDecoder)(nil)
)

// LengthFieldCodec create a length field based codec
func NewLengthFieldDecoder(
	byteOrder binary.ByteOrder,
	maxFrameLength int,
	lengthFieldOffset int,
	lengthFieldLength int,
	lengthAdjustment int,
	initialBytesToStrip int,
) Decoder {

	utils.AssertIf(maxFrameLength <= 0, "maxFrameLength must be a positive integer")
	utils.AssertIf(lengthFieldOffset < 0, "lengthFieldOffset must be a non-negative integer")
	utils.AssertIf(initialBytesToStrip < 0, "initialBytesToStrip must be a non-negative integer")
	utils.AssertIf(lengthFieldLength != 1 && lengthFieldLength != 2 &&
		lengthFieldLength != 4 && lengthFieldLength != 8, "lengthFieldLength must be either 1, 2, 3, 4, or 8")
	utils.AssertIf(lengthFieldOffset > maxFrameLength-lengthFieldLength,
		"maxFrameLength must be equal to or greater than lengthFieldOffset + lengthFieldLength")

	if byteOrder == nil {
		byteOrder = binary.BigEndian
	}
	codec := &lengthFieldDecoder{
		byteOrder:           byteOrder,
		maxFrameLength:      maxFrameLength,
		lengthFieldOffset:   lengthFieldOffset,
		lengthFieldLength:   lengthFieldLength,
		lengthAdjustment:    lengthAdjustment,
		initialBytesToStrip: initialBytesToStrip,
		// Encoder:             LengthFieldPrepender(byteOrder, lengthFieldLength, 0, false),
	}
	return codec
}

type lengthFieldDecoder struct {
	byteOrder           binary.ByteOrder // 字节序，大端 & 小端
	maxFrameLength      int              // 最大允许数据包长度
	lengthFieldOffset   int              // 长度域的偏移量，表示跳过指定长度个字节之后的才是长度域
	lengthFieldLength   int              // 长度域的长度，单位字节
	lengthAdjustment    int              // 包体长度调整的大小，长度域的数值表示的长度加上这个修正值表示的就是带header的包长度
	initialBytesToStrip int              // 拿到一个完整的数据包之后向业务解码器传递之前，应该跳过多少字节

	// OutboundHandler
	// Encoder
}

func (*lengthFieldDecoder) Name() string {
	return "length-field-decoder"
}

func (l *lengthFieldDecoder) HandleRead(ctx InboundContext, message any) {
	// reader := utils.MustToReader(message)
	reader, ok := message.(agonet.Reader)
	if ok {
		out, err := l.doDecode(reader) // TODO out 是否池化复用
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

func (l *lengthFieldDecoder) doDecode(reader agonet.Reader) ([]any, error) {

	// 读取长度域
	lengthFieldEndOffset := l.lengthFieldOffset + l.lengthFieldLength // 长度域的结束偏移量

	lengthBuff, err := reader.Peek(lengthFieldEndOffset)
	if err != nil {
		if errors.Is(err, io.ErrShortBuffer) { // 长度域长度不足
			return nil, aerrors.ErrIncompletePacket
		}
		return nil, err
	}

	// 解析长度域的数值，获取报文长度
	frameLength := unpackFieldLength(l.byteOrder, l.lengthFieldLength, lengthBuff)

	if frameLength < 0 {
		return nil, errors.New("invalid frame length")
	}

	// 包体长度修正
	frameLength += int64(l.lengthAdjustment + lengthFieldEndOffset)

	if frameLength < int64(lengthFieldEndOffset) {
		return nil, errors.New("Adjusted frame length is less than lengthFieldEndOffset")
	}

	if frameLength > int64(l.maxFrameLength) {
		// TODO exceededFrameLength // TODO 处理超长报文
		return nil, errors.New("exceeded frame length")
	}

	if reader.ReadableBytes() < int(frameLength) {
		return nil, nil
	}

	if l.initialBytesToStrip > int(frameLength) {
		// TODO  failOnFrameLengthLessThanInitialBytesToStrip(in, frameLength, initialBytesToStrip);
		return nil, errors.New("Adjustd frame length is less than initialBytesToStrip")
	}

	frameMsg, err := reader.Next(int(frameLength))
	if err != nil {
		if errors.Is(err, io.ErrShortBuffer) { // FIXME 根据前文判断，此处不应该返回错误
			return nil, aerrors.ErrIncompletePacket
		}
		return nil, errors.Join(err, errors.New("read frame message failed"))
	}

	frameMsg = frameMsg[l.initialBytesToStrip:]

	return []any{frameMsg}, nil
}

func unpackFieldLength(byteOrder binary.ByteOrder, fieldLen int, buff []byte) (frameLength int64) {
	switch fieldLen {
	case 1:
		frameLength = int64(buff[0])
	case 2:
		frameLength = int64(byteOrder.Uint16(buff))
	case 4:
		frameLength = int64(byteOrder.Uint32(buff))
	case 8:
		frameLength = int64(byteOrder.Uint64(buff))
	default:
		utils.Assert(fmt.Errorf("should not reach here"))
	}
	return
}

func packFieldLength(byteOrder binary.ByteOrder, fieldLen int, dataLen int64) []byte {
	lengthBuff := make([]byte, fieldLen)
	switch fieldLen {
	case 1:
		lengthBuff[0] = byte(dataLen)
	case 2:
		byteOrder.PutUint16(lengthBuff, uint16(dataLen))
	case 4:
		byteOrder.PutUint32(lengthBuff, uint32(dataLen))
	case 8:
		byteOrder.PutUint64(lengthBuff, uint64(dataLen))
	default:
		utils.Assert(fmt.Errorf("should not reach here"))
	}
	return lengthBuff
}

// private void exceededFrameLength(ByteBuf in, long frameLength) {
//     long discard = frameLength - in.readableBytes();
//     tooLongFrameLength = frameLength;

//     if (discard < 0) {
//         // buffer contains more bytes then the frameLength so we can discard all now
//         in.skipBytes((int) frameLength);
//     } else {
//         // Enter the discard mode and discard everything received so far.
//         discardingTooLongFrame = true;
//         bytesToDiscard = discard;
//         in.skipBytes(in.readableBytes());
//     }
//     failIfNecessary(true);
// }
