package ifile

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
)

// CsvHandler CSV文件处理器
type CsvHandler struct {
	filePath string
	file     *os.File
	reader   *csv.Reader
	writer   *csv.Writer
	data     [][]string // 缓存数据
}

// CsvOption CSV处理器的配置选项
type CsvOption struct {
	// SkipEmptyRows 是否跳过空行
	// true: 读取时会忽略所有空行（即所有字段都为空的行）
	// false: 保留所有行，包括空行
	// 默认值: true
	SkipEmptyRows bool

	// Comma 指定CSV文件的分隔符
	// 常用值:
	// ',' - 标准CSV格式（默认）
	// ';' - 分号分隔的CSV
	// '\t' - TSV格式（制表符分隔）
	// 默认值: ','
	Comma rune

	// TrimLeadingSpace 是否移除字段的前导空格
	// true: 自动移除每个字段开头的空格
	// false: 保留所有空格
	// 例如：当为true时，" abc " 会变成 "abc "
	// 默认值: true
	TrimLeadingSpace bool

	// LazyQuotes 是否使用宽松的引号规则
	// true: 允许字段中包含非转义的引号
	// false: 严格遵守CSV引号规则
	// 例如：当为true时，可以处理 a"b,c 这样的非标准格式
	// 默认值: true
	LazyQuotes bool
}

// DefaultCsvOption 默认CSV配置
func DefaultCsvOption() *CsvOption {
	return &CsvOption{
		SkipEmptyRows:    true,
		Comma:            ',',
		TrimLeadingSpace: true,
		LazyQuotes:       true,
	}
}

// NewCsvHandler 创建新的CSV处理器
func NewCsvHandler(filePath string) (*CsvHandler, error) {
	var f *os.File
	var err error

	if exists(filePath) {
		f, err = os.OpenFile(filePath, os.O_RDWR, 0644)
	} else {
		f, err = os.Create(filePath)
	}

	if err != nil {
		return nil, fmt.Errorf("打开CSV文件失败: %w", err)
	}

	handler := &CsvHandler{
		filePath: filePath,
		file:     f,
		reader:   csv.NewReader(f),
		writer:   csv.NewWriter(f),
	}

	// 配置CSV reader
	handler.reader.FieldsPerRecord = -1 // 允许每行的字段数不同

	// 读取所有数据到内存
	handler.data, err = handler.reader.ReadAll()
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("读取CSV数据失败: %w", err)
	}

	return handler, nil
}

// isEmptyRow 检查一行是否为空
func isEmptyRow(record []string) bool {
	for _, field := range record {
		if strings.TrimSpace(field) != "" {
			return false
		}
	}
	return true
}

// ReadRow 读取指定行
func (h *CsvHandler) ReadRow(row int) ([]string, error) {
	if row <= 0 {
		return nil, fmt.Errorf("行号 %d 超出范围", row)
	}

	if row > len(h.data) {
		return nil, fmt.Errorf("行号 %d 超出范围", row)
	}

	return h.data[row-1], nil
}

// WriteRow 写入整行数据
func (h *CsvHandler) WriteRow(row int, values []string) error {
	if row <= 0 {
		return fmt.Errorf("行号 %d 超出范围", row)
	}

	// 如果需要，扩展数据切片
	for len(h.data) < row {
		h.data = append(h.data, make([]string, 0))
	}

	h.data[row-1] = values
	return nil
}

// ReadCol 读取指定列的所有数据
func (h *CsvHandler) ReadCol(col int) ([]string, error) {
	if col <= 0 {
		return nil, fmt.Errorf("列号 %d 超出范围", col)
	}

	var colData []string
	for _, row := range h.data {
		if col <= len(row) {
			colData = append(colData, row[col-1])
		} else {
			colData = append(colData, "") // 空单元格
		}
	}
	return colData, nil
}

// WriteCol 写入整列数据
func (h *CsvHandler) WriteCol(col int, values []string) error {
	if col <= 0 {
		return fmt.Errorf("列号 %d 超出范围", col)
	}

	// 确保有足够的行
	for i := len(h.data); i < len(values); i++ {
		h.data = append(h.data, make([]string, 0))
	}

	// 确保每行有足够的列
	for i, value := range values {
		for len(h.data[i]) < col {
			h.data[i] = append(h.data[i], "")
		}
		h.data[i][col-1] = value
	}
	return nil
}

// Save 保存文件
func (h *CsvHandler) Save() error {
	// 清空文件
	h.file.Truncate(0)
	h.file.Seek(0, 0)

	// 写入所有数据
	h.writer.WriteAll(h.data)
	if err := h.writer.Error(); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}
	return nil
}

// SaveAs 另存为新文件
func (h *CsvHandler) SaveAs(newPath string) error {
	f, err := os.Create(newPath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	writer.WriteAll(h.data)
	if err := writer.Error(); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

// Close 关闭文件
func (h *CsvHandler) Close() error {
	return h.file.Close()
}

// GetData 获取所有数据
func (h *CsvHandler) GetData() [][]string {
	return h.data
}

// SetData 设置所有数据
func (h *CsvHandler) SetData(data [][]string) {
	h.data = data
}

// InsertRow 插入行
func (h *CsvHandler) InsertRow(row int, values []string) error {
	if row <= 0 {
		return fmt.Errorf("行号 %d 超出范围", row)
	}

	// 确保有足够的空间
	if row > len(h.data)+1 {
		return fmt.Errorf("行号 %d 超出范围", row)
	}

	// 在指定位置插入新行
	h.data = append(h.data[:row-1], append([][]string{values}, h.data[row-1:]...)...)
	return nil
}

// DeleteRow 删除行
func (h *CsvHandler) DeleteRow(row int) error {
	if row <= 0 || row > len(h.data) {
		return fmt.Errorf("行号 %d 超出范围", row)
	}

	h.data = append(h.data[:row-1], h.data[row:]...)
	return nil
}

// NewCsvHandlerWithOption 使用自定义选项创建CSV处理器
func NewCsvHandlerWithOption(filePath string, opt *CsvOption) (*CsvHandler, error) {
	if opt == nil {
		opt = DefaultCsvOption()
	}

	var f *os.File
	var err error

	if exists(filePath) {
		f, err = os.OpenFile(filePath, os.O_RDWR, 0644)
	} else {
		f, err = os.Create(filePath)
	}

	if err != nil {
		return nil, fmt.Errorf("打开CSV文件失败: %w", err)
	}

	handler := &CsvHandler{
		filePath: filePath,
		file:     f,
		reader:   csv.NewReader(f),
		writer:   csv.NewWriter(f),
	}

	// 应用配置
	handler.reader.Comma = opt.Comma
	handler.reader.LazyQuotes = opt.LazyQuotes
	handler.reader.TrimLeadingSpace = opt.TrimLeadingSpace
	handler.reader.FieldsPerRecord = -1

	// 读取所有数据到内存
	var rows [][]string
	for {
		record, err := handler.reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("读取CSV数据失败: %w", err)
		}

		// 处理行中的空单元格
		processedRecord := make([]string, len(record))
		for i, field := range record {
			if strings.TrimSpace(field) == "" {
				processedRecord[i] = "" // 保留空单元格
			} else {
				processedRecord[i] = field
			}
		}

		// 根据配置决定是否跳过空行
		if !opt.SkipEmptyRows || !isEmptyRow(processedRecord) {
			rows = append(rows, processedRecord)
		}
	}
	handler.data = rows

	return handler, nil
}
