package ievm

import (
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// Solidity 与 Go 类型对应关系（go-ethereum/abi 编解码常用约定）
// - address           -> common.Address
// - bool              -> bool
// - string            -> string
// - bytes             -> []byte
// - bytes1..bytes32   -> [1]byte..[32]byte（结构体字段可用定长数组；一般解包也可视作 []byte）
// - uint / uint256    -> *big.Int  （Solidity 的 uint 为 uint256 别名）
// - int / int256      -> *big.Int
// - uint8/16/32/64    -> uint8/uint16/uint32/uint64
// - int8/16/32/64     -> int8/int16/int32/int64
// - fixed bytes/ints  -> 同上相应 Go 基本/数组类型
// - address[]         -> []common.Address
// - T[]               -> []T （T 按上述映射）
// - tuple             -> struct（按字段顺序映射），或 []interface{}（使用 Pack/Unpack 原始返回）

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
