package testtls

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"testing"
)

func TestServer(t *testing.T) {
	// 加载证书和密钥
	// cert, err := tls.LoadX509KeyPair("certs/tls/server.crt", "certs/tls/server.key")
	cert, err := tls.LoadX509KeyPair("../certs/fgmsm/rsa_sign.cer", "../certs/fgmsm/rsa_sign_key.pem")
	if err != nil {
		log.Fatal("证书加载失败:", err)
	}

	// TLS 配置
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,

		// 关闭SAN校验
		InsecureSkipVerify: true, // 禁用证书验证（仅用于测试）
	}

	// 创建 TLS 监听器
	listener, err := tls.Listen("tcp", ":8443", config)
	if err != nil {
		log.Fatal("监听失败:", err)
	}
	defer listener.Close()

	fmt.Println("TLS Echo 服务器启动在 :8443")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受连接失败: %v", err)
			continue
		}

		go handleEchoConnection(conn)
	}
}

func handleEchoConnection(conn net.Conn) {
	defer conn.Close()

	tlsConn := conn.(*tls.Conn)

	// 可选：强制完成握手
	if err := tlsConn.Handshake(); err != nil {
		log.Printf("握手失败: %v", err)
		return
	}

	// 获取连接信息
	state := tlsConn.ConnectionState()
	log.Printf("新连接来自: %s, TLS版本: %x, 加密套件: %s",
		conn.RemoteAddr(),
		state.Version,
		tls.CipherSuiteName(state.CipherSuite))

	// Echo 逻辑
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Printf("读取错误: %v", err)
			}
			break
		}

		log.Printf("收到 %d 字节: %s", n, string(buf[:n]))

		// 原样返回
		_, err = conn.Write(buf[:n])
		if err != nil {
			log.Printf("写入错误: %v", err)
			break
		}
	}
}
