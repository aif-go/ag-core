package test

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

type RConfig struct {
	Port int
}

func TestReflectConfig(t *testing.T) {
	config := &RConfig{Port: 8080}
	writerDone := make(chan bool)

	// 写者 goroutine（使用反射）
	go func() {
		v := reflect.ValueOf(config).Elem().FieldByName("Port")
		time.Sleep(10 * time.Millisecond) // 让读者先跑一会儿
		v.SetInt(9090)                    // 反射写入
		fmt.Println("Writer: set to 9090")
		writerDone <- true
	}()

	// 读者 goroutine（直接读取）
	go func() {
		for {
			// 直接读取，无锁
			currentPort := config.Port
			if currentPort == 9090 {
				fmt.Println("Reader: saw 9090")
				break
			}
		}
	}()

	<-writerDone
	time.Sleep(100 * time.Millisecond) // 等待读者可能发现
}
