package test

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
	"github.com/aif-go/ag-core/contribute/agsarama"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/IBM/sarama"
)

func TestAgsarama(t *testing.T) {
	env, _ := ag_conf.NewStandardEnvironment()

	err := ag_conf.LoadConfigFile(env, "conf.yaml")
	if err != nil {
		t.Fatalf("LoadConfigFile failed: %v", err)
	}

	// ps, err := ag_conf.NewPropertySourceFromFile("conf.yaml")
	// if err != nil {
	// 	t.Fatalf("NewPropertySourceByFile failed: %v", err)
	// }
	// env.GetPropertySources().AddLast(ps)
	binder := ag_conf.NewConfigurationPropertiesBinder(env)

	conf := agsarama.NewDefaultConfig()
	err = binder.Bind(conf, agsarama.AgsaramaConfigPrefix)
	if err != nil {
		t.Fatalf("Bind failed: %v", err)
	}

	confjson, _ := json.MarshalIndent(conf, "", " ")
	fmt.Println(string(confjson))

	// saramaConfig, err := conf.ToSaramaConfig()
	// if err != nil {
	// 	t.Fatalf("ToSaramaConfig failed: %v", err)
	// }

	// fmt.Println(saramaConfig)

	// 根据配置创建sarama client
	client, err := agsarama.NewClientWithAgConfig(conf)
	if err != nil {
		t.Fatalf("NewClientWithAgConfig failed: %v", err)
	}
	defer client.Close()

	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		t.Fatalf("NewClusterAdminFromClient failed: %v", err)
	}
	defer admin.Close()

	// err = admin.CreateTopic(testtopic, &sarama.TopicDetail{
	// 	NumPartitions:     1,
	// 	ReplicationFactor: 1,
	// }, false)
	// if err != nil {
	// 	t.Fatalf("CreateTopic failed: %v", err)
	// }
	// defer func() {
	// 	err = admin.DeleteTopic(testtopic)
	// 	if err != nil {
	// 		t.Fatalf("DeleteTopic failed: %v", err)
	// 	}
	// }()

	// 通过client 测试producer
	err = _testProducerByClient(client)
	if err != nil {
		t.Fatalf("testProducerByClient failed: %v", err)
	}

	// 通过client 测试consumer
	err = _testConsumerByClient(client)
	if err != nil {
		t.Fatalf("testConsumerByClient failed: %v", err)
	}
}

var mcount int = 5
var testtopic string = "agsarama_test"

// _testProducerByClient 测试producer
func _testProducerByClient(client sarama.Client) error {
	syncProducer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return err
	}
	defer syncProducer.Close()

	for i := 0; i < mcount; i++ {
		message := &sarama.ProducerMessage{
			Topic: testtopic,
			Value: sarama.StringEncoder(fmt.Sprintf("hello-%d", i)),
		}

		partition, offset, err := syncProducer.SendMessage(message)
		if err != nil {
			return err
		}
		fmt.Printf("SendMessage: partition=%d, offset=%d\n", partition, offset)
	}
	return nil
}

// _testConsumerByClient 测试consumer
func _testConsumerByClient(client sarama.Client) error {
	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		return err
	}
	defer consumer.Close()

	pConsumer, err := consumer.ConsumePartition(testtopic, 0, 0)
	if err != nil {
		return err
	}
	defer pConsumer.Close()
	mchan := pConsumer.Messages()

	i := 0
	for msg := range mchan {
		if i >= mcount {
			break
		}
		value := string(msg.Value)
		fmt.Printf("ConsumeMessage: %v\n", value)
		i++
	}

	return nil
}
