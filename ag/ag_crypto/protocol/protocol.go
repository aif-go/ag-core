// Package protocol 定义密文的文本协议：{cipher}{name:sm4}{key:v2}payload。
//
// 它是生产方（CLI、加密工具）和消费方（配置、数据库密码等）共享的契约，
// 只负责字符串解析/格式化与语法校验，不依赖任何加解密实现。
//
// payload 对协议层不透明，但有两条边界：
//   - 只识别 {name:...} 和 {key:...} 两个 header；拼错的 header（如 {nam:...}）
//     会被当作 payload 的一部分，不会报“未知 header”；
//   - payload 不能以 {name: 或 {key: 开头，也不能是“以 { 开头且没有 }”的串，
//     否则会被当成 header / 未闭合 header。Base64 编码天然满足这两条。
package protocol

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// Prefix 是密文的固定前缀。
	Prefix = "{cipher}"

	headerName = "name"
	headerKey  = "key"

	maxNameLen  = 32
	maxKeyIDLen = 64
)

// Ref 是解析后的密文引用。
// Payload 对协议层不透明：具体格式由各加密器自己定义。
type Ref struct {
	Name    string // 规范化小写；空 = 默认加密器
	KeyID   string // 空 = 该加密器的默认密钥
	Payload string // 不透明
}

// IsCipherText 判断字符串是否以密文前缀开头。
func IsCipherText(s string) bool {
	return strings.HasPrefix(s, Prefix)
}

// Parse 解析 {cipher}{name:sm4}{key:v2}payload。
//
// ok=false 表示不是密文；ok=true 且 err!=nil 表示是密文但格式非法。
// 只识别 name/key 两个 header；其他 {...} 视为 payload 开始，
// 保证 payload 对协议层不透明。payload 的边界见包注释。
func Parse(s string) (ref Ref, ok bool, err error) {
	if !strings.HasPrefix(s, Prefix) {
		// 旧文档的冒号形式：明确报错，而不是当普通字符串静默跳过
		if strings.HasPrefix(s, "{cipher:") {
			return Ref{}, true, fmt.Errorf(
				"old cipher protocol form %q, use %s{name:...}{key:...}",
				preview(s, 32), Prefix)
		}
		return Ref{}, false, nil
	}

	rest := s[len(Prefix):]
	var name, keyID string
	seenName, seenKey := false, false

parse:
	for strings.HasPrefix(rest, "{") {
		end := strings.IndexByte(rest, '}')
		if end < 0 {
			return Ref{}, true, fmt.Errorf("unterminated cipher header in %q", preview(s, 32))
		}
		body := rest[1:end]
		k, v, hasColon := strings.Cut(body, ":")
		if !hasColon {
			break parse
		}
		switch strings.ToLower(strings.TrimSpace(k)) {
		case headerName:
			if seenName {
				return Ref{}, true, errors.New("duplicate {name:...} header")
			}
			n := strings.ToLower(strings.TrimSpace(v))
			if verr := ValidateName(n); verr != nil {
				return Ref{}, true, verr
			}
			name, seenName = n, true
		case headerKey:
			if seenKey {
				return Ref{}, true, errors.New("duplicate {key:...} header")
			}
			id := strings.TrimSpace(v)
			if verr := ValidateKeyID(id); verr != nil {
				return Ref{}, true, verr
			}
			keyID, seenKey = id, true
		default:
			// 未知 header：payload 从这里开始，保持不透明
			break parse
		}
		rest = rest[end+1:]
	}

	if rest == "" {
		return Ref{}, true, errors.New("empty cipher payload")
	}
	if keyID != "" && name == "" {
		return Ref{}, true, errors.New("{key:...} requires {name:...}")
	}
	return Ref{Name: name, KeyID: keyID, Payload: rest}, true, nil
}

// Format 生成 {cipher}{name:...}{key:...}payload。
func Format(name, keyID, payload string) (string, error) {
	if payload == "" {
		return "", errors.New("empty cipher payload")
	}
	name = strings.ToLower(strings.TrimSpace(name))
	keyID = strings.TrimSpace(keyID)
	if name == "" && keyID != "" {
		return "", errors.New("{key:...} requires {name:...}")
	}

	var b strings.Builder
	b.WriteString(Prefix)
	if name != "" {
		if err := ValidateName(name); err != nil {
			return "", err
		}
		b.WriteByte('{')
		b.WriteString(headerName)
		b.WriteByte(':')
		b.WriteString(name)
		b.WriteByte('}')
	}
	if keyID != "" {
		if err := ValidateKeyID(keyID); err != nil {
			return "", err
		}
		b.WriteByte('{')
		b.WriteString(headerKey)
		b.WriteByte(':')
		b.WriteString(keyID)
		b.WriteByte('}')
	}
	b.WriteString(payload)
	return b.String(), nil
}

// ValidateName 校验加密器名称：小写字母、数字、下划线、连字符。
func ValidateName(name string) error {
	if name == "" {
		return errors.New("empty encryptor name")
	}
	if len(name) > maxNameLen {
		return fmt.Errorf("encryptor name too long: %q", preview(name, maxNameLen))
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			continue
		}
		return fmt.Errorf("invalid encryptor name %q", name)
	}
	return nil
}

// ValidateKeyID 校验密钥版本标识：大小写字母、数字、下划线、连字符、点。
func ValidateKeyID(id string) error {
	if id == "" {
		return errors.New("empty key id")
	}
	if len(id) > maxKeyIDLen {
		return fmt.Errorf("key id too long: %q", preview(id, maxKeyIDLen))
	}
	for i := 0; i < len(id); i++ {
		c := id[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '_' || c == '-' || c == '.' {
			continue
		}
		return fmt.Errorf("invalid key id %q", id)
	}
	return nil
}

func preview(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
