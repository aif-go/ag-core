//go:build db

package test

import (
	"fmt"
	"os"
	"testing"
)

// TestMain 未设置数据库 DSN 环境变量时整体跳过集成测试（CI 门禁项，§4.4）。
func TestMain(m *testing.M) {
	if os.Getenv("MYSQL_DSN") == "" && os.Getenv("DB2_DSN") == "" {
		fmt.Println("SKIP: 未设置 MYSQL_DSN / DB2_DSN，跳过数据库集成测试（CI 门禁项）")
		os.Exit(0)
	}
	os.Exit(m.Run())
}
