package ifile

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Handler 文件处理器结构体
type Handler struct {
	file *os.File
}

// InitHandler 初始化文件处理器
func InitHandler(fileName string) *Handler {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err.Error())
	}
	return &Handler{file: file}
}

// WriteLine 写入一行数据
func (h *Handler) WriteLine(str string) {
	if !strings.HasSuffix(str, "\n") {
		str += "\n"
	}
	_, _ = h.file.WriteString(str)
}

// Close 关闭文件
func (h *Handler) Close() {
	_ = h.file.Close()
}

// ReadFile 读取整个文件内容
func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

// WriteFile 写入内容到文件
func WriteFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

// ReadLines 读取文件的所有行
func ReadLines(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// WriteLines 写入多行到文件
func WriteLines(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

// AppendLines 追加多行到文件末尾
func AppendLines(filename string, lines []string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

// ReadJSON 从文件读取JSON并解析到结构体
func ReadJSON(filename string, v interface{}) error {
	data, err := ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

// WriteJSON 将结构体转换为JSON并写入文件
func WriteJSON(filename string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return WriteFile(filename, data)
}

// Exists 检查文件或目录是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsDir 检查路径是否为目录
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// EnsureDir 确保目录存在，如果不存在则创建
func EnsureDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

// GetFileSize 获取文件大小
func GetFileSize(filename string) (int64, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// GetFileExt 获取文件扩展名
func GetFileExt(filename string) string {
	return filepath.Ext(filename)
}

// GetFileName 获取文件名（不含扩展名）
func GetFileName(filename string) string {
	return strings.TrimSuffix(filepath.Base(filename), filepath.Ext(filename))
}

// DeleteFile 删除文件
func DeleteFile(filename string) error {
	return os.Remove(filename)
}

// DeleteDir 删除目录及其内容
func DeleteDir(dir string) error {
	return os.RemoveAll(dir)
}

// ListDir 列出目录下的所有文件和子目录
func ListDir(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path != dir {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// ListFiles 列出目录下的所有文件（不包括目录）
func ListFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
