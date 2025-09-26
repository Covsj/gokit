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
			// 使用更健壮的偏移量计算，避免 customCharset 为空时 panic
			if len(customCharset) == 0 {
				return 5 // 或者其他默认值
			}
			return len(str) % len(customCharset)
		}
	}
	if stableIndexFunc == nil {
		stableIndexFunc = func(str string) map[int]bool {
			res := map[int]bool{}
			// 示例：将索引为 0 和 5 的字符标记为稳定
			// 您可以根据需要自定义此逻辑
			if len(str) > 0 {
				res[0] = true
			}
			if len(str) > 5 {
				res[5] = true
			}
			for i := range str {
				if i != 0 && i != 5 && i%3 == 0 {
					res[i] = true
				}
			}
			return res
		}
	}
	if customCharset == "" {
		customCharset = "ijkKabc45672LMNO=1YZ%l?mn0uFG+HIJ3VWX~89AB#CDEqr^stPQRST*Udefghopv!wx$yz"
	}
	if dupCountFunc == nil {
		dupCountFunc = func(str string) int {
			// 使用更健壮的重复次数计算
			if len(str) == 0 {
				return 1 // 避免对空字符串取模
			}
			count := len(str) % 6
			if count == 0 { // 确保至少执行一次
				return 1
			}
			return count
		}
	}
	return &CaesarUtil{
		OffsetFunc:      offsetFunc,
		StableIndexFunc: stableIndexFunc,
		CustomCharset:   customCharset,
		DupCountFunc:    dupCountFunc,
	}
}

// EncodeMultipleTimes 对字符串进行多次凯撒加密或解密
func (c *CaesarUtil) EncodeMultipleTimes(str string, encrypt bool) string {
	result := str
	var dupCount int
	if c.DupCountFunc != nil {
		dupCount = c.DupCountFunc(str)
	} else {
		// 提供一个默认的 dupCountFunc 实现
		dupCount = 3 // 默认至少执行一次
		if len(str) > 0 {
			count := len(str) % 6
			if count > 0 {
				dupCount = count
			}
		}
	}

	// 确保 dupCount 至少为 1
	if dupCount <= 0 {
		dupCount = 1
	}

	for i := 0; i < dupCount; i++ {
		result = c.encode(result, encrypt)
	}
	return result
}

// Encode 使用凯撒加密对字符串进行编码或解码
func (c *CaesarUtil) encode(str string, encrypt bool) string {
	if !encrypt {
		// 解码前先反转
		str = reverseString(str)
	}

	var offset int
	if c.OffsetFunc != nil {
		offset = c.OffsetFunc(str)
	}
	// 如果偏移量为0或者大于等于字符集长度，则使用默认偏移量13（或其他合理值）
	// 确保偏移量在有效范围内 [1, len(customCharset)-1]
	if offset == 0 || (len(c.CustomCharset) > 0 && offset >= len(c.CustomCharset)) {
		offset = 13 % len(c.CustomCharset)           // 使用模运算确保偏移量有效
		if offset == 0 && len(c.CustomCharset) > 1 { // 如果取模后还是0，并且字符集不只有一个字符
			offset = 1 // 设置为最小有效偏移量
		}
	}

	processedCharset := c.CustomCharset
	// Consider if DES is truly necessary or if a simpler unique char approach suffices
	// If DES encryption is used, ensure proper key and IV handling
	if len(processedCharset) > 0 {
		// 使用 DES 加密生成新的字符集 - 确保这是必要的，并且密钥和 IV 安全
		// 注意：硬编码密钥通常不安全，应考虑其他密钥管理方法
		key := HashGenerator(processedCharset, "Md5").ToRawBytes()
		// ECB 模式不需要 IV，但使用更安全的模式（如 CBC）并提供 IV 是推荐的
		aesEncrypted := EncryptDES(
			BaseEncode(BaseEncode(processedCharset, "32").ToString(), "64").ToString(),
			"ECB", "PKCS7", // 考虑使用更安全的模式如 CBC
			key, nil).ToRawString() // ECB不需要IV，但如果是CBC则需要提供
		processedCharset = uniqueChars(processedCharset + aesEncrypted)
	}

	stableIndex := map[int]bool{}
	if c.StableIndexFunc != nil {
		stableIndex = c.StableIndexFunc(str)
	}

	// 确保 processedCharset 不为空
	if len(processedCharset) == 0 {
		// 如果字符集为空，无法进行映射，直接返回原始（或已反转）的字符串
		if encrypt {
			return reverseString(str) // 加密时最后反转
		}
		return str // 解码时已在开头反转
	}

	mapping := createMapping(processedCharset, offset, encrypt)

	var result strings.Builder
	result.Grow(len(str)) // 预分配容量以提高性能

	for i, char := range str {
		if stableIndex[i] {
			result.WriteRune(char)
			continue
		}
		// 直接使用 rune 进行查找
		if mappedChar, ok := mapping[char]; ok {
			result.WriteRune(mappedChar)
		} else {
			// 如果字符不在映射表中，保持原样
			result.WriteRune(char)
		}
	}

	if encrypt {
		// 加密后反转
		return reverseString(result.String())
	}

	// 解码时在函数开始时已反转，这里直接返回
	return result.String()
}

// reverseString 反转字符串 (使用标准 rune slice 原地反转)
func reverseString(str string) string {
	runes := []rune(str)
	n := len(runes)
	for i := 0; i < n/2; i++ {
		runes[i], runes[n-1-i] = runes[n-1-i], runes[i]
	}
	return string(runes)
}

// uniqueChars 返回字符串中的唯一字符，保持原始顺序
func uniqueChars(str string) string {
	seen := make(map[rune]struct{}) // 使用空的 struct{} 更节省内存
	var result strings.Builder
	result.Grow(len(str)) // 预分配容量
	for _, char := range str {
		if _, ok := seen[char]; !ok {
			seen[char] = struct{}{}
			result.WriteRune(char)
		}
	}
	return result.String()
}

// createMapping 创建字符映射 (键类型为 rune)
func createMapping(charset string, offset int, encrypt bool) map[rune]rune {
	mapping := make(map[rune]rune)
	runes := []rune(charset) // 转换为 rune slice
	length := len(runes)

	// 确保 length 大于 0 且 offset 有效
	if length == 0 {
		return mapping // 返回空映射
	}
	// 确保 offset 在 [0, length-1] 范围内
	offset = (offset%length + length) % length // 处理负数偏移和过大偏移

	for i, r := range runes {
		targetIndex := (i + offset) % length
		if encrypt {
			mapping[r] = runes[targetIndex]
		} else {
			// 解密时，目标字符映射回原始字符
			mapping[runes[targetIndex]] = r
		}
	}
	return mapping
}
