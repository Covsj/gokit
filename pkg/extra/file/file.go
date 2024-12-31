package file

import (
	"os"
	"strings"
)

type Handler struct {
	file *os.File
}

func InitHandler(fileName string) *Handler {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err.Error())
	}
	return &Handler{file: file}
}

func (h *Handler) WriteLine(str string) {
	if !strings.HasSuffix(str, "\n") {
		str += "\n"
	}
	_, _ = h.file.WriteString(str)
}

func (h *Handler) Close() {
	_ = h.file.Close()
}
