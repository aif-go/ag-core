package testsimple

import (
	"github.com/aif-go/ag-core/contribute/agonet"

	// "github.com/aif-go/ag-core/contribute/agonet/simple/utils"
	"encoding/binary"
	"fmt"
	"log/slog"
	"net"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	handler := &TestClientEventHandler{}
	client, err := agonet.NewClient(handler, &agonet.ClientConfig{})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	err = client.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	tcon, err := client.Dial("tcp", "localhost:8443")
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	var con net.Conn
	con = tcon
	defer con.Close()

	var writefunc = func(msg string) {
		bmsg := []byte(msg)

		lengthBuff := packFieldLength(binary.BigEndian, 2, int64(len(bmsg)))
		fullmsg := append(lengthBuff, bmsg...)
		con.Write(fullmsg)
	}

	writefunc("123456789")
	time.Sleep(time.Millisecond)

	var length []byte

	// 测试拆包
	fmt.Println("==== 测试拆包 ====")
	length = packFieldLength(binary.BigEndian, 2, 13)
	fmt.Printf("length: %v len: %d\n", length, len(length))
	con.Write(length[:1])
	time.Sleep(time.Millisecond)
	con.Write(length[1:])
	time.Sleep(time.Millisecond)
	con.Write([]byte("who"))
	time.Sleep(time.Millisecond)
	con.Write([]byte("is"))
	time.Sleep(time.Millisecond)
	con.Write([]byte("your"))
	time.Sleep(time.Millisecond)
	con.Write([]byte("dady"))

	// 测试粘包
	fmt.Println("==== 测试粘包 ====")
	length = packFieldLength(binary.BigEndian, 2, 3)
	length2 := packFieldLength(binary.BigEndian, 2, 6)
	zmsg := append(length, []byte("abc")...)
	zmsg = append(zmsg, length2...)
	zmsg = append(zmsg, []byte("def")...)
	fmt.Println("p1")
	con.Write(zmsg)

	time.Sleep(time.Second)

	fmt.Println("p2")
	con.Write([]byte("hig"))
	time.Sleep(time.Millisecond)

	// writefunc("hello")
	// time.Sleep(time.Millisecond)

	// writefunc("hello2")
	// time.Sleep(time.Millisecond)

	// writefunc("11111111")
	// time.Sleep(time.Millisecond)

	// writefunc("张三")
	// time.Sleep(time.Millisecond)
	// writefunc("aaaaa")
	// writefunc("bbbbb")
	// writefunc("ccccc")
	time.Sleep(time.Second * 5)

}

type TestClientEventHandler struct {
	agonet.BuiltinEventEngine
}

func (e *TestClientEventHandler) OnTraffic(c agonet.Conn) (action agonet.Action) {
	fmt.Println("==== OnTraffic ====")
	msg := make([]byte, 50)
	_, err := c.Read(msg)
	if err != nil {
		return agonet.Close
	}

	// hexMsg := hex.EncodeToString(msg)
	// slog.Info(fmt.Sprintf("client Received reply message: %s, hex: %s", msg, hexMsg))
	slog.Info(fmt.Sprintf("client Received reply message: %s", msg))

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
		panic("should not reach here")
	}
	return lengthBuff
}
