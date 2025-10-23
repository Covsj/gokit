package ievm

import (
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/tyler-smith/go-bip39"
)

// ==================== 地址验证工具 ====================

// IsValidAddress 验证地址格式
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

// IsZeroAddress 检查是否为零地址
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

// ==================== 金额转换工具 ====================

// FormatETH 格式化金额（Wei 转 ETH）
func FormatETH(wei *big.Int) string {
	if wei == nil {
		return "0"
	}

	// 1 ETH = 10^18 Wei
	eth := new(big.Float).SetInt(wei)
	eth.Quo(eth, big.NewFloat(1e18))

	return eth.Text('f', 18)
}

// ParseETH 解析金额（ETH 转 Wei）
func ParseETH(eth string) (*big.Int, error) {
	if eth == "" {
		return big.NewInt(0), nil
	}

	// 移除可能的空格
	eth = strings.TrimSpace(eth)

	// 解析为浮点数
	ethFloat, err := strconv.ParseFloat(eth, 64)
	if err != nil {
		return nil, fmt.Errorf("解析 ETH 金额失败: %w", err)
	}

	// 转换为 Wei
	weiFloat := ethFloat * 1e18
	wei := big.NewInt(int64(weiFloat))

	return wei, nil
}

// FormatGwei 格式化金额（Wei 转 Gwei）
func FormatGwei(wei *big.Int) string {
	if wei == nil {
		return "0"
	}

	// 1 Gwei = 10^9 Wei
	gwei := new(big.Float).SetInt(wei)
	gwei.Quo(gwei, big.NewFloat(1e9))

	return gwei.Text('f', 9)
}

// ParseGwei 解析金额（Gwei 转 Wei）
func ParseGwei(gwei string) (*big.Int, error) {
	if gwei == "" {
		return big.NewInt(0), nil
	}

	// 移除可能的空格
	gwei = strings.TrimSpace(gwei)

	// 解析为浮点数
	gweiFloat, err := strconv.ParseFloat(gwei, 64)
	if err != nil {
		return nil, fmt.Errorf("解析 Gwei 金额失败: %w", err)
	}

	// 转换为 Wei
	weiFloat := gweiFloat * 1e9
	wei := big.NewInt(int64(weiFloat))

	return wei, nil
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
func ToDecimalsFloat(s string, decimals int64) (*big.Int, error) {
	if s == "" {
		return big.NewInt(0), nil
	}

	// 手工十进制解析以避免引入 big.Rat 额外舍入差异
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
	if i := strings.IndexByte(s, '.'); i >= 0 {
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

// ==================== 数据转换工具 ====================

// BytesToHex 字节数组转十六进制字符串
func BytesToHex(data []byte) string {
	return hexutil.Encode(data)
}

// HexToBytes 十六进制字符串转字节数组
func HexToBytes(hex string) ([]byte, error) {
	return hexutil.Decode(hex)
}

// BigIntToHex 大整数转十六进制字符串
func BigIntToHex(value *big.Int) string {
	if value == nil {
		return "0x0"
	}
	return hexutil.EncodeBig(value)
}

// HexToBigInt 十六进制字符串转大整数
func HexToBigInt(hex string) (*big.Int, error) {
	return hexutil.DecodeBig(hex)
}

// ==================== 字符串工具 ====================

// TruncateAddress 截断地址显示（显示前6位和后4位）
func TruncateAddress(address string) string {
	if !IsValidAddress(address) {
		return address
	}

	if len(address) < 10 {
		return address
	}

	return address[:6] + "..." + address[len(address)-4:]
}

// TruncateHash 截断哈希显示（显示前8位和后8位）
func TruncateHash(hash string) string {
	if len(hash) < 16 {
		return hash
	}

	return hash[:8] + "..." + hash[len(hash)-8:]
}

// ==================== 助记词工具 ====================

// GenerateMnemonic 生成 12 个词助记词
func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", err
	}
	return bip39.NewMnemonic(entropy)
}

// ==================== 验证工具 ====================

// ValidateAmount 验证金额是否有效
func ValidateAmount(amount *big.Int) error {
	if amount == nil {
		return fmt.Errorf("金额不能为空")
	}

	if amount.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("金额不能为负数")
	}

	return nil
}

// ValidateGasLimit 验证 Gas 限制是否有效
func ValidateGasLimit(gasLimit uint64) error {
	if gasLimit == 0 {
		return fmt.Errorf("Gas 限制不能为0")
	}

	if gasLimit > 30000000 { // 以太坊区块 Gas 限制
		return fmt.Errorf("Gas 限制过大")
	}

	return nil
}

// ValidateGasPrice 验证 Gas 价格是否有效
func ValidateGasPrice(gasPrice *big.Int) error {
	if gasPrice == nil {
		return fmt.Errorf("Gas 价格不能为空")
	}

	if gasPrice.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("Gas 价格必须大于0")
	}

	return nil
}

// ==================== 比较工具 ====================

// CompareAddresses 比较两个地址是否相等
func CompareAddresses(addr1, addr2 string) bool {
	if !IsValidAddress(addr1) || !IsValidAddress(addr2) {
		return false
	}

	return strings.EqualFold(addr1, addr2)
}

// CompareHashes 比较两个哈希是否相等
func CompareHashes(hash1, hash2 string) bool {
	return strings.EqualFold(hash1, hash2)
}
