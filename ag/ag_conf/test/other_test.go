package test

import (
	"fmt"
	"strings"
	"testing"
)

func TestStringsEqual(t *testing.T) {

	a := "abc"
	b := "aBc"

	fmt.Println(strings.EqualFold(a, b))
}
