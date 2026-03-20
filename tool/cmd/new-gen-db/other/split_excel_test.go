package other

import "testing"

func TestSplitExcelByKeyword(t *testing.T) {
	// 创建一个临时Excel文件
	filePath := "../online-struct-xmgj.xlsx"
	outputPath := "../"
	keyword := "自定义脚本名字"
	err := SplitExcelByKeyword(filePath, outputPath, keyword)
	if err != nil {
		t.Errorf("SplitExcelByKeyword failed: %v", err)
	}
}


func TestFormat(t *testing.T){
	
}