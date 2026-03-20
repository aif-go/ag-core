package other

import (
	"ag-core/tool/cmd/new-gen-db/utils"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// SplitExcelByKeyword 将一个excel中的所有sheet按照关键字分割成两个sheet写入到新的excel中
// 新sheet的名字: 关键字之前的sheet和原sheet保持一致，关键字之后的sheet名字为: 原sheet名字@custom
// 例如: 原sheet名字为: user_info, 关键字为: user, 那么新sheet的名字为: user_info@custom
// 如果原sheet没有包含关键字, 那么不进行分割，直接将原sheet复制到新的excel中
// 新sheet的格式规则:
// 1. 所有单元格的字体为宋体，大小为11
// 2. 所有的单元格都加边框，除非单元格的内容为空
// 3. 如果该行的单元格的内容是汉字，则该行的内容加粗
// filePath: 原Excel文件路径
// outputPath: 新Excel文件路径
// keyword: 分割关键字
func SplitExcelByKeyword(filePath, outputPath, keyword string) error {
	// 如果outputPath是目录或没有文件名，则使用默认文件名：原文件名_日期
	// 判断是否为目录：以路径分隔符结尾或没有扩展名
	// outputPathBase := filepath.Base(outputPath)
	if filepath.Ext(outputPath) == "" {
		// outputPath是目录或没有扩展名，需要拼接文件名
		// 获取原文件名（不含扩展名）
		baseName := filepath.Base(filePath)
		ext := filepath.Ext(baseName)
		nameWithoutExt := strings.TrimSuffix(baseName, ext)
		// 生成日期字符串
		dateStr := time.Now().Format("20060102")
		// 构建默认输出路径
		outputPath = filepath.Join(outputPath, nameWithoutExt+"_"+dateStr+ext)
	}

	// 打开原Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return fmt.Errorf("打开Excel文件失败: %w", err)
	}

	// 创建新的Excel文件
	newF := excelize.NewFile()

	// 创建基础样式：宋体11号字，加边框
	baseStyle, err := newF.NewStyle(`{
		"font":{"family":"宋体","size":11},
		"border":[{"type":"left","color":"000000","style":1},{"type":"top","color":"000000","style":1},{"type":"bottom","color":"000000","style":1},{"type":"right","color":"000000","style":1}]
	}`)
	if err != nil {
		return fmt.Errorf("创建基础样式失败: %w", err)
	}

	// 创建加粗样式：宋体11号字，加粗，加边框
	boldStyle, err := newF.NewStyle(`{
		"font":{"family":"宋体","size":11,"bold":true},
		"border":[{"type":"left","color":"000000","style":1},{"type":"top","color":"000000","style":1},{"type":"bottom","color":"000000","style":1},{"type":"right","color":"000000","style":1}]
	}`)
	if err != nil {
		return fmt.Errorf("创建加粗样式失败: %w", err)
	}

	// 遍历所有工作表
	for _, sheetName := range f.GetSheetMap() {

		if len(sheetName)>30{
			fmt.Printf("sheet:%s name length > 30, please check\n",sheetName)
		}

		// 获取所有行
		rows := f.GetRows(sheetName)

		// 查找关键字所在的行
		keywordRowIndex := -1
		for i, row := range rows {
			for _, cell := range row {
				if strings.Contains(strings.TrimSpace(cell), keyword) {
					keywordRowIndex = i
					break
				}
			}
			if keywordRowIndex != -1 {
				break
			}
		}

		// 如果没有找到关键字，直接复制整个sheet
		if keywordRowIndex == -1 {
			// 创建新sheet
			newSheetName := sheetName
			// // 如果sheet名已存在，添加后缀
			// if newF.GetSheetIndex(newSheetName) != -1 {
			// 	newSheetName = fmt.Sprintf("%s_%d", sheetName, len(newF.GetSheetMap())+1)
			// }
			// newF.NewSheet(newSheetName)

			// 复制所有行并应用样式
			for i, row := range rows {
				// 最初开始写入的时候，第一行增加表名的操作
				if i == 0 {
					newF.SetCellStyle(newSheetName,"A1","A1",boldStyle)
					newF.SetCellValue(newSheetName,"A1","表名")
					newF.SetCellStyle(newSheetName,"B1","B1",boldStyle)
					newF.SetCellValue(newSheetName,"B1",newSheetName)
				}
				// 检查该行是否有值
				rowHasValue := rowHasAnyValue(row)
				// 检查该行是否包含汉字
				rowHasChinese := rowContainsChinese(row)
				styleID := baseStyle
				if rowHasChinese {
					styleID = boldStyle
				}

				// 如果该行有值，则整行都应用样式
				if rowHasValue {
					// 获取该行的最大列数
					maxCol := len(row)
					if maxCol == 0 {
						maxCol = 1
					}
					// 应用样式到该行的所有单元格
					for j := 0; j < maxCol; j++ {
						cellName := getCellName(j, i+2)
						newF.SetCellStyle(newSheetName, cellName, cellName, styleID)
					}
				}

				// 写入单元格值
				for j, cell := range row {
					cellName := getCellName(j, i+2)
					newF.SetCellValue(newSheetName, cellName, cell)
				}
			}

			// 自动调整列宽
			adjustColumnWidth(newF, newSheetName, rows)
			continue
		}

		// 找到关键字，分割sheet
		// 第一部分：关键字之前的行（不包含关键字所在行）
		beforeKeywordSheetName := sheetName
		if newF.GetSheetIndex(beforeKeywordSheetName) != -1 {
			beforeKeywordSheetName = fmt.Sprintf("%s", sheetName)
		}
		newF.NewSheet(beforeKeywordSheetName)

		// 第二部分：关键字之后的行（不包含关键字所在行）
		afterKeywordSheetName := fmt.Sprintf("%s%s", sheetName, utils.CUSTOM_RULE_SUFFIX)
		newF.NewSheet(afterKeywordSheetName)

		// 写入第一部分并应用样式
		for i := 0; i < keywordRowIndex; i++ {
			if i >= len(rows) {
				break
			}
			// 最初开始写入的时候，第一行增加表名的操作
			if i == 0 {
				newF.SetCellStyle(beforeKeywordSheetName,"A1","A1",boldStyle)
				newF.SetCellValue(beforeKeywordSheetName,"A1","表名")
				newF.SetCellStyle(beforeKeywordSheetName,"B1","B1",boldStyle)
				newF.SetCellValue(beforeKeywordSheetName,"B1",sheetName)
			}
			row := rows[i]
			// 检查该行是否有值
			rowHasValue := rowHasAnyValue(row)
			// 检查该行是否包含汉字
			rowHasChinese := rowContainsChinese(row)
			styleID := baseStyle
			if rowHasChinese {
				styleID = boldStyle
			}

			// 如果该行有值，则整行都应用样式
			if rowHasValue {
				// 获取该行的最大列数
				maxCol := len(row)
				if maxCol == 0 {
					maxCol = 1
				}
				// 应用样式到该行的所有单元格
				for j := 0; j < maxCol; j++ {
					// 第一行留给表名使用
					cellName := getCellName(j, i+2)
					newF.SetCellStyle(beforeKeywordSheetName, cellName, cellName, styleID)
				}
			}

			// 写入单元格值
			for j := 0; j < len(row); j++ {
				// 第一行留给表名使用
				cellName := getCellName(j, i+2)
				newF.SetCellValue(beforeKeywordSheetName, cellName, row[j])
			}
		}

		// 写入第二部分并应用样式
		afterRowIndex := 0
		after:=true
		if len(rows)-(keywordRowIndex + 2) == 0{
			after=false
			// 删除新建的无用sheet
			newF.DeleteSheet(afterKeywordSheetName)
		}
		if after {
			for i := keywordRowIndex + 1; i < len(rows); i++ {
				row := rows[i]
				// 检查该行是否有值
				rowHasValue := rowHasAnyValue(row)
				// 检查该行是否包含汉字
				rowHasChinese := rowContainsChinese(row)
				styleID := baseStyle
				if rowHasChinese {
					styleID = boldStyle
				}

				// 如果该行有值，则整行都应用样式
				if rowHasValue {
					// 获取该行的最大列数
					maxCol := len(row)
					if maxCol == 0 {
						maxCol = 1
					}
					// 应用样式到该行的所有单元格
					for j := 0; j < maxCol; j++ {
						cellName := getCellName(j, afterRowIndex+1)
						newF.SetCellStyle(afterKeywordSheetName, cellName, cellName, styleID)
					}
				}

				// 写入单元格值
				for j := 0; j < len(row); j++ {
					cellName := getCellName(j, afterRowIndex+1)
					newF.SetCellValue(afterKeywordSheetName, cellName, row[j])
				}
				afterRowIndex++
			}
		}
		// 自动调整列宽
		adjustColumnWidth(newF, beforeKeywordSheetName, rows[:keywordRowIndex])
		if after {	
			adjustColumnWidth(newF, afterKeywordSheetName, rows[keywordRowIndex+1:])
		}
	}

	// 删除默认的Sheet1
	newF.DeleteSheet("Sheet1")

	// 保存新Excel文件
	if err := newF.SaveAs(outputPath); err != nil {
		return fmt.Errorf("保存Excel文件失败: %w", err)
	}

	return nil
}

// getCellName 根据列号和行号生成单元格名称（例如：A1, B2）
func getCellName(col, row int) string {
	colName := excelize.ToAlphaString(col)
	return colName + strconv.Itoa(row)
}

// rowHasAnyValue 检查一行中是否有任何非空值
func rowHasAnyValue(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return true
		}
	}
	return false
}

// rowContainsChinese 检查一行中是否包含汉字
func rowContainsChinese(row []string) bool {
	// for _, cell := range row {
	if containsChinese(row[0]) {
		return true
	}
	// }
	return false
}

// containsChinese 检查字符串中是否包含汉字
func containsChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// adjustColumnWidth 根据单元格值长度自动调整列宽，最大40
func adjustColumnWidth(f *excelize.File, sheetName string, rows [][]string) {
	// 计算每列的最大宽度
	maxWidths := make(map[int]int)
	for _, row := range rows {
		for j, cell := range row {
			// 计算单元格内容的宽度（汉字算2个字符宽度）
			width := calculateCellWidth(cell)
			if width > maxWidths[j] {
				maxWidths[j] = width
			}
		}
	}

	// 设置列宽
	for col, width := range maxWidths {
		// 转换为Excel列宽（字符宽度除以1.5左右，因为Excel的列宽单位是字符宽度）
		// 最大宽度为40
		colWidth := float64(width) / 1.5
		// 设置最小宽度为10，确保至少有明显的宽度
		if colWidth < 10 {
			colWidth = 15
		}
		// 最大宽度为40
		if colWidth > 40 {
			colWidth = 50
		}
		colName := excelize.ToAlphaString(col)
		f.SetColWidth(sheetName, colName, colName, colWidth)
	}
}

// calculateCellWidth 计算单元格内容的宽度（汉字算2个字符宽度）
func calculateCellWidth(cell string) int {
	width := 0
	for _, r := range cell {
		if unicode.Is(unicode.Han, r) {
			width += 2 // 汉字算2个字符宽度
		} else {
			width += 1
		}
	}
	return width
}
