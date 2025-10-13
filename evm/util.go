package evm

import (
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

// GenerateMnemonic 生成 12 个词助记词
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}
