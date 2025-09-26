package ifile

import "os"

// exists 检查文件是否存在
func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
