package simple

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/simple/utils"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	// _ CodecHandler = (*lengthFieldStrDecoder)(nil)
	// _ Codec               = (*lengthFieldStrDecoder)(nil)
	_ InboundHandler = (*lengthFieldStrDecoder)(nil)
	_ Decoder        = (*lengthFieldStrDecoder)(nil)
)

// LengthFieldCodec create a length field based codec
func NewLengthFieldStrDecoder(
	maxFrameLength int,
	lengthFieldOffset int,
	lengthFieldLength int,
	lengthAdjustment int,
	initialBytesToStrip int,
) Decoder {

	utils.AssertIf(maxFrameLength <= 0, "maxFrameLength must be a positive integer")
	utils.AssertIf(lengthFieldOffset < 0, "lengthFieldOffset must be a non-negative integer")
	utils.AssertIf(initialBytesToStrip < 0, "initialBytesToStrip must be a non-negative integer")
	// utils.AssertIf(lengthFieldLength != 1 && lengthFieldLength != 2 &&
	// 	lengthFieldLength != 4 && lengthFieldLength != 8, "lengthFieldLength must be either 1, 2, 4 or 8")
	utils.AssertIf(lengthFieldOffset > maxFrameLength-lengthFieldLength,
		"maxFrameLength must be equal to or greater than lengthFieldOffset + lengthFieldLength")

	codec := &lengthFieldStrDecoder{
		maxFrameLength:      maxFrameLength,
		lengthFieldOffset:   lengthFieldOffset,
		lengthFieldLength:   lengthFieldLength,
		lengthAdjustment:    lengthAdjustment,
		initialBytesToStrip: initialBytesToStrip,
		// Encoder:             LengthFieldPrepender(byteOrder, lengthFieldLength, 0, false),
	}
	return codec
}

type lengthFieldStrDecoder struct {
	maxFrameLength      int // 最大允许数据包长度
	lengthFieldOffset   int // 长度域的偏移量，表示跳过指定长度个字节之后的才是长度域
	lengthFieldLength   int // 长度域的长度，单位字节
	lengthAdjustment    int // 包体长度调整的大小，长度域的数值表示的长度加上这个修正值表示的就是带header的包长度
	initialBytesToStrip int // 拿到一个完整的数据包之后向业务解码器传递之前，应该跳过多少字节

}

func (*lengthFieldStrDecoder) Name() string {
	return "length-field-string-decoder"
}

func (l *lengthFieldStrDecoder) HandleRead(ctx InboundContext, message any) {
	// reader := utils.MustToReader(message)
	reader, ok := message.(agonet.Reader)
	if ok {
		out, err := l.doDecode(reader)
		if err != nil {
			utils.Assert(err) // 异常直接panic穿透给EventHandler
			return
		}

		for _, item := range out {
			ctx.FireRead(item)
		}
	} else {
		ctx.FireRead(message)
	}
}

func (l *lengthFieldStrDecoder) doDecode(reader agonet.Reader) ([]any, error) {
	// 读取长度域
	lengthFieldEndOffset := l.lengthFieldOffset + l.lengthFieldLength // 长度域的结束偏移量

	lengthBuff, err := reader.Peek(lengthFieldEndOffset)
	if err != nil {
		if errors.Is(err, io.ErrShortBuffer) { // 长度域长度不足
			// return nil, aerrors.ErrIncompletePacket
			// 半包，等待后续数据，不再抛出异常
			return nil, nil
		}
		return nil, err
	}

	// 解析长度域的数值，获取报文长度
	// frameLength := unpackFieldLength(l.byteOrder, l.lengthFieldLength, lengthBuff)
	frameLength := unpackFieldLengthStr(l.lengthFieldLength, lengthBuff)

	if frameLength < 0 {
		return nil, errors.New("invalid frame length")
	}

	// 包体长度修正
	frameLength += int64(l.lengthAdjustment + lengthFieldEndOffset)

	if frameLength < int64(lengthFieldEndOffset) {
		return nil, errors.New("Adjusted frame length is less than lengthFieldEndOffset")
	}

	// 检查报文长度是否超过最大允许长度
	if frameLength > int64(l.maxFrameLength) {
		// TODO exceededFrameLength // TODO 处理超长报文
		return nil, errors.New("exceeded frame length")
	}

	// 检查是否有足够的数据可读取，半包处理
	if reader.InboundBuffered() < int(frameLength) {
		// return nil, aerrors.ErrIncompletePacket
		// 半包，等待后续数据，不再抛出异常
		return nil, nil
	}

	msgLength := frameLength
	if l.initialBytesToStrip > 0 {
		if l.initialBytesToStrip > int(frameLength) {
			// TODO  failOnFrameLengthLessThanInitialBytesToStrip(in, frameLength, initialBytesToStrip);
			return nil, errors.New("Adjustd frame length is less than initialBytesToStrip")
		}

		_, err = reader.Discard(l.initialBytesToStrip)
		if err != nil {
			return nil, err
		}
		msgLength -= int64(l.initialBytesToStrip)
	}

	frameMsg, err := reader.Next(int(msgLength))
	if err != nil {
		if errors.Is(err, io.ErrShortBuffer) { // FIXME 根据前文判断，此处不应该返回错误
			// return nil, aerrors.ErrIncompletePacket
			// 半包，等待后续数据，不再抛出异常
			return nil, nil
		}
		return nil, errors.Join(err, errors.New("read frame message failed"))
	}

	return []any{frameMsg}, nil
}

// 解包：字符串字节数组 → 数字长度
// buff: 字符串的字节数组（如 []byte("0123")）
func unpackFieldLengthStr(fieldLen int, buff []byte) (frameLength int64) {
	// 安全校验
	if len(buff) < fieldLen {
		utils.Assert(fmt.Errorf("buffer too small, fieldLen:%d, buffLen:%d", fieldLen, len(buff)))
	}

	// 截取固定长度字符串
	lenStr := string(buff[:fieldLen])
	// 去掉空白（按需保留）
	lenStr = strings.TrimSpace(lenStr)

	// 字符串转整数
	num, err := strconv.ParseInt(lenStr, 10, 64)
	if err != nil {
		utils.Assert(fmt.Errorf("parse length string failed: %s, err:%v", lenStr, err))
	}

	return num
}

// 打包：数字长度 → 固定长度字符串字节数组
// fieldLen: 字符串长度（如 4 → "0123"）
func packFieldLengthStr(fieldLen int, dataLen int64) []byte {
	// 转字符串
	lenStr := strconv.FormatInt(dataLen, 10)

	// 超长校验（防止溢出）
	if len(lenStr) > fieldLen {
		utils.Assert(fmt.Errorf("length string too long: %s, max:%d", lenStr, fieldLen))
	}

	// 格式化：左补0，固定长度
	// %0*d → 不足补0，如 4,123 → 0123
	formatStr := fmt.Sprintf("%0*d", fieldLen, dataLen)
	// 如需右对齐空格：formatStr := fmt.Sprintf("%*d", fieldLen, dataLen)

	// 转字节数组返回
	return []byte(formatStr)
}
