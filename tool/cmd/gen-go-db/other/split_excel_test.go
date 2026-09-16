package other

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/360EntSecGroup-Skylar/excelize"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/utils"
)

// buildFixtureWorkbook 在测试内构造 xlsx fixture，替代外部文件 online-struct-xmgj.xlsx（§5.4）。
// sheet1（simple_table）：不含关键字，整体复制；
// sheet2（rule_table）：中部含关键字行，拆分为 rule_table 与 rule_table@。
// 注意：copyAterSheet 在关键字后仅剩 1 行时会早退（既有行为），故关键字后放置 2 行数据。
func buildFixtureWorkbook(t *testing.T, path string) {
	t.Helper()
	f := excelize.NewFile()

	sheet1 := "simple_table"
	f.NewSheet(sheet1)
	f.SetCellValue(sheet1, "A1", "id")
	f.SetCellValue(sheet1, "B1", "name")
	f.SetCellValue(sheet1, "A2", "1")
	f.SetCellValue(sheet1, "B2", "alice")

	sheet2 := "rule_table"
	f.NewSheet(sheet2)
	f.SetCellValue(sheet2, "A1", "id")
	f.SetCellValue(sheet2, "B1", "address")
	f.SetCellValue(sheet2, "A2", "1")
	f.SetCellValue(sheet2, "B2", "北京市")
	f.SetCellValue(sheet2, "A3", "自定义脚本名字")
	f.SetCellValue(sheet2, "B3", "BIZ_DATE in @BizDateSlice")
	f.SetCellValue(sheet2, "A4", "2")
	f.SetCellValue(sheet2, "B4", "上海市")
	f.SetCellValue(sheet2, "A5", "3")
	f.SetCellValue(sheet2, "B5", "广州市")

	f.DeleteSheet("Sheet1")
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
}

// getSheetRows 读取输出 workbook 指定 sheet 的全部行，sheet 不存在则直接失败。
func getSheetRows(t *testing.T, path, sheet string) [][]string {
	t.Helper()
	out, err := excelize.OpenFile(path)
	if err != nil {
		t.Fatalf("打开输出文件失败: %v", err)
	}
	if out.GetSheetIndex(sheet) == -1 {
		t.Fatalf("输出缺少期望的 sheet: %s（实际 sheets: %v）", sheet, out.GetSheetMap())
	}
	return out.GetRows(sheet)
}

// joinRows 将全部行拼接为便于断言的字符串。
func joinRows(rows [][]string) string {
	parts := make([]string, 0, len(rows))
	for _, row := range rows {
		parts = append(parts, strings.Join(row, ","))
	}
	return strings.Join(parts, ";")
}

// TestSplitExcelByKeyword_NoKeyword 整体复制：不含关键字的 sheet 原样出现在输出中（含表名表头行）。
func TestSplitExcelByKeyword_NoKeyword(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.xlsx")
	buildFixtureWorkbook(t, input)

	output := filepath.Join(dir, "out.xlsx")
	if err := SplitExcelByKeyword(input, output, "自定义脚本名字"); err != nil {
		t.Fatalf("SplitExcelByKeyword failed: %v", err)
	}

	joined := joinRows(getSheetRows(t, output, "simple_table"))
	if !strings.Contains(joined, "表名,simple_table") {
		t.Errorf("simple_table 应带表名表头行: %s", joined)
	}
	if !strings.Contains(joined, "id,name") || !strings.Contains(joined, "1,alice") {
		t.Errorf("simple_table 数据行缺失或被改变: %s", joined)
	}
	if strings.Contains(joined, "自定义脚本名字") {
		t.Errorf("simple_table 不应包含关键字: %s", joined)
	}
}

// TestSplitExcelByKeyword_KeywordSplit 关键字拆分：关键字行前的内容留在原 sheet，关键字行后的内容进入 sheet@。
func TestSplitExcelByKeyword_KeywordSplit(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.xlsx")
	buildFixtureWorkbook(t, input)

	output := filepath.Join(dir, "out.xlsx")
	if err := SplitExcelByKeyword(input, output, "自定义脚本名字"); err != nil {
		t.Fatalf("SplitExcelByKeyword failed: %v", err)
	}

	before := joinRows(getSheetRows(t, output, "rule_table"))
	if !strings.Contains(before, "表名,rule_table") {
		t.Errorf("rule_table 应带表名表头行: %s", before)
	}
	if strings.Contains(before, "自定义脚本名字") {
		t.Errorf("rule_table 不应包含关键字行: %s", before)
	}
	if !strings.Contains(before, "北京市") {
		t.Errorf("rule_table 应包含关键字之前的行: %s", before)
	}
	if strings.Contains(before, "上海市") || strings.Contains(before, "广州市") {
		t.Errorf("rule_table 不应包含关键字之后的行: %s", before)
	}

	after := joinRows(getSheetRows(t, output, "rule_table"+utils.CUSTOM_RULE_SUFFIX))
	if strings.Contains(after, "自定义脚本名字") {
		t.Errorf("rule_table@ 不应包含关键字行: %s", after)
	}
	if strings.Contains(after, "北京市") {
		t.Errorf("rule_table@ 不应包含关键字之前的行: %s", after)
	}
	if !strings.Contains(after, "上海市") || !strings.Contains(after, "广州市") {
		t.Errorf("rule_table@ 应包含关键字之后的全部行: %s", after)
	}
}

// TestSplitExcelByKeyword_InputFileError 输入文件不存在时返回错误。
func TestSplitExcelByKeyword_InputFileError(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "not_exist.xlsx")
	if err := SplitExcelByKeyword(missing, t.TempDir(), "自定义脚本名字"); err == nil {
		t.Error("输入文件不存在应返回错误，实际返回 nil")
	}
}
