//go:build db

package test

import (
	"fmt"
	"os"
	"testing"
)

// TestMain 未设置数据库 DSN 环境变量时整体跳过集成测试（CI 门禁项，§4.4）。
func TestMain(m *testing.M) {
	os.Setenv("MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/process?charset=utf8mb4&parseTime=true&loc=Local") // 测试统一使用上海时区，避免时区差异导致的时间断言失败
	if os.Getenv("MYSQL_DSN") == "" && os.Getenv("DB2_DSN") == "" {
		fmt.Println("SKIP: 未设置 MYSQL_DSN / DB2_DSN，跳过数据库集成测试（CI 门禁项）")
		fmt.Println("设置示例:")
		fmt.Println(`  bash:       export MYSQL_DSN='user:pass@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=Local'`)
		fmt.Println(`  PowerShell: $env:MYSQL_DSN='user:pass@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=true&loc=Local'`)
		fmt.Println(`  DB2:        $env:DB2_DSN='HOSTNAME=127.0.0.1;PORT=50000;DATABASE=testdb;UID=user;PWD=pass'`)
		fmt.Println(`可选 $env:DB_TYPE='mysql'|'db2'（两个 DSN 同时设置需区分时用，默认自动推导）`)
		os.Exit(0)
	}
	os.Exit(m.Run())
}
