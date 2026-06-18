package agsarama

import (
	"fmt"
	"testing"

	"github.com/IBM/sarama"
)

func TestPartitionerType_ToSarama(t *testing.T) {
	tests := []struct {
		name    string
		input   PartitionerType
		wantErr bool
	}{
		{"hash", PartitionerTypeHash, false},
		{"manual lowercase", PartitionerTypeManual, false},
		{"Manual uppercase", "Manual", false},
		{"MANUAL allcaps", "MANUAL", false},
		{"random", PartitionerTypeRandom, false},
		{"roundrobin", PartitionerTypeRoundRobin, false},
		{"invalid", "unknown", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.ToSarama()
			if tt.wantErr {
				if err == nil {
					t.Fatal("ToSarama() expected error, got nil")
				}
				if got != nil {
					t.Fatal("ToSarama() expected nil constructor for invalid type")
				}
				return
			}
			if err != nil {
				t.Fatalf("ToSarama() error = %v, want nil", err)
			}
			if got == nil {
				t.Fatal("ToSarama() returned nil constructor")
			}
		})
	}
}

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

	err = ExtendSaramaConfigWithOptions(saramaConfig)
	if err != nil {
		t.Errorf("ExtendSaramaConfigWithOptions failed: %v", err)
	}
}
