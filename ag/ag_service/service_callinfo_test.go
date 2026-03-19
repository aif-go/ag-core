package ag_service

import (
	"fmt"
	"testing"
)

func TestCallinfo(t *testing.T) {
	ci := &CallInfo{}
	ci2 := xxx(ci)

	ci2.AddTag("key", "value")

	fmt.Printf("ci2 hasTag: %v, tag: %v\n", ci2.HasTag("key"), ci2.GetTag("key"))

	fmt.Printf("ci1 hasTag: %v, tag: %v\n", ci.HasTag("key"), ci.GetTag("key"))
}

func xxx(ci *CallInfo) CallInfo {
	return *ci
}

type Ccc struct {
}

var ccc = c1().C2()

func c1() *Ccc {
	return &Ccc{}
}
func (c *Ccc) C2() *Ccc {
	return c
}
