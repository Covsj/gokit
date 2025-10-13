package evm

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// Pack 将 ABI 方法与参数编码为 data
func Pack(abiJSON, method string, args ...interface{}) ([]byte, error) {
	a, err := abi.JSON(stringsNewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("解析 ABI 失败: %w", err)
	}
	data, err := a.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("ABI 打包失败: %w", err)
	}
	return data, nil
}

// Unpack 将返回数据解码为指定方法的输出
func Unpack(abiJSON, method string, output []byte) ([]interface{}, error) {
	a, err := abi.JSON(stringsNewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("解析 ABI 失败: %w", err)
	}
	res, err := a.Unpack(method, output)
	if err != nil {
		return nil, fmt.Errorf("ABI 解码失败: %w", err)
	}
	return res, nil
}

// 轻量字符串 reader，避免引入 strings 依赖扩散
func stringsNewReader(s string) *stringsReader { return &stringsReader{s: s, i: 0} }

type stringsReader struct {
	s string
	i int
}

func (r *stringsReader) Read(p []byte) (int, error) {
	if r.i >= len(r.s) {
		return 0, ioEOF{}
	}
	n := copy(p, r.s[r.i:])
	r.i += n
	return n, nil
}

type ioEOF struct{}

func (ioEOF) Error() string { return "EOF" }
