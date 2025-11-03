package aghertzclient

import (
	"github.com/bytedance/sonic"
	"google.golang.org/protobuf/proto"
)

const (
	MIMEApplicationJSON = "application/json"
	MIMEPROTOBUF        = "application/x-protobuf"
)

// 内置序列化器实现
var (
	Serializers map[string]Serializer = map[string]Serializer{
		MIMEApplicationJSON: &JsonSerializer{},
		MIMEPROTOBUF:        &ProtobufSerializer{},
	}
)

func GetSerializer(contentType string) Serializer {
	if serializer, ok := Serializers[contentType]; ok {
		return serializer
	}
	return nil
}

func RegisterSerializer(contentType string, serializer Serializer) {
	Serializers[contentType] = serializer
}

// 序列化器接口
type Serializer interface {
	Marshal(v interface{}) ([]byte, error)
	ContentType() string
}

type JsonSerializer struct{}

func (s *JsonSerializer) Marshal(v interface{}) ([]byte, error) {
	return sonic.Marshal(v)
}

func (s *JsonSerializer) ContentType() string {
	return MIMEApplicationJSON
}

type ProtobufSerializer struct{}

func (s *ProtobufSerializer) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}
func (s *ProtobufSerializer) ContentType() string {
	return MIMEPROTOBUF
}
