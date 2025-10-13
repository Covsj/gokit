package ievm

import (
	"fmt"
	"math/big"
	"reflect"
	"regexp"

	"github.com/ethereum/go-ethereum/common"
	"github.com/tyler-smith/go-bip39"
)

// ValidateAddress 粗略校验 0x 开头且 42 长度
// IsValidAddress validate hex address
func IsValidAddress(iaddress interface{}) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	switch v := iaddress.(type) {
	case string:
		return re.MatchString(v)
	case common.Address:
		return re.MatchString(v.Hex())
	default:
		return false
	}
}

// IsZeroAddress validate if it's a 0 address
func IsZeroAddress(iaddress interface{}) bool {
	var address common.Address
	switch v := iaddress.(type) {
	case string:
		address = common.HexToAddress(v)
	case common.Address:
		address = v
	default:
		return false
	}

	zeroAddressBytes := common.FromHex("0x0000000000000000000000000000000000000000")
	addressBytes := address.Bytes()
	return reflect.DeepEqual(addressBytes, zeroAddressBytes)
}

// FromDecimals 将整数金额按小数位转换为可读浮点
func FromDecimals(v *big.Int, decimals int64) *big.Float {
	if v == nil {
		return big.NewFloat(0)
	}
	base := new(big.Float).SetInt(v)
	den := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil))
	return new(big.Float).Quo(base, den)
}

// ToDecimals 将可读值按小数位转换为整数金额
func ToDecimals(v uint64, decimals int64) *big.Int {
	res := new(big.Int).Mul(new(big.Int).SetUint64(v), new(big.Int).Exp(big.NewInt(10), big.NewInt(decimals), nil))
	return res
}

// ToDecimalsFloat 将任意精度十进制字符串按小数位转换为整数金额
// 例如："1.23"、decimals=6 -> 1230000
func ToDecimalsFloat(s string, decimals int64) (*big.Int, error) {
	if s == "" {
		return big.NewInt(0), nil
	}
	// 手工十进制解析以避免引入 big.Rat 额外舍入差异
	// 允许可选的正负号
	neg := false
	switch s[0] {
	case '-':
		neg = true
		s = s[1:]
	case '+':
		s = s[1:]
	}
	intPart := s
	fracPart := ""
	if i := indexByte(s, '.'); i >= 0 {
		intPart = s[:i]
		fracPart = s[i+1:]
	}
	// 去除前导零
	for len(intPart) > 1 && intPart[0] == '0' {
		intPart = intPart[1:]
	}
	// 处理小数位：截断或右补零到 decimals 位
	if int64(len(fracPart)) > decimals {
		fracPart = fracPart[:decimals]
	} else {
		for int64(len(fracPart)) < decimals {
			fracPart += "0"
		}
	}
	merged := intPart + fracPart
	if merged == "" {
		merged = "0"
	}
	// 解析为大整数
	out := new(big.Int)
	_, ok := out.SetString(merged, 10)
	if !ok {
		return nil, fmt.Errorf("非法十进制: %s", s)
	}
	if neg {
		out.Neg(out)
	}
	return out, nil
}

// indexByte 返回字节在字符串中的索引，不存在返回 -1
func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// GenerateMnemonic 生成 12 个词助记词
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}
