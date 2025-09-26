package ifile

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// XlsxReader 流式处理xlsx文件的处理器
// 适用于超大文件，逐行处理，避免内存溢出
type XlsxReader struct {
	FilePath string
	File     *excelize.File
}

// NewXlsxReader 创建流式xlsx处理器
func NewXlsxReader(filePath string) (*XlsxReader, error) {
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

	return &XlsxReader{
		FilePath: filePath,
		File:     f,
	}, nil
}

// ReadRowStream 流式读取行数据
// 返回一个channel，逐行发送数据
func (h *XlsxReader) ReadRowStream(sheet string) (<-chan []string, <-chan error) {
	rowChan := make(chan []string, 100) // 缓冲100行
	errChan := make(chan error, 1)

	go func() {
		defer close(rowChan)
		defer close(errChan)

		// 使用流式读取
		rows, err := h.File.Rows(sheet)
		if err != nil {
			errChan <- fmt.Errorf("获取行迭代器失败: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			row, err := rows.Columns()
			if err != nil {
				errChan <- fmt.Errorf("读取行数据失败: %w", err)
				return
			}
			rowChan <- row
		}

		if err := rows.Error(); err != nil {
			errChan <- fmt.Errorf("行迭代器错误: %w", err)
		}
	}()

	return rowChan, errChan
}

// ReadColStream 流式读取列数据
// 返回一个channel，逐行发送指定列的数据
func (h *XlsxReader) ReadColStream(sheet, col string) (<-chan string, <-chan error) {
	colChan := make(chan string, 100) // 缓冲100个值
	errChan := make(chan error, 1)

	go func() {
		defer close(colChan)
		defer close(errChan)

		colIndex, err := excelize.ColumnNameToNumber(col)
		if err != nil {
			errChan <- fmt.Errorf("列名转换失败: %w", err)
			return
		}

		// 使用流式读取
		rows, err := h.File.Rows(sheet)
		if err != nil {
			errChan <- fmt.Errorf("获取行迭代器失败: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			row, err := rows.Columns()
			if err != nil {
				errChan <- fmt.Errorf("读取行数据失败: %w", err)
				return
			}

			// 获取指定列的值
			if len(row) >= colIndex {
				colChan <- row[colIndex-1]
			} else {
				colChan <- "" // 空单元格
			}
		}

		if err := rows.Error(); err != nil {
			errChan <- fmt.Errorf("行迭代器错误: %w", err)
		}
	}()

	return colChan, errChan
}

// ProcessRows 处理行数据的通用方法
// 接受一个处理函数，对每一行进行处理
func (h *XlsxReader) ProcessRows(sheet string, processor func([]string) error) error {
	rows, err := h.File.Rows(sheet)
	if err != nil {
		return fmt.Errorf("获取行迭代器失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("读取行数据失败: %w", err)
		}

		if err := processor(row); err != nil {
			return fmt.Errorf("处理行数据失败: %w", err)
		}
	}

	if err := rows.Error(); err != nil {
		return fmt.Errorf("行迭代器错误: %w", err)
	}

	return nil
}

// ProcessRowsWithIndex 处理行数据的通用方法（带行号）
func (h *XlsxReader) ProcessRowsWithIndex(sheet string, processor func(int, []string) error) error {
	rows, err := h.File.Rows(sheet)
	if err != nil {
		return fmt.Errorf("获取行迭代器失败: %w", err)
	}
	defer rows.Close()

	rowIndex := 1
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("读取行数据失败: %w", err)
		}

		if err := processor(rowIndex, row); err != nil {
			return fmt.Errorf("处理行数据失败: %w", err)
		}
		rowIndex++
	}

	if err := rows.Error(); err != nil {
		return fmt.Errorf("行迭代器错误: %w", err)
	}

	return nil
}

// WriteRowStream 流式写入行数据
// 接受一个channel，逐行写入数据
func (h *XlsxReader) WriteRowStream(sheet string, rowChan <-chan []string) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)

		rowIndex := 1
		for row := range rowChan {
			for colIndex, value := range row {
				cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex)
				if err != nil {
					errChan <- fmt.Errorf("生成单元格坐标失败: %w", err)
					return
				}
				if err := h.File.SetCellValue(sheet, cell, value); err != nil {
					errChan <- fmt.Errorf("写入单元格失败: %w", err)
					return
				}
			}
			rowIndex++
		}
	}()

	return errChan
}

// CopySheet 复制工作表（流式处理）
func (h *XlsxReader) CopySheet(sourceSheet, targetSheet string) error {
	// 创建目标工作表
	_, err := h.File.NewSheet(targetSheet)
	if err != nil {
		return fmt.Errorf("创建目标工作表失败: %w", err)
	}

	// 流式复制数据
	return h.ProcessRowsWithIndex(sourceSheet, func(rowIndex int, row []string) error {
		for colIndex, value := range row {
			cell, err := excelize.CoordinatesToCellName(colIndex+1, rowIndex)
			if err != nil {
				return fmt.Errorf("生成单元格坐标失败: %w", err)
			}
			if err := h.File.SetCellValue(targetSheet, cell, value); err != nil {
				return fmt.Errorf("写入单元格失败: %w", err)
			}
		}
		return nil
	})
}

// FilterRows 过滤行数据（流式处理）
// 返回一个channel，包含过滤后的行数据
func (h *XlsxReader) FilterRows(sheet string, filter func([]string) bool) (<-chan []string, <-chan error) {
	filteredChan := make(chan []string, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(filteredChan)
		defer close(errChan)

		err := h.ProcessRows(sheet, func(row []string) error {
			if filter(row) {
				select {
				case filteredChan <- row:
				default:
					// 如果channel满了，等待一下
					filteredChan <- row
				}
			}
			return nil
		})

		if err != nil {
			errChan <- err
		}
	}()

	return filteredChan, errChan
}

// CountRows 统计行数（流式处理，不加载全部数据到内存）
func (h *XlsxReader) CountRows(sheet string) (int, error) {
	count := 0
	err := h.ProcessRows(sheet, func(row []string) error {
		count++
		return nil
	})
	return count, err
}

// CountCols 统计最大列数（流式处理）
func (h *XlsxReader) CountCols(sheet string) (int, error) {
	maxCols := 0
	err := h.ProcessRows(sheet, func(row []string) error {
		if len(row) > maxCols {
			maxCols = len(row)
		}
		return nil
	})
	return maxCols, err
}

// GetSheetInfo 获取工作表信息（流式处理）
func (h *XlsxReader) GetSheetInfo(sheet string) (rowCount, colCount int, err error) {
	rowCount, err = h.CountRows(sheet)
	if err != nil {
		return 0, 0, err
	}

	colCount, err = h.CountCols(sheet)
	if err != nil {
		return 0, 0, err
	}

	return rowCount, colCount, nil
}

// Save 保存文件
func (h *XlsxReader) Save() error {
	if h.FilePath == "" {
		return fmt.Errorf("文件路径未定义")
	}
	if err := h.File.SaveAs(h.FilePath); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}
	return nil
}

// SaveAs 另存为新文件
func (h *XlsxReader) SaveAs(newPath string) error {
	if err := h.File.SaveAs(newPath); err != nil {
		return fmt.Errorf("另存为失败: %w", err)
	}
	return nil
}

// Close 关闭文件
func (h *XlsxReader) Close() error {
	return h.File.Close()
}

// GetSheetList 获取所有工作表名称
func (h *XlsxReader) GetSheetList() []string {
	return h.File.GetSheetList()
}

// CreateSheet 创建新的工作表
func (h *XlsxReader) CreateSheet(name string) error {
	_, err := h.File.NewSheet(name)
	if err != nil {
		return fmt.Errorf("创建工作表失败: %w", err)
	}
	return nil
}

// DeleteSheet 删除工作表
func (h *XlsxReader) DeleteSheet(name string) error {
	err := h.File.DeleteSheet(name)
	if err != nil {
		return fmt.Errorf("删除工作表失败: %w", err)
	}
	return nil
}

// ReadCell 读取指定单元格的值
func (h *XlsxReader) ReadCell(sheet, cell string) (string, error) {
	value, err := h.File.GetCellValue(sheet, cell)
	if err != nil {
		return "", fmt.Errorf("读取单元格失败: %w", err)
	}
	return value, nil
}

// WriteCell 写入指定单元格
func (h *XlsxReader) WriteCell(sheet, cell, value string) error {
	err := h.File.SetCellValue(sheet, cell, value)
	if err != nil {
		return fmt.Errorf("写入单元格失败: %w", err)
	}
	return nil
}

// StreamProcessor 流式处理器接口
type StreamProcessor interface {
	ProcessRow(rowIndex int, row []string) error
	OnComplete() error
}

// ProcessWithStream 使用流式处理器处理数据
func (h *XlsxReader) ProcessWithStream(sheet string, processor StreamProcessor) error {
	rows, err := h.File.Rows(sheet)
	if err != nil {
		return fmt.Errorf("获取行迭代器失败: %w", err)
	}
	defer rows.Close()

	rowIndex := 1
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("读取行数据失败: %w", err)
		}

		if err := processor.ProcessRow(rowIndex, row); err != nil {
			return fmt.Errorf("处理行数据失败: %w", err)
		}
		rowIndex++
	}

	if err := rows.Error(); err != nil {
		return fmt.Errorf("行迭代器错误: %w", err)
	}

	// 处理完成后的清理工作
	if err := processor.OnComplete(); err != nil {
		return fmt.Errorf("完成处理失败: %w", err)
	}

	return nil
}

// BatchProcessor 批量处理器
type BatchProcessor struct {
	BatchSize int
	Processor func([][]string) error
	batch     [][]string
}

// NewBatchProcessor 创建批量处理器
func NewBatchProcessor(batchSize int, processor func([][]string) error) *BatchProcessor {
	return &BatchProcessor{
		BatchSize: batchSize,
		Processor: processor,
		batch:     make([][]string, 0, batchSize),
	}
}

// ProcessRow 处理单行数据
func (bp *BatchProcessor) ProcessRow(rowIndex int, row []string) error {
	bp.batch = append(bp.batch, row)

	if len(bp.batch) >= bp.BatchSize {
		if err := bp.Processor(bp.batch); err != nil {
			return err
		}
		bp.batch = bp.batch[:0] // 清空batch
	}

	return nil
}

// OnComplete 完成处理
func (bp *BatchProcessor) OnComplete() error {
	if len(bp.batch) > 0 {
		return bp.Processor(bp.batch)
	}
	return nil
}

// ProcessBatch 批量处理数据
func (h *XlsxReader) ProcessBatch(sheet string, batchSize int, processor func([][]string) error) error {
	batchProcessor := NewBatchProcessor(batchSize, processor)
	return h.ProcessWithStream(sheet, batchProcessor)
}

// SheetRow 表示带有工作表名的行数据
type SheetRow struct {
	Sheet    string
	RowIndex int
	Row      []string
}

// ReadAllSheetsData 读取所有工作表的所有行数据（一次性返回，不建议超大文件使用）
func (h *XlsxReader) ReadAllSheetsData() (map[string][][]string, error) {
	result := make(map[string][][]string)
	sheets := h.GetSheetList()

	for _, sheet := range sheets {
		rows, err := h.File.Rows(sheet)
		if err != nil {
			return nil, fmt.Errorf("获取行迭代器失败: %w", err)
		}
		var data [][]string
		for rows.Next() {
			row, err := rows.Columns()
			if err != nil {
				rows.Close()
				return nil, fmt.Errorf("读取行数据失败: %w", err)
			}
			// 复制一份，避免底层复用切片导致的数据覆盖
			copied := make([]string, len(row))
			copy(copied, row)
			data = append(data, copied)
		}
		if err := rows.Error(); err != nil {
			rows.Close()
			return nil, fmt.Errorf("行迭代器错误: %w", err)
		}
		rows.Close()
		result[sheet] = data
	}
	return result, nil
}

// ReadAllSheetsStream 跨工作表流式读取所有行
// 返回一个包含 SheetRow 的通道，以及一个错误通道。
func (h *XlsxReader) ReadAllSheetsStream() (<-chan SheetRow, <-chan error) {
	out := make(chan SheetRow, 128)
	errCh := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errCh)

		sheets := h.GetSheetList()
		for _, sheet := range sheets {
			rows, err := h.File.Rows(sheet)
			if err != nil {
				errCh <- fmt.Errorf("获取行迭代器失败: %w", err)
				return
			}
			rowIndex := 1
			for rows.Next() {
				row, err := rows.Columns()
				if err != nil {
					rows.Close()
					errCh <- fmt.Errorf("读取行数据失败: %w", err)
					return
				}
				// 复制一份，避免底层复用切片导致的数据覆盖
				copied := make([]string, len(row))
				copy(copied, row)
				out <- SheetRow{Sheet: sheet, RowIndex: rowIndex, Row: copied}
				rowIndex++
			}
			if err := rows.Error(); err != nil {
				rows.Close()
				errCh <- fmt.Errorf("行迭代器错误: %w", err)
				return
			}
			rows.Close()
		}
	}()

	return out, errCh
}
