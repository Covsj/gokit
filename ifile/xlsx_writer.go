package ifile

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type XlsxWriter struct {
	F            *excelize.File
	Headers      []string
	Path         string
	ColumnWidths map[int]int
	XSplit       int
	YSplit       int
}

func NewXlsxWriter(path string) (*XlsxWriter, error) {

	f := excelize.NewFile()
	if err := f.SaveAs(path); err != nil {
		return nil, err
	}
	f, err := excelize.OpenFile(path)
	if err != nil {
		return nil, err
	}
	return &XlsxWriter{F: f,
		Path:         path,
		ColumnWidths: map[int]int{},
		Headers:      make([]string, 0)}, nil
}

func (w *XlsxWriter) WriteHeader(sheet string, header []string) error {
	if sheet == "" {
		sheet = "Sheet1"
	}
	for i, h := range header {
		col, err := excelize.ColumnNumberToName(i + 1)
		if err != nil {
			return err
		}
		err = w.F.SetCellValue(sheet, fmt.Sprintf("%s1", col), h)
		if err != nil {
			return err
		}
		w.ColumnWidths[i] = len(h)
	}
	w.setHeaderStyle(w.F, sheet, len(header))
	w.Headers = header
	return nil
}

func (w *XlsxWriter) WriteData(sheet string, data [][]any) error {
	if sheet == "" {
		sheet = "Sheet1"
	}

	colCount := 0
	for rowIndex, values := range data {
		rowIdx := rowIndex + 2
		colCount = len(values)
		for j, v := range values {
			col, _ := excelize.ColumnNumberToName(j + 1)
			w.F.SetCellValue(sheet, fmt.Sprintf("%s%d", col, rowIdx), v)
			w.ColumnWidths[j] = max(w.ColumnWidths[j], len(fmt.Sprintf("%v", v)))
		}
	}
	w.setDataStyle(w.F, sheet, len(data), colCount)
	w.setRatioColumnFormat(w.F, sheet, w.Headers, len(data))

	if err := w.F.Save(); err != nil {
		return err
	}
	return w.Close()
}
func (w *XlsxWriter) Close() error {
	return w.F.Close()
}

// 设置表头样式
func (w *XlsxWriter) setHeaderStyle(f *excelize.File, sheet string, colCount int) {
	// 设置表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 11,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
			WrapText:   true, // 启用自动换行
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#E6F3FF"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err == nil {
		// 应用表头样式
		f.SetCellStyle(sheet, "A1", w.getColumnName(colCount)+"1", headerStyle)

		// 设置表头行高为32
		f.SetRowHeight(sheet, 1, 32)

		if w.XSplit == 0 {
			w.XSplit = 2
		}
		if w.YSplit == 0 {
			w.YSplit = 1
		}

		f.SetPanes(sheet, &excelize.Panes{
			Freeze:      true,
			Split:       false,
			XSplit:      w.XSplit, // 固定列
			YSplit:      w.YSplit, // 固定行
			TopLeftCell: "C2",
			ActivePane:  "topRight",
		})
	}
}

// 设置数据样式
func (w *XlsxWriter) setDataStyle(f *excelize.File, sheet string, rowCount, colCount int) {
	// 设置数据区域样式
	dataStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err == nil {
		// 应用数据样式
		f.SetCellStyle(sheet, "A2", w.getColumnName(colCount)+fmt.Sprintf("%d", rowCount+1), dataStyle)
	}

	// 设置列宽
	w.setColumnWidths(f, sheet, colCount)
}

// 设置列宽
func (w *XlsxWriter) setColumnWidths(f *excelize.File, sheet string, colCount int) {
	// 设置列宽
	for i := 0; i < colCount; i++ {
		colName := w.getColumnName(i + 1)
		f.SetColWidth(sheet, colName, colName, (float64(w.ColumnWidths[i])*0.8)+5)
	}
}

// 设置比例类列的显示格式为保留两位小数
func (w *XlsxWriter) setRatioColumnFormat(f *excelize.File, sheet string, headers []string, rowCount int) {

	colCount := len(headers)
	ratioCols, percentCols := []int{}, []int{}
	for i, v := range headers {
		if strings.Contains(v, "比") {
			percentCols = append(percentCols, i+1)
		} else if strings.Contains(v, "支出") {
			ratioCols = append(ratioCols, i+1)
		} else if strings.Contains(v, "总额") {
			ratioCols = append(ratioCols, i+1)
		}
	}

	lastRow := rowCount + 1 // 数据从第2行开始
	for _, colIdx := range ratioCols {
		colName := w.getColumnName(colIdx)
		// 获取原有样式
		styleID, _ := f.GetCellStyle(sheet, colName+"2")
		style, _ := f.GetStyle(styleID)
		style.NumFmt = 2 // 0.00
		newStyleID, _ := f.NewStyle(style)
		_ = f.SetCellStyle(sheet, colName+"2", colName+fmt.Sprintf("%d", lastRow), newStyleID)
	}
	for _, colIdx := range percentCols {
		colName := w.getColumnName(colIdx)
		styleID, _ := f.GetCellStyle(sheet, colName+"2")
		style, _ := f.GetStyle(styleID)
		style.NumFmt = 10 // 0.00%
		newStyleID, _ := f.NewStyle(style)
		_ = f.SetCellStyle(sheet, colName+"2", colName+fmt.Sprintf("%d", lastRow), newStyleID)
	}

	w.setHeaderStyle(f, sheet, colCount)
}

// 获取列名
func (w *XlsxWriter) getColumnName(colNum int) string {
	colName, _ := excelize.ColumnNumberToName(colNum)
	return colName
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
