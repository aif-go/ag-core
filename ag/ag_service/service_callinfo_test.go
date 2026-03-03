package ag_service

import (
	"fmt"
	"testing"
)

func TestCallinfo(t *testing.T) {
	ci := &CallInfo{}

	ci.AddTag("key", "value")

	fmt.Printf("hasTag: %v, tag: %v\n", ci.HasTag("key"), ci.GetTag("key"))
}
