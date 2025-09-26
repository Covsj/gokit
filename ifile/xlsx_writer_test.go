package ifile

import (
	"path/filepath"
	"testing"

	"github.com/Covsj/gokit/ilog"
)

func TestXlsxWriter(t *testing.T) {

	testFileTempPath := filepath.Join("testdata", "xlsx_writer_test.xlsx")

	writer, err := NewXlsxWriter(testFileTempPath)
	if err != nil {
		ilog.Error("初始化XLSX Writer失败", "错误", err)
		return
	}

	headers := []string{"日期", "渠道", "曝光", "点击", "转化", "转化比", "支出", "总额", "备注"}
	if err = writer.WriteHeader("Sheet1", headers); err != nil {
		ilog.Error("WriteHeader失败", "错误", err)
		return
	}

	rows := [][]any{
		{"2025-09-24", "广告A", 12000, 845, 120, 0.1419, 1234.56, 9876.54, "正常"},
		{"2025-09-24", "广告B", 9800, 650, 88, 0.1354, 956.12, 7654.32, "正常"},
		{"2025-09-24", "自然流量", 15000, 1200, 240, 0.2000, 0.00, 4321.00, "良好"},
		{"2025-09-25", "广告A", 11000, 780, 105, 0.1346, 1111.11, 8765.43, "正常"},
		{"2025-09-25", "广告B", 10200, 720, 90, 0.1250, 999.99, 6543.21, "观察"},
		{"2025-09-25", "自然流量", 15800, 1320, 260, 0.1969, 0.00, 5000.00, "良好"},
		{"2025-09-26", "广告A", 12500, 900, 130, 0.1444, 1300.00, 9100.10, "正常"},
		{"2025-09-26", "广告B", 9900, 700, 95, 0.1357, 888.88, 7000.77, "正常"},
		{"2025-09-26", "自然流量", 16000, 1400, 280, 0.2000, 0.00, 5500.55, "优"},
		{"2025-09-27", "广告A", 13000, 950, 140, 0.1474, 1400.40, 9300.30, "正常"},
	}
	if err = writer.WriteData("Sheet1", rows); err != nil {
		ilog.Error("WriteData失败", "错误", err)
		return
	}
}
