package agmetadata

import (
	"context"
	"fmt"
	"testing"
	"time"
)

type agHeadKey2 struct{}

func TestMetadata(t *testing.T) {
	ctx := context.Background()

	ctx1 := AppendMdToContext(ctx, MD{"a": "a"})

	ctx2 := AppendMdToContext(ctx1, MD{"b": "b"})

	md := GetMdFromContext(ctx)
	fmt.Println(md) // map[]

	md1 := GetMdFromContext(ctx1)
	fmt.Println(md1) // map[a:a b:b]

	md2 := GetMdFromContext(ctx2)
	fmt.Println(md2) // map[a:a b:b]

	cv := ctx2.Value(agHeadKey2{})
	fmt.Println(cv) // <nil>

	// 设计上禁止在上下文外部修改元数据
	md2["c"] = "c"
	md1 = GetMdFromContext(ctx1)
	_, ok := md1["c"]
	if ok {
		t.Fatal("md1 should not contain key c")
	}

	fmt.Println(md1) // map[a:a b:b]

	md2 = GetMdFromContext(ctx2)
	fmt.Println(md2) // map[a:a b:b]
}

func TestMetadata2(t *testing.T) {
	md := MD{"a": "a"}
	AppendMD(md, MD{"a": "a1"})

	fmt.Println(md) // map[a:a1]
}

func TestMetadata3(t *testing.T) {
	ctx := context.Background()

	n1 := time.Now()
	ctx1 := context.WithValue(ctx, "aaa", n1)

	time.Sleep(time.Second)

	n2 := time.Now()
	ctx2 := context.WithValue(ctx1, "aaa", n2)

	aaa1 := ctx1.Value("aaa")
	fmt.Println(aaa1) // a1

	aaa2 := ctx2.Value("aaa")
	fmt.Println(aaa2) // a2
}
