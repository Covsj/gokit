package ievm

import (
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// ==================== 签名方法 ====================

// SignHash 直接对哈希进行签名
func (a *IAccount) SignHash(hash []byte) (string, error) {
	if len(hash) != 32 {
		return "", fmt.Errorf("哈希长度必须为32字节")
	}

	// 使用私钥签名
	sig, err := crypto.Sign(hash, a.key)
	if err != nil {
		return "", fmt.Errorf("crypto.Sign失败: %w", err)
	}

	// 确保签名是65字节（以太坊标准）
	if len(sig) != 65 {
		return "", errors.New("签名长度无效")
	}

	// 调整v值以符合以太坊标准
	v := int(sig[64])
	switch v {
	case 0, 1:
		v += 27 // 0/1 → 27/28 (legacy)
	case 27, 28:
		// 已经是正确的格式
	default:
		return "", errors.New("无效的v值")
	}
	sig[64] = byte(v)

	return hexutil.Encode(sig), nil
}

// SignPersonal 对消息执行 personal_sign（带前缀）
func (a *IAccount) SignPersonal(message []byte) (string, error) {
	const EthereumPrefix = "\x19Ethereum Signed Message:\n"

	if len(message) == 0 {
		return "", fmt.Errorf("消息不能为空")
	}

	// 添加以太坊签名前缀
	prefix := fmt.Sprintf("%s%d", EthereumPrefix, len(message))
	prefixedMessage := append([]byte(prefix), message...)
	hash := crypto.Keccak256(prefixedMessage)

	// 使用统一的签名方法
	return a.SignHash(hash)
}

// GetSignEIP712Hash 获取EIP712签名前哈希
func (a *IAccount) GetSignEIP712Hash(domain apitypes.TypedDataDomain,
	typeMap map[string][]apitypes.Type, primaryType string, messageMap map[string]any) ([]byte, error) {

	// 输入验证
	if primaryType == "" {
		return nil, fmt.Errorf("primaryType 不能为空")
	}
	if len(messageMap) == 0 {
		return nil, fmt.Errorf("messageMap 不能为空")
	}
	if len(typeMap) == 0 {
		return nil, fmt.Errorf("typeMap 不能为空")
	}

	// 构建类型定义
	types := apitypes.Types{}
	for typeName, typeFields := range typeMap {
		if len(typeFields) == 0 {
			return nil, fmt.Errorf("类型 %s 的字段不能为空", typeName)
		}
		types[typeName] = typeFields
	}

	// 构建消息
	message := apitypes.TypedDataMessage{}
	for k, v := range messageMap {
		message[k] = v
	}

	typedData := apitypes.TypedData{
		Domain:      domain,
		Types:       types,
		PrimaryType: primaryType,
		Message:     message,
	}

	// 计算结构体哈希
	hash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, fmt.Errorf("结构体哈希失败: %w", err)
	}

	// 计算域名分隔符哈希
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, fmt.Errorf("域名哈希失败: %w", err)
	}

	// 创建最终哈希
	finalHash := crypto.Keccak256(
		[]byte("\x19\x01"),
		domainSeparator,
		hash,
	)

	// 直接对哈希进行签名，不使用 SignPersonal
	return finalHash, nil
}

// SignEIP712 进行EIP712签名
func (a *IAccount) SignEIP712(domain apitypes.TypedDataDomain,
	typeMap map[string][]apitypes.Type, primaryType string, messageMap map[string]any) (string, error) {
	finalHash, err := a.GetSignEIP712Hash(domain, typeMap, primaryType, messageMap)
	if err != nil {
		return "", err
	}

	// 直接对哈希进行签名，不使用 SignPersonal
	return a.SignHash(finalHash)
}

// ==================== 签名验证 ====================

// VerifyEIP712Signature 验证 EIP712 签名
func (a *IAccount) VerifyEIP712Signature(domain apitypes.TypedDataDomain,
	typeMap map[string][]apitypes.Type, primaryType string, messageMap map[string]any, signature string) (bool, error) {
	finalHash, err := a.GetSignEIP712Hash(domain, typeMap, primaryType, messageMap)
	if err != nil {
		return false, err
	}

	// 验证签名
	return a.VerifySignature(finalHash, signature)
}

// VerifySignature 验证签名
func (a *IAccount) VerifySignature(hash []byte, signature string) (bool, error) {
	if len(hash) != 32 {
		return false, fmt.Errorf("哈希长度必须为32字节")
	}

	// 解析签名
	sig, err := hexutil.Decode(signature)
	if err != nil {
		return false, fmt.Errorf("解析签名失败: %w", err)
	}

	if len(sig) != 65 {
		return false, fmt.Errorf("签名长度无效")
	}

	// 恢复公钥
	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return false, fmt.Errorf("恢复公钥失败: %w", err)
	}

	// 验证公钥是否匹配
	expectedAddr := crypto.PubkeyToAddress(*pubKey)
	actualAddr := a.address

	return expectedAddr == actualAddr, nil
}

// RecoverAddress 从签名恢复地址
func (a *IAccount) RecoverAddress(hash []byte, signature string) (common.Address, error) {
	if len(hash) != 32 {
		return common.Address{}, fmt.Errorf("哈希长度必须为32字节")
	}

	// 解析签名
	sig, err := hexutil.Decode(signature)
	if err != nil {
		return common.Address{}, fmt.Errorf("解析签名失败: %w", err)
	}

	if len(sig) != 65 {
		return common.Address{}, fmt.Errorf("签名长度无效")
	}

	// 恢复公钥
	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return common.Address{}, fmt.Errorf("恢复公钥失败: %w", err)
	}

	// 计算地址
	address := crypto.PubkeyToAddress(*pubKey)
	return address, nil
}

// ==================== 签名工具方法 ====================

// SignMessage 签名任意消息（自动添加前缀）
func (a *IAccount) SignMessage(message string) (string, error) {
	return a.SignPersonal([]byte(message))
}

// SignBytes 签名字节数组（自动添加前缀）
func (a *IAccount) SignBytes(data []byte) (string, error) {
	return a.SignPersonal(data)
}

// SignHex 签名十六进制字符串
func (a *IAccount) SignHex(hexString string) (string, error) {
	data, err := hexutil.Decode(hexString)
	if err != nil {
		return "", fmt.Errorf("解析十六进制字符串失败: %w", err)
	}
	return a.SignPersonal(data)
}

// ==================== 签名验证工具方法 ====================

// VerifyMessageSignature 验证消息签名
func (a *IAccount) VerifyMessageSignature(message string, signature string) (bool, error) {
	// 计算消息哈希（带前缀）
	const EthereumPrefix = "\x19Ethereum Signed Message:\n"
	prefix := fmt.Sprintf("%s%d", EthereumPrefix, len(message))
	prefixedMessage := append([]byte(prefix), []byte(message)...)
	hash := crypto.Keccak256(prefixedMessage)

	return a.VerifySignature(hash, signature)
}

// VerifyBytesSignature 验证字节数组签名
func (a *IAccount) VerifyBytesSignature(data []byte, signature string) (bool, error) {
	// 计算数据哈希（带前缀）
	const EthereumPrefix = "\x19Ethereum Signed Message:\n"
	prefix := fmt.Sprintf("%s%d", EthereumPrefix, len(data))
	prefixedMessage := append([]byte(prefix), data...)
	hash := crypto.Keccak256(prefixedMessage)

	return a.VerifySignature(hash, signature)
}

// ==================== 批量签名验证 ====================

// BatchVerifySignatures 批量验证签名
func (a *IAccount) BatchVerifySignatures(hashes [][]byte, signatures []string) ([]bool, error) {
	if len(hashes) != len(signatures) {
		return nil, fmt.Errorf("哈希和签名数量不匹配")
	}

	results := make([]bool, len(hashes))

	for i, hash := range hashes {
		valid, err := a.VerifySignature(hash, signatures[i])
		if err != nil {
			results[i] = false
			continue
		}
		results[i] = valid
	}

	return results, nil
}

// ==================== 签名恢复工具 ====================

// RecoverAddressFromMessage 从消息签名恢复地址
func (a *IAccount) RecoverAddressFromMessage(message string, signature string) (common.Address, error) {
	// 计算消息哈希（带前缀）
	const EthereumPrefix = "\x19Ethereum Signed Message:\n"
	prefix := fmt.Sprintf("%s%d", EthereumPrefix, len(message))
	prefixedMessage := append([]byte(prefix), []byte(message)...)
	hash := crypto.Keccak256(prefixedMessage)

	return a.RecoverAddress(hash, signature)
}

// RecoverAddressFromBytes 从字节数组签名恢复地址
func (a *IAccount) RecoverAddressFromBytes(data []byte, signature string) (common.Address, error) {
	// 计算数据哈希（带前缀）
	const EthereumPrefix = "\x19Ethereum Signed Message:\n"
	prefix := fmt.Sprintf("%s%d", EthereumPrefix, len(data))
	prefixedMessage := append([]byte(prefix), data...)
	hash := crypto.Keccak256(prefixedMessage)

	return a.RecoverAddress(hash, signature)
}
