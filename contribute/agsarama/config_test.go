package agsarama

import (
	"fmt"
	"testing"

	"github.com/IBM/sarama"
)

func TestSaramaConfig(t *testing.T) {
	sconf := sarama.NewConfig()
	fmt.Sprintf("%v", sconf)
}

func TestNewConfig(t *testing.T) {
	conf := NewDefaultConfig()
	if conf == nil {
		t.Fatal("NewConfig returned nil")
	}
}

func TestToSaramaConfig(t *testing.T) {
	conf := NewDefaultConfig()

	// 设置一些配置
	conf.ClientID = "test-client"
	conf.Producer.RequiredAcks = RequiredAcksWaitForAll
	conf.Producer.Compression = CompressionGZIP
	conf.Consumer.IsolationLevel = IsolationLevelReadCommitted

	saramaConfig, err := conf.ToSaramaConfig()
	if err != nil {
		t.Fatalf("ToSaramaConfig failed: %v", err)
	}

	if saramaConfig.ClientID != "test-client" {
		t.Errorf("ClientID mismatch: got %s, want %s", saramaConfig.ClientID, "test-client")
	}

	if saramaConfig.Producer.RequiredAcks != sarama.WaitForAll {
		t.Errorf("RequiredAcks mismatch: got %v, want %v", saramaConfig.Producer.RequiredAcks, sarama.WaitForAll)
	}

	if saramaConfig.Producer.Compression != sarama.CompressionGZIP {
		t.Errorf("Compression mismatch: got %v, want %v", saramaConfig.Producer.Compression, sarama.CompressionGZIP)
	}

	if saramaConfig.Consumer.IsolationLevel != sarama.ReadCommitted {
		t.Errorf("IsolationLevel mismatch: got %v, want %v", saramaConfig.Consumer.IsolationLevel, sarama.ReadCommitted)
	}
}

func TestValidate(t *testing.T) {
	conf := NewDefaultConfig()

	err := conf.Validate()
	if err != nil {
		t.Errorf("Validate failed on default config: %v", err)
	}

	// 测试无效配置
	conf.Version = "invalid-version"
	err = conf.Validate()
	if err == nil {
		t.Error("Validate should fail on invalid version")
	}
}

func TestExtendSaramaConfigWithOptions(t *testing.T) {
	conf := NewDefaultConfig()
	saramaConfig, err := conf.ToSaramaConfig()
	if err != nil {
		t.Fatalf("ToSaramaConfig failed: %v", err)
	}

	// 使用 config_options.go 中的函数
	// 先添加一些示例配置选项
	err = ExtendSaramaConfigWithOptions(saramaConfig)
	if err != nil {
		t.Errorf("ExtendSaramaConfigWithOptions failed: %v", err)
	}
}
