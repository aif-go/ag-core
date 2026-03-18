package testtls

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"testing"

	"gitee.com/Trisia/gotlcp/pa"
	"gitee.com/Trisia/gotlcp/tlcp"
)

func TestEGM_Server(t *testing.T) {
	sigCert, err := egms_LoadSigCert()
	if err != nil {
		t.Fatalf("加载签名证书失败: %v", err)
	}
	encCert, err := egms_LoadEncCert()
	if err != nil {
		t.Fatalf("加载加密证书失败: %v", err)
	}

	caPool, err := egms_LoadCertPool()
	if err != nil {
		t.Fatalf("加载CA证书失败: %v", err)
	}

	tlcpCfg := &tlcp.Config{
		Certificates: []tlcp.Certificate{*sigCert, *encCert},
		RootCAs:      caPool,
	}

	// listener, err := gmtls.Listen("tcp", ":8443", tlsConfig)
	listener, err := pa.Listen("tcp", ":8443", tlcpCfg, nil)
	if err != nil {
		log.Fatalf("创建国密TLS监听失败: %v", err)
	}
	fmt.Println("📡 纯net+国密TLS服务端启动: [::]:8443")
	defer listener.Close()

	// 5. 接受客户端连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受连接失败: %v", err)
			continue
		}
		fmt.Printf("🔗 新客户端连接: %s\n", conn.RemoteAddr())

		tcpConn, ok := conn.(*net.TCPConn)
		if ok {
			tcpConn.SetKeepAlive(true)
			tcpConn.SetKeepAlivePeriod(30)
		}

		switch a := conn.(type) {
		case *net.TCPConn: // 支持 TCP 连接
			fmt.Printf("TCPConn: %T\n", a)
		default:
			fmt.Printf("连接类型: %T\n", a)
		}

		// pcon, ok := conn.(*pa.ProtocolSwitchServerConn)

		go handleEGMClient(conn) // 异步处理客户端
	}
}

// 处理国密TLS客户端连接
func handleEGMClient(conn net.Conn) {
	defer conn.Close()

	// 类型断言为国密TLS连接
	tlcpConn, ok := conn.(*tlcp.Conn)
	if ok {
		// 手动完成TLS握手
		if err := tlcpConn.Handshake(); err != nil {
			log.Printf("国密TLS握手失败: %v", err)
			return
		}

		// 获取国密TLS连接信息
		state := tlcpConn.ConnectionState()
		fmt.Printf("\n🔗 新客户端连接: %s\n", conn.RemoteAddr())
		fmt.Printf("   TLS版本: %v\n", state.Version)
		fmt.Printf("   加密套件: %v\n", state.CipherSuite)
	}

	// 创建带缓冲的读取器，方便按行读取数据
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('#') // FIXME 此处类似于拆包及粘包处理
		if err != nil {
			if err != io.EOF {
				log.Printf("读取数据失败: %v", err)
			}
			return
		}

		fmt.Printf("📥 收到客户端数据: %s\n", msg)

		response := fmt.Sprintf("R: %s", msg)
		_, err = conn.Write([]byte(response))
		if err != nil {
			log.Printf("发送响应失败: %v", err)
		}
	}
}


