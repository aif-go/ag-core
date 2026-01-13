package test

import (
	"fmt"
	"testing"
)

func TestOvertest(t *testing.T) {
	b := &Base{}
	d := &Derived{
		IBase: b,
	}

	var c IBase
	c = d

	fmt.Println(b.Process()) // Base processing
	fmt.Println(d.Process()) // Derived: Base processing
	fmt.Println(c.Process()) // Derived: Base processing
	fmt.Println(d.Helper())  // Helper method (继承)

}

// 基础结构体
type Base struct{}

func (b Base) Process() string {
	return "Base processing"
}

func (b Base) Helper() string {
	return "Helper method"
}

// 派生结构体重写部分方法
type Derived struct {
	IBase // 组合 Base
}

// 重写 Process 方法
func (d Derived) Process() string {
	// 可以调用基类方法
	// baseResult := d.IBase.Process()
	return "Derived: " + "Derived"
}

type IBase interface {
	Process() string
	Helper() string
}
