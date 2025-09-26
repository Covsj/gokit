package ifile

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

const (
	streamTestFile     = "stream_test.xlsx"
	streamTestFileTemp = "stream_test_temp.xlsx"
)

func TestXlsxReader(t *testing.T) {
	// 使用绝对路径
	testFilePath := filepath.Join("testdata", streamTestFile)
	testFileTempPath := filepath.Join("testdata", streamTestFileTemp)

	// 清理测试文件
	defer func() {
		if err := os.Remove(testFilePath); err != nil && !os.IsNotExist(err) {
			t.Logf("删除测试文件失败: %v", err)
		}
		if err := os.Remove(testFileTempPath); err != nil && !os.IsNotExist(err) {
			t.Logf("删除临时测试文件失败: %v", err)
		}
	}()

	t.Run("基础操作测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 写入单元格
		err = handler.WriteCell("Sheet1", "A1", "测试数据")
		assert.NoError(t, err)

		// 读取单元格
		value, err := handler.ReadCell("Sheet1", "A1")
		assert.NoError(t, err)
		assert.Equal(t, "测试数据", value)

		// 保存文件
		err = handler.Save()
		assert.NoError(t, err)

		// 另存为
		err = handler.SaveAs(testFileTempPath)
		assert.NoError(t, err)
	})

	// 跨工作表读取测试
	t.Run("跨工作表读取测试", func(t *testing.T) {
		// 使用独立的测试文件避免数据干扰
		crossSheetTestFile := filepath.Join("testdata", "cross_sheet_test.xlsx")
		defer func() {
			if err := os.Remove(crossSheetTestFile); err != nil && !os.IsNotExist(err) {
				t.Logf("删除测试文件失败: %v", err)
			}
		}()

		handler, err := NewXlsxReader(crossSheetTestFile)
		assert.NoError(t, err)
		defer handler.Close()

		// 创建两个工作表并写入数据
		err = handler.CreateSheet("AllSheetsA")
		assert.NoError(t, err)
		err = handler.CreateSheet("AllSheetsB")
		assert.NoError(t, err)

		for i := 1; i <= 3; i++ {
			err = handler.WriteCell("AllSheetsA", fmt.Sprintf("A%d", i), fmt.Sprintf("A-%d", i))
			assert.NoError(t, err)
		}
		for i := 1; i <= 2; i++ {
			err = handler.WriteCell("AllSheetsB", fmt.Sprintf("A%d", i), fmt.Sprintf("B-%d", i))
			assert.NoError(t, err)
		}
		err = handler.Save()
		assert.NoError(t, err)

		// 一次性读取所有数据
		all, err := handler.ReadAllSheetsData()
		assert.NoError(t, err)
		assert.Contains(t, all, "AllSheetsA")
		assert.Contains(t, all, "AllSheetsB")
		assert.Len(t, all["AllSheetsA"], 3)
		assert.Len(t, all["AllSheetsB"], 2)
		assert.Equal(t, "A-1", all["AllSheetsA"][0][0])
		assert.Equal(t, "B-2", all["AllSheetsB"][1][0])

		// 流式读取所有数据
		rowCh, errCh := handler.ReadAllSheetsStream()
		total := 0
		aCount := 0
		bCount := 0
		for sr := range rowCh {
			total++
			if sr.Sheet == "AllSheetsA" {
				aCount++
			}
			if sr.Sheet == "AllSheetsB" {
				bCount++
			}
		}
		select {
		case err := <-errCh:
			assert.NoError(t, err)
		default:
		}
		assert.Equal(t, 5, total)
		assert.Equal(t, 3, aCount)
		assert.Equal(t, 2, bCount)
	})

	t.Run("流式读取行测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 写入测试数据
		for i := 1; i <= 5; i++ {
			err = handler.WriteCell("Sheet1", fmt.Sprintf("A%d", i), fmt.Sprintf("行%d", i))
			assert.NoError(t, err)
		}
		err = handler.Save()
		assert.NoError(t, err)

		// 流式读取行
		rowChan, errChan := handler.ReadRowStream("Sheet1")

		rows := make([][]string, 0)
		for row := range rowChan {
			rows = append(rows, row)
		}

		// 检查错误
		select {
		case err := <-errChan:
			assert.NoError(t, err)
		default:
		}

		assert.Len(t, rows, 5)
		assert.Equal(t, "行1", rows[0][0])
		assert.Equal(t, "行5", rows[4][0])
	})

	t.Run("流式读取列测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 写入测试数据
		for i := 1; i <= 5; i++ {
			err = handler.WriteCell("Sheet1", fmt.Sprintf("A%d", i), fmt.Sprintf("列A行%d", i))
			assert.NoError(t, err)
		}
		err = handler.Save()
		assert.NoError(t, err)

		// 流式读取列
		colChan, errChan := handler.ReadColStream("Sheet1", "A")

		values := make([]string, 0)
		for value := range colChan {
			values = append(values, value)
		}

		// 检查错误
		select {
		case err := <-errChan:
			assert.NoError(t, err)
		default:
		}

		assert.Len(t, values, 5)
		assert.Equal(t, "列A行1", values[0])
		assert.Equal(t, "列A行5", values[4])
	})

	t.Run("ProcessRows测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 创建新的工作表避免数据冲突
		err = handler.CreateSheet("ProcessRowsSheet")
		assert.NoError(t, err)

		// 写入测试数据
		for i := 1; i <= 3; i++ {
			err = handler.WriteCell("ProcessRowsSheet", fmt.Sprintf("A%d", i), fmt.Sprintf("数据%d", i))
			assert.NoError(t, err)
		}
		err = handler.Save()
		assert.NoError(t, err)

		// 使用ProcessRows处理数据
		processedRows := make([][]string, 0)
		err = handler.ProcessRows("ProcessRowsSheet", func(row []string) error {
			processedRows = append(processedRows, row)
			return nil
		})
		assert.NoError(t, err)

		assert.Len(t, processedRows, 3)
		assert.Equal(t, "数据1", processedRows[0][0])
		assert.Equal(t, "数据3", processedRows[2][0])
	})

	t.Run("ProcessRowsWithIndex测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 创建新的工作表避免数据冲突
		err = handler.CreateSheet("ProcessRowsWithIndexSheet")
		assert.NoError(t, err)

		// 写入测试数据
		for i := 1; i <= 3; i++ {
			err = handler.WriteCell("ProcessRowsWithIndexSheet", fmt.Sprintf("A%d", i), fmt.Sprintf("数据%d", i))
			assert.NoError(t, err)
		}
		err = handler.Save()
		assert.NoError(t, err)

		// 使用ProcessRowsWithIndex处理数据
		processedData := make(map[int][]string)
		err = handler.ProcessRowsWithIndex("ProcessRowsWithIndexSheet", func(rowIndex int, row []string) error {
			processedData[rowIndex] = row
			return nil
		})
		assert.NoError(t, err)

		assert.Len(t, processedData, 3)
		assert.Equal(t, "数据1", processedData[1][0])
		assert.Equal(t, "数据3", processedData[3][0])
	})

	t.Run("批量处理测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 写入测试数据
		for i := 1; i <= 10; i++ {
			err = handler.WriteCell("Sheet1", fmt.Sprintf("A%d", i), fmt.Sprintf("数据%d", i))
			assert.NoError(t, err)
		}
		err = handler.Save()
		assert.NoError(t, err)

		// 批量处理数据
		batches := make([][][]string, 0)
		err = handler.ProcessBatch("Sheet1", 3, func(batch [][]string) error {
			batches = append(batches, batch)
			return nil
		})
		assert.NoError(t, err)

		// 应该有4个批次：3+3+3+1
		assert.Len(t, batches, 4)
		assert.Len(t, batches[0], 3)
		assert.Len(t, batches[1], 3)
		assert.Len(t, batches[2], 3)
		assert.Len(t, batches[3], 1)
	})

	t.Run("过滤行测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 写入测试数据
		testData := []string{"重要", "普通", "重要", "普通", "重要"}
		for i, data := range testData {
			err = handler.WriteCell("Sheet1", fmt.Sprintf("A%d", i+1), data)
			assert.NoError(t, err)
		}
		err = handler.Save()
		assert.NoError(t, err)

		// 过滤包含"重要"的行
		filteredChan, errChan := handler.FilterRows("Sheet1", func(row []string) bool {
			return len(row) > 0 && row[0] == "重要"
		})

		filteredRows := make([][]string, 0)
		for row := range filteredChan {
			filteredRows = append(filteredRows, row)
		}

		// 检查错误
		select {
		case err := <-errChan:
			assert.NoError(t, err)
		default:
		}

		assert.Len(t, filteredRows, 3)
		for _, row := range filteredRows {
			assert.Equal(t, "重要", row[0])
		}
	})

	t.Run("复制工作表测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 写入测试数据
		for i := 1; i <= 3; i++ {
			err = handler.WriteCell("Sheet1", fmt.Sprintf("A%d", i), fmt.Sprintf("数据%d", i))
			assert.NoError(t, err)
		}
		err = handler.Save()
		assert.NoError(t, err)

		// 复制工作表
		err = handler.CopySheet("Sheet1", "Sheet1_Copy")
		assert.NoError(t, err)

		// 验证复制结果
		value, err := handler.ReadCell("Sheet1_Copy", "A1")
		assert.NoError(t, err)
		assert.Equal(t, "数据1", value)

		value, err = handler.ReadCell("Sheet1_Copy", "A3")
		assert.NoError(t, err)
		assert.Equal(t, "数据3", value)
	})

	t.Run("统计信息测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 创建新的工作表避免数据冲突
		err = handler.CreateSheet("StatsSheet")
		assert.NoError(t, err)

		// 写入测试数据
		for i := 1; i <= 5; i++ {
			err = handler.WriteCell("StatsSheet", fmt.Sprintf("A%d", i), fmt.Sprintf("数据%d", i))
			assert.NoError(t, err)
			err = handler.WriteCell("StatsSheet", fmt.Sprintf("B%d", i), fmt.Sprintf("值%d", i))
			assert.NoError(t, err)
		}
		err = handler.Save()
		assert.NoError(t, err)

		// 统计行数
		rowCount, err := handler.CountRows("StatsSheet")
		assert.NoError(t, err)
		assert.Equal(t, 5, rowCount)

		// 统计列数
		colCount, err := handler.CountCols("StatsSheet")
		assert.NoError(t, err)
		assert.Equal(t, 2, colCount)

		// 获取工作表信息
		rowCount2, colCount2, err := handler.GetSheetInfo("StatsSheet")
		assert.NoError(t, err)
		assert.Equal(t, 5, rowCount2)
		assert.Equal(t, 2, colCount2)
	})

	t.Run("工作表操作测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 创建新工作表
		err = handler.CreateSheet("TestSheet")
		assert.NoError(t, err)

		// 获取工作表列表
		sheets := handler.GetSheetList()
		assert.Contains(t, sheets, "TestSheet")

		// 删除工作表
		err = handler.DeleteSheet("TestSheet")
		assert.NoError(t, err)

		sheets = handler.GetSheetList()
		assert.NotContains(t, sheets, "TestSheet")
	})

	t.Run("错误处理测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 写入一些数据
		err = handler.WriteCell("Sheet1", "A1", "测试数据")
		assert.NoError(t, err)
		err = handler.Save()
		assert.NoError(t, err)

		// 测试读取不存在的工作表 - 使用ProcessRows方法测试
		err = handler.ProcessRows("不存在的工作表", func(row []string) error {
			return nil
		})
		assert.Error(t, err)

		// 测试读取不存在的列 - 使用ProcessRows方法测试
		err = handler.ProcessRows("Sheet1", func(row []string) error {
			return nil
		})
		// 这个应该成功，因为Sheet1存在
		assert.NoError(t, err)

		// 测试空路径保存
		emptyHandler := &XlsxReader{
			File: excelize.NewFile(),
		}
		err = emptyHandler.Save()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "文件路径未定义")
		emptyHandler.Close()
	})

	t.Run("流式写入测试", func(t *testing.T) {
		handler, err := NewXlsxReader(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 创建新的工作表避免数据冲突
		err = handler.CreateSheet("WriteStreamSheet")
		assert.NoError(t, err)

		// 准备写入数据
		rowChan := make(chan []string, 3)
		rowChan <- []string{"行1", "数据1"}
		rowChan <- []string{"行2", "数据2"}
		rowChan <- []string{"行3", "数据3"}
		close(rowChan)

		// 流式写入
		errChan := handler.WriteRowStream("WriteStreamSheet", rowChan)
		select {
		case err := <-errChan:
			assert.NoError(t, err)
		default:
		}

		// 保存文件
		err = handler.Save()
		assert.NoError(t, err)

		// 验证写入结果
		value, err := handler.ReadCell("WriteStreamSheet", "A1")
		assert.NoError(t, err)
		assert.Equal(t, "行1", value)

		value, err = handler.ReadCell("WriteStreamSheet", "B3")
		assert.NoError(t, err)
		assert.Equal(t, "数据3", value)
	})
}

// TestStreamProcessor 测试流式处理器接口
func TestStreamProcessor(t *testing.T) {
	testFilePath := filepath.Join("testdata", "stream_processor_test.xlsx")

	defer func() {
		if err := os.Remove(testFilePath); err != nil && !os.IsNotExist(err) {
			t.Logf("删除测试文件失败: %v", err)
		}
	}()

	handler, err := NewXlsxReader(testFilePath)
	assert.NoError(t, err)
	defer handler.Close()

	// 写入测试数据
	for i := 1; i <= 5; i++ {
		err = handler.WriteCell("Sheet1", fmt.Sprintf("A%d", i), fmt.Sprintf("数据%d", i))
		assert.NoError(t, err)
	}
	err = handler.Save()
	assert.NoError(t, err)

	// 创建自定义处理器
	processor := &TestProcessor{
		processedRows: make(map[int][]string),
	}

	// 使用流式处理器处理数据
	err = handler.ProcessWithStream("Sheet1", processor)
	assert.NoError(t, err)

	// 验证处理结果
	assert.Len(t, processor.processedRows, 5)
	assert.Equal(t, "数据1", processor.processedRows[1][0])
	assert.Equal(t, "数据5", processor.processedRows[5][0])
	assert.True(t, processor.completed)
}

// TestProcessor 测试用的流式处理器
type TestProcessor struct {
	processedRows map[int][]string
	completed     bool
}

func (tp *TestProcessor) ProcessRow(rowIndex int, row []string) error {
	tp.processedRows[rowIndex] = row
	return nil
}

func (tp *TestProcessor) OnComplete() error {
	tp.completed = true
	return nil
}
