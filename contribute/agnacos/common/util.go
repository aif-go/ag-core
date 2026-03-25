package common

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cast"
)

// ParseIPPort 解析IP:端口格式的字符串，返回IP和端口的集合。 eg: 127.0.0.1:8848,127.0.0.2
func ParseIPPort(input string) ([]IpPort, error) {
	var result []IpPort

	// 按逗号分割输入字符串
	entries := strings.Split(input, ",")

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		// 分割IP和端口
		parts := strings.Split(entry, ":")

		host := parts[0]
		if host == "" {
			return nil, fmt.Errorf("无效的主机地址: %s", parts[0])
		}

		// 验证主机格式（IP地址或域名）
		ip := net.ParseIP(host)
		if ip != nil {
			// 是有效的IP地址，使用标准化格式
			host = ip.String()
		}
		// else {
		// 	// 可能是域名，进行基本验证 TODO
		// 	if !isValidDomain(host) {
		// 		return nil, fmt.Errorf("无效的域名格式: %s", host)
		// 	}
		// }

		// 处理端口
		var port uint64
		switch len(parts) {
		case 1:
			// 如果没有指定端口，使用默认端口80
			port = 8848
		case 2:
			// 验证端口格式
			portstr := parts[1]
			var err error
			if _, err = net.LookupPort("tcp", portstr); err != nil {
				return nil, fmt.Errorf("无效的端口: %s", portstr)
			}
			if port, err = cast.ToUint64E(portstr); err != nil {
				return nil, fmt.Errorf("无效的端口: %s", portstr)
			}
		default:
			return nil, errors.New("格式错误，应为IP:端口或IP")
		}

		result = append(result, IpPort{
			Ip:   host,
			Port: port,
		})
	}

	return result, nil
}

type IpPort struct {
	Ip   string
	Port uint64
}
