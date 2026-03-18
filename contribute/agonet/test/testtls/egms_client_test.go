package testtls

import (
	"io"
	"log"
	"testing"
	"time"

	"gitee.com/Trisia/gotlcp/tlcp"
)

func TestEGM_Client(t *testing.T) {
	authCert, err := egms_LoadAuthCert()
	if err != nil {
		t.Fatalf("加载认证证书失败: %v", err)
	}

	caPool, err := egms_LoadCertPool()
	if err != nil {
		t.Fatalf("加载CA证书失败: %v", err)
	}

	tlcpCfg := &tlcp.Config{
		Certificates: []tlcp.Certificate{*authCert},
		RootCAs:      caPool,
	}

	conn, err := tlcp.Dial("tcp", ":8443", tlcpCfg)
	if err != nil {
		t.Fatalf("连接EGM服务器失败: %v", err)
	}
	defer conn.Close()
	// err = conn.Handshake()
	if err != nil {
		t.Fatalf("国密TLS握手失败: %v", err)
	}

	log.Println("已连接到服务器")

	// 发送消息
	messages := []string{
		"Hello, TLCP!",
		// "How are you?",
		// "Goodbye!",
	}

	for _, msg := range messages {
		// 发送
		log.Printf("发送: %s", msg)
		_, err = conn.Write([]byte(msg + "#"))
		if err != nil {
			log.Fatal("发送失败:", err)
		}

		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		// 接收响应
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Fatal("读取失败:", err)
			}
			break
		}

		log.Printf("收到: %s", string(buf[:n]))

		time.Sleep(1 * time.Second)
	}
}
