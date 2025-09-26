package ievm

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// SignMessage 处理消息签名的通用方法
func SignMessage(privateKey *ecdsa.PrivateKey, message []byte) ([]byte, error) {
	hash := SignHashForMsg(string(message))
	signature, err := crypto.Sign(hash, privateKey)
	if err != nil {
		return nil, fmt.Errorf("签名失败: %w", err)
	}

	signature[crypto.RecoveryIDOffset] += 27
	return signature, nil
}

// VerifySignature 验证签名的通用方法
func VerifySignature(pubkey, message, signedMsg string) (bool, error) {
	pubBytes, err := hexutil.Decode(pubkey)
	if err != nil {
		return false, fmt.Errorf("decode public key failed: %w", err)
	}

	messageBytes, err := hexutil.Decode(message)
	if err != nil {
		return false, fmt.Errorf("decode message failed: %w", err)
	}

	signatureBytes, err := hexutil.Decode(signedMsg)
	if err != nil {
		return false, fmt.Errorf("decode signature failed: %w", err)
	}

	hash := SignHashForMsg(string(messageBytes))
	signatureBytes = signatureBytes[:len(signatureBytes)-1]

	return crypto.VerifySignature(pubBytes, hash, signatureBytes), nil
}

// SignHashForMsg 计算以太坊消息的哈希值
// 在消息前添加以太坊特定的前缀，以防止跨协议攻击
func SignHashForMsg(message string) []byte {
	prefixedMsg := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	return crypto.Keccak256([]byte(prefixedMsg))
}

// IsValidSignature 验证原始字节格式的签名
func IsValidSignature(publicKey, message, signature []byte) bool {
	messageHash := SignHashForMsg(string(message))
	// 移除恢复ID
	signature = signature[:len(signature)-1]
	return crypto.VerifySignature(publicKey, messageHash, signature)
}
