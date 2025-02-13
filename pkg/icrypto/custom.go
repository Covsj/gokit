package icrypto

import (
	"strings"
)

// CaesarUtil 提供了凯撒加密和解密的功能
type CaesarUtil struct {
	OffsetFunc      func(str string) int          // 用于计算偏移量的函数
	StableIndexFunc func(str string) map[int]bool // 用于确定稳定索引的函数
	CustomCharset   string                        // 自定义字符集
	DupCountFunc    func(str string) int          // 用于计算重复次数的函数
}

// NewCaesarUtil 创建一个新的 CaesarUtil 实例
func NewCaesarUtil(offsetFunc func(str string) int, stableIndexFunc func(str string) map[int]bool, customCharset string, dupCountFunc func(str string) int) *CaesarUtil {
	if offsetFunc == nil {
		offsetFunc = func(str string) int {
			return len(str) % 3
		}
	}
	if stableIndexFunc == nil {
		stableIndexFunc = func(str string) map[int]bool {
			res := map[int]bool{}
			for i := range str {
				if i != 0 && i != 5 && i%3 == 0 {
					res[i] = true
				}
			}
			return res
		}
	}
	if customCharset == "" {
		customCharset = "ijkl?mn0uFG+HIJKLMNO=123VWXYZ%abc4567~89AB#CDEqr^stPQRST*Udefghopv!wx$yz"
	}
	if dupCountFunc == nil {
		dupCountFunc = func(str string) int {
			return len(str) % 6
		}
	}
	return &CaesarUtil{
		OffsetFunc:      offsetFunc,
		StableIndexFunc: stableIndexFunc,
		CustomCharset:   customCharset,
		DupCountFunc:    dupCountFunc,
	}
}

// Encode 使用凯撒加密对字符串进行编码
func (c *CaesarUtil) Encode(str string, encrypt bool) string {
	if !encrypt {
		str = reverseString(str)
	}

	var offset int
	if c.OffsetFunc != nil {
		offset = c.OffsetFunc(str)
	}
	if offset == 0 {
		offset = 13
	}

	customCharset := c.CustomCharset

	// 使用 DES 加密生成新的字符集
	aes := EncryptDES(
		BaseEncode(BaseEncode(customCharset, "32").ToString(), "64").ToString(),
		"ECB", "PKCS7",
		HashGenerator(customCharset, "Md5").ToRawBytes(), nil).ToRawString()
	customCharset = uniqueChars(customCharset + aes)

	stableIndex := map[int]bool{}
	if c.StableIndexFunc != nil {
		stableIndex = c.StableIndexFunc(str)
	}

	mapping := createMapping(customCharset, offset, encrypt)

	var result strings.Builder
	for i, char := range str {
		if stableIndex[i] {
			result.WriteRune(char)
			continue
		}
		if mappedChar, ok := mapping[string(char)]; ok {
			result.WriteString(mappedChar)
		} else {
			result.WriteRune(char)
		}
	}

	if encrypt {
		return reverseString(result.String())
	}

	return result.String()
}

// EncodeMultipleTimes 对字符串进行多次凯撒加密
func (c *CaesarUtil) EncodeMultipleTimes(str string, encrypt bool) string {
	result := str
	var dupCount int
	if c.DupCountFunc != nil {
		dupCount = c.DupCountFunc(str)
	}
	if dupCount == 0 {
		dupCount = 1
	}
	for i := 0; i < dupCount; i++ {
		result = c.Encode(result, encrypt)
	}
	return result
}

// reverseString 反转字符串
func reverseString(str string) string {
	runes := []rune(str)
	length := len(runes)

	switch {
	case length <= 10:
		return reverse(runes, 0, length-1)
	case length <= 20:
		return segmentedReverse(runes, 3)
	case length <= 30:
		return segmentedReverse(runes, 4)
	default:
		return segmentedReverse(runes, 5)
	}
}

// reverse 反转字符数组的指定范围
func reverse(runes []rune, start, end int) string {
	for i, j := start, end; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes[start : end+1])
}

// segmentedReverse 分段反转字符数组
func segmentedReverse(runes []rune, segments int) string {
	length := len(runes)
	chunkSize := length / segments
	var result strings.Builder
	for i := 0; i < segments; i++ {
		start := i * chunkSize
		end := start + chunkSize - 1
		if i == segments-1 {
			end = length - 1
		}
		result.WriteString(reverse(runes, start, end))
	}
	return result.String()
}

// uniqueChars 返回字符串中的唯一字符
func uniqueChars(str string) string {
	seen := make(map[rune]bool)
	var result strings.Builder
	for _, char := range str {
		if !seen[char] {
			seen[char] = true
			result.WriteRune(char)
		}
	}
	return result.String()
}

// createMapping 创建字符映射
func createMapping(charset string, offset int, encrypt bool) map[string]string {
	mapping := make(map[string]string)
	length := len(charset)
	for i, r := range charset {
		if encrypt {
			mapping[string(r)] = string(charset[(i+offset)%length])
		} else {
			mapping[string(charset[(i+offset)%length])] = string(r)
		}
	}
	return mapping
}
