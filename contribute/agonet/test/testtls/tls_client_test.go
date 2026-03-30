package testtls

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"os"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	// 加载 CA 证书（用于验证服务器）
	// caCert, err := os.ReadFile("certs/tls/ca.crt")
	caCert, err := os.ReadFile("../certs/fgmsm/RSA_CA.cer")
	if err != nil {
		log.Fatal("读取CA证书失败:", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatal("添加CA证书失败")
	}

	// TLS 配置
	config := &tls.Config{
		RootCAs:    caCertPool,
		ServerName: "localhost", // 必须匹配证书中的 CN 或 SAN
		MinVersion: tls.VersionTLS12,
		// 关闭SAN校验
		InsecureSkipVerify: true, // 禁用证书验证（仅用于测试）
	}

	// 建立 TLS 连接
	conn, err := tls.Dial("tcp", "localhost:8443", config)
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	defer func() {
		conn.CloseWrite()
		time.Sleep(time.Second)
		// if err := conn.Close(); err != nil {
		// 	log.Printf("关闭连接失败: %v", err)
		// }
	}()

	log.Println("已连接到服务器")

	// 发送消息
	messages := []string{
		"Hello, TLS!",
		"How are you?",
		"Goodbye!",
	}

	for _, msg := range messages {
		// 发送
		log.Printf("发送: %s", msg)
		_, err = conn.Write([]byte(msg + "\n"))
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
