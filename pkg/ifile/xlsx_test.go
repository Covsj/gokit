package ifile

import (
	"os"
	"path/filepath"
	"testing"

	log "github.com/Covsj/gokit/pkg/ilog"
	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

const (
	testFile     = "test.xlsx"
	testFileTemp = "test_temp.xlsx"
)

func init() {
	// 确保测试文件目录存在
	if err := os.MkdirAll("testdata", 0755); err != nil {
		log.ErrorF("创建测试目录失败: %v", err)
	}
}

func TestXlsxHandler(t *testing.T) {
	// 使用绝对路径
	testFilePath := filepath.Join("testdata", testFile)
	testFileTempPath := filepath.Join("testdata", testFileTemp)

	// 清理测试文件
	defer func() {
		// if err := os.Remove(testFilePath); err != nil && !os.IsNotExist(err) {
		// 	log.ErrorF("删除测试文件失败: %v", err)
		// }
		// if err := os.Remove(testFileTempPath); err != nil && !os.IsNotExist(err) {
		// 	log.ErrorF("删除临时测试文件失败: %v", err)
		// }
	}()

	t.Run("基础操作测试", func(t *testing.T) {
		handler, err := NewXlsxHandler(testFilePath)
		if err != nil {
			log.ErrorF("创建xlsx处理器失败: %v", err)
		}
		assert.NoError(t, err)
		defer handler.Close()

		// 写入单元格
		err = handler.WriteCell("Sheet1", "A1", "测试数据")
		if err != nil {
			log.ErrorF("写入单元格失败: %v", err)
		}
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

	t.Run("行列操作测试", func(t *testing.T) {
		handler, err := NewXlsxHandler(testFilePath)
		if err != nil {
			log.ErrorF("创建xlsx处理器失败: %v", err)
		}
		assert.NoError(t, err)
		defer handler.Close()

		// 写入行数据
		rowData := []string{"行1列1", "行1列2", "行1列3"}
		err = handler.WriteRow("Sheet1", 1, rowData)
		assert.NoError(t, err)

		// 读取行数据
		readRow, err := handler.ReadRow("Sheet1", 1)
		assert.NoError(t, err)
		assert.Equal(t, rowData, readRow)

		// 写入列数据
		colData := []string{"列A行1", "列A行2", "列A行3"}
		err = handler.WriteCol("Sheet1", "A", colData)
		assert.NoError(t, err)

		// 读取列数据
		readCol, err := handler.ReadCol("Sheet1", "A")
		assert.NoError(t, err)
		assert.Equal(t, colData, readCol)
	})

	t.Run("工作表操作测试", func(t *testing.T) {
		handler, err := NewXlsxHandler(testFilePath)
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

	t.Run("样式操作测试", func(t *testing.T) {
		handler, err := NewXlsxHandler(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 设置单元格样式
		style := &excelize.Style{
			Font: &excelize.Font{
				Bold:  true,
				Size:  12,
				Color: "#FF0000",
			},
		}
		err = handler.SetCellStyle("Sheet1", "A1", "A1", style)
		assert.NoError(t, err)

		// 合并单元格
		err = handler.MergeCells("Sheet1", "B1", "C1")
		assert.NoError(t, err)
	})

	t.Run("整表操作测试", func(t *testing.T) {
		handler, err := NewXlsxHandler(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 准备测试数据
		testData := &XlsxData{
			Sheets: map[string][][]string{
				"Sheet1": {
					{"表头1", "表头2", "表头3"},
					{"数据1", "数据2", "数据3"},
				},
				"Sheet2": {
					{"名称", "年龄", "性别"},
					{"张三", "20", "男"},
				},
			},
		}

		// 写入整个文件数据
		err = handler.WriteAll(testData)
		assert.NoError(t, err)

		// 读取整个文件数据
		readData, err := handler.ReadAll()
		assert.NoError(t, err)
		assert.Equal(t, testData.Sheets, readData.Sheets)

		// 获取工作表范围
		sheetRange, err := handler.GetSheetRange("Sheet1")
		assert.NoError(t, err)
		assert.Equal(t, "C2", sheetRange)
	})

	t.Run("行列设置测试", func(t *testing.T) {
		handler, err := NewXlsxHandler(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 插入行
		err = handler.InsertRow("Sheet1", 1)
		assert.NoError(t, err)

		// 删除行
		err = handler.DeleteRow("Sheet1", 1)
		assert.NoError(t, err)

		// 设置列宽
		err = handler.SetColWidth("Sheet1", "A", "A", 20)
		assert.NoError(t, err)

		// 设置行高
		err = handler.SetRowHeight("Sheet1", 1, 30)
		assert.NoError(t, err)
	})

	t.Run("图片操作测试", func(t *testing.T) {
		handler, err := NewXlsxHandler(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 创建临时图片文件
		tmpImgFile := "test.png"
		defer os.Remove(tmpImgFile)

		err = os.WriteFile(tmpImgFile, []byte("fake image data"), 0644)
		assert.NoError(t, err)

		// 插入图片
		err = handler.AddPicture("Sheet1", "A1", tmpImgFile, nil)
		assert.Error(t, err) // 预期会失败，因为是假的图片数据
	})

	t.Run("错误处理测试", func(t *testing.T) {
		// 创建一个新的handler用于测试
		handler, err := NewXlsxHandler(testFilePath)
		assert.NoError(t, err)
		defer handler.Close()

		// 先写入一些数据，确保有内容可以测试
		err = handler.WriteCell("Sheet1", "A1", "测试数据")
		assert.NoError(t, err)
		err = handler.Save()
		assert.NoError(t, err)

		// 读取不存在的单元格
		_, err = handler.ReadCell("不存在的工作表", "A1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sheet")

		// 读取超出范围的行
		_, err = handler.ReadRow("Sheet1", 1000)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "超出范围")

		// 测试空路径保存
		emptyHandler := &XlsxHandler{
			file: excelize.NewFile(),
			// 不设置 filePath，保持为空字符串
		}
		err = emptyHandler.Save()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "文件路径未定义")
		emptyHandler.Close() // 确保关闭文件

		// 测试读取不存在的行
		_, err = handler.ReadRow("Sheet1", -1)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "超出范围")

		// 测试读取不存在的列
		_, err = handler.ReadCol("Sheet1", "XFD") // Excel 最大列
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "超出范围")

		// 测试合并无效的单元格
		err = handler.MergeCells("Sheet1", "A1", "invalid_cell")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "合并单元格失败")
	})
}

func TestExists(t *testing.T) {
	testFilePath := filepath.Join("testdata", testFile)

	// 测试文件不存在的情况
	assert.False(t, exists("不存在的文件.xlsx"))

	// 创建临时文件测试存在的情况
	f := excelize.NewFile()
	err := f.SaveAs(testFilePath)
	if err != nil {
		log.ErrorF("保存测试文件失败: %v", err)
	}
	assert.NoError(t, err)
	defer func() {
		if err := os.Remove(testFilePath); err != nil && !os.IsNotExist(err) {
			log.ErrorF("删除测试文件失败: %v", err)
		}
	}()

	assert.True(t, exists(testFilePath))
}
