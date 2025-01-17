package ifile

import (
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

// XlsxHandler xlsx文件处理器
type XlsxHandler struct {
	filePath string
	file     *excelize.File
}

// NewXlsxHandler 创建新的xlsx处理器
func NewXlsxHandler(filePath string) (*XlsxHandler, error) {
	var f *excelize.File
	var err error

	// 检查文件是否存在
	if exists(filePath) {
		f, err = excelize.OpenFile(filePath)
	} else {
		f = excelize.NewFile()
	}

	if err != nil {
		return nil, fmt.Errorf("打开xlsx文件失败: %w", err)
	}

	return &XlsxHandler{
		filePath: filePath,
		file:     f,
	}, nil
}

// ReadCell 读取指定单元格的值
func (h *XlsxHandler) ReadCell(sheet, cell string) (string, error) {
	value, err := h.file.GetCellValue(sheet, cell)
	if err != nil {
		return "", fmt.Errorf("读取单元格失败: %w", err)
	}
	return value, nil
}

// WriteCell 写入指定单元格
func (h *XlsxHandler) WriteCell(sheet, cell, value string) error {
	err := h.file.SetCellValue(sheet, cell, value)
	if err != nil {
		return fmt.Errorf("写入单元格失败: %w", err)
	}
	return nil
}

// ReadRow 读取指定行
func (h *XlsxHandler) ReadRow(sheet string, row int) ([]string, error) {
	if row <= 0 {
		return nil, fmt.Errorf("行号 %d 超出范围", row)
	}

	rows, err := h.file.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("读取行失败: %w", err)
	}

	if row > len(rows) {
		return nil, fmt.Errorf("行号 %d 超出范围", row)
	}

	return rows[row-1], nil
}

// WriteRow 写入整行数据
func (h *XlsxHandler) WriteRow(sheet string, row int, values []string) error {
	for i, value := range values {
		cell, err := excelize.CoordinatesToCellName(i+1, row)
		if err != nil {
			return fmt.Errorf("生成单元格坐标失败: %w", err)
		}

		if err := h.WriteCell(sheet, cell, value); err != nil {
			return err
		}
	}
	return nil
}

// Save 保存文件
func (h *XlsxHandler) Save() error {
	if h.filePath == "" {
		return fmt.Errorf("文件路径未定义")
	}
	if err := h.file.SaveAs(h.filePath); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}
	return nil
}

// SaveAs 另存为新文件
func (h *XlsxHandler) SaveAs(newPath string) error {
	if err := h.file.SaveAs(newPath); err != nil {
		return fmt.Errorf("另存为失败: %w", err)
	}
	return nil
}

// Close 关闭文件
func (h *XlsxHandler) Close() error {
	return h.file.Close()
}

// CreateSheet 创建新的工作表
func (h *XlsxHandler) CreateSheet(name string) error {
	_, err := h.file.NewSheet(name)
	if err != nil {
		return fmt.Errorf("创建工作表失败: %w", err)
	}
	return nil
}

// DeleteSheet 删除工作表
func (h *XlsxHandler) DeleteSheet(name string) error {
	err := h.file.DeleteSheet(name)
	if err != nil {
		return fmt.Errorf("删除工作表失败: %w", err)
	}
	return nil
}

// GetSheetList 获取所有工作表名称
func (h *XlsxHandler) GetSheetList() []string {
	return h.file.GetSheetList()
}

// exists 检查文件是否存在
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadCol 读取指定列的所有数据
func (h *XlsxHandler) ReadCol(sheet string, col string) ([]string, error) {
	colIndex, err := excelize.ColumnNameToNumber(col)
	if err != nil {
		return nil, fmt.Errorf("列名转换失败: %w", err)
	}

	rows, err := h.file.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("读取列失败: %w", err)
	}

	// 检查是否有任何数据
	if len(rows) == 0 {
		return nil, fmt.Errorf("工作表为空")
	}

	// 检查列是否超出范围
	maxCol := 0
	for _, row := range rows {
		if len(row) > maxCol {
			maxCol = len(row)
		}
	}
	if colIndex > maxCol {
		return nil, fmt.Errorf("列号 %s 超出范围", col)
	}

	var colData []string
	for _, row := range rows {
		if len(row) >= colIndex {
			colData = append(colData, row[colIndex-1])
		} else {
			colData = append(colData, "") // 空单元格
		}
	}
	return colData, nil
}

// WriteCol 写入整列数据
func (h *XlsxHandler) WriteCol(sheet, col string, values []string) error {
	for i, value := range values {
		cell := fmt.Sprintf("%s%d", col, i+1)
		if err := h.WriteCell(sheet, cell, value); err != nil {
			return fmt.Errorf("写入列数据失败: %w", err)
		}
	}
	return nil
}

// MergeCells 合并单元格
func (h *XlsxHandler) MergeCells(sheet, startCell, endCell string) error {
	err := h.file.MergeCell(sheet, startCell, endCell)
	if err != nil {
		return fmt.Errorf("合并单元格失败: %w", err)
	}
	return nil
}

// SetCellStyle 设置单元格样式
func (h *XlsxHandler) SetCellStyle(sheet, startCell, endCell string, style *excelize.Style) error {
	styleID, err := h.file.NewStyle(style)
	if err != nil {
		return fmt.Errorf("创建样式失败: %w", err)
	}

	err = h.file.SetCellStyle(sheet, startCell, endCell, styleID)
	if err != nil {
		return fmt.Errorf("设置样式失败: %w", err)
	}
	return nil
}

// InsertRow 插入行
func (h *XlsxHandler) InsertRow(sheet string, row int) error {
	err := h.file.InsertRows(sheet, row, 1)
	if err != nil {
		return fmt.Errorf("插入行失败: %w", err)
	}
	return nil
}

// DeleteRow 删除行
func (h *XlsxHandler) DeleteRow(sheet string, row int) error {
	err := h.file.RemoveRow(sheet, row)
	if err != nil {
		return fmt.Errorf("删除行失败: %w", err)
	}
	return nil
}

// SetColWidth 设置列宽
func (h *XlsxHandler) SetColWidth(sheet, startCol, endCol string, width float64) error {
	err := h.file.SetColWidth(sheet, startCol, endCol, width)
	if err != nil {
		return fmt.Errorf("设置列宽失败: %w", err)
	}
	return nil
}

// SetRowHeight 设置行高
func (h *XlsxHandler) SetRowHeight(sheet string, row int, height float64) error {
	err := h.file.SetRowHeight(sheet, row, height)
	if err != nil {
		return fmt.Errorf("设置行高失败: %w", err)
	}
	return nil
}

// AddPicture 插入图片
func (h *XlsxHandler) AddPicture(sheet, cell, picturePath string, opts *excelize.GraphicOptions) error {
	err := h.file.AddPicture(sheet, cell, picturePath, opts)
	if err != nil {
		return fmt.Errorf("插入图片失败: %w", err)
	}
	return nil
}

// GetSheetData 获取整个工作表的数据
func (h *XlsxHandler) GetSheetData(sheet string) ([][]string, error) {
	rows, err := h.file.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("读取工作表数据失败: %w", err)
	}
	return rows, nil
}

// SetSheetData 写入整个工作表的数据
func (h *XlsxHandler) SetSheetData(sheet string, data [][]string) error {
	for rowIndex, row := range data {
		if err := h.WriteRow(sheet, rowIndex+1, row); err != nil {
			return fmt.Errorf("写入工作表数据失败: %w", err)
		}
	}
	return nil
}

// XlsxData 表示整个Excel文件的数据结构
type XlsxData struct {
	Sheets map[string][][]string // key为sheet名称，value为sheet数据
}

// ReadAll 读取整个xlsx文件的所有数据
func (h *XlsxHandler) ReadAll() (*XlsxData, error) {
	sheets := h.GetSheetList()
	data := &XlsxData{
		Sheets: make(map[string][][]string),
	}

	for _, sheet := range sheets {
		sheetData, err := h.GetSheetData(sheet)
		if err != nil {
			return nil, fmt.Errorf("读取工作表 %s 失败: %w", sheet, err)
		}
		data.Sheets[sheet] = sheetData
	}

	return data, nil
}

// WriteAll 写入整个xlsx文件的数据
func (h *XlsxHandler) WriteAll(data *XlsxData) error {
	// 先删除所有现有的工作表
	for _, sheet := range h.GetSheetList() {
		if err := h.DeleteSheet(sheet); err != nil {
			return fmt.Errorf("删除工作表 %s 失败: %w", sheet, err)
		}
	}

	// 写入新的数据
	for sheetName, sheetData := range data.Sheets {
		// 创建新的工作表
		if err := h.CreateSheet(sheetName); err != nil {
			return fmt.Errorf("创建工作表 %s 失败: %w", sheetName, err)
		}

		// 写入工作表数据
		if err := h.SetSheetData(sheetName, sheetData); err != nil {
			return fmt.Errorf("写入工作表 %s 数据失败: %w", sheetName, err)
		}
	}

	return nil
}

// GetSheetRange 获取工作表的有效范围
func (h *XlsxHandler) GetSheetRange(sheet string) (string, error) {
	rows, err := h.GetSheetData(sheet)
	if err != nil {
		return "", fmt.Errorf("获取工作表数据失败: %w", err)
	}

	if len(rows) == 0 {
		return "A1", nil
	}

	// 找到最大行数和最大列数
	maxRow := len(rows)
	maxCol := 0
	for _, row := range rows {
		if len(row) > maxCol {
			maxCol = len(row)
		}
	}

	// 转换最大列数为列名
	endCol, err := excelize.ColumnNumberToName(maxCol)
	if err != nil {
		return "", fmt.Errorf("列数转换失败: %w", err)
	}

	return fmt.Sprintf("%s%d", endCol, maxRow), nil
}
