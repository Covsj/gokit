package ievm

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// 表示以太坊账户，包含私钥和地址信息
type EVMAccount struct {
	privateKey *ecdsa.PrivateKey // 账户的ECDSA私钥
	address    common.Address    // 对应的以太坊地址
}

// 创建一个空的EVMAccount实例
func New(mnemonic string, accountIndex int, privateKeyHex string) (evmAcc *EVMAccount, err error) {
	if mnemonic != "" {
		return NewWithMnemonicIndex(mnemonic, accountIndex)
	}
	if privateKeyHex != "" {
		return NewWithPrivateKey(privateKeyHex)
	}
	mnemonic = GenerateMnemonic()

	return NewWithMnemonic(mnemonic)
}

// NewWithMnemonic 使用助记词创建账户，使用默认的BIP44路径
func NewWithMnemonic(mnemonic string) (*EVMAccount, error) {
	if mnemonic == "" {
		return nil, fmt.Errorf("助记词不能为空")
	}
	return createWalletFromMnemonic(mnemonic, DefaultDerivationPath)
}

// NewWithMnemonicIndex 使用助记词和账户索引创建账户
func NewWithMnemonicIndex(mnemonic string, accountIndex int) (*EVMAccount, error) {
	if mnemonic == "" {
		return nil, fmt.Errorf("助记词不能为空")
	}
	if accountIndex < 0 {
		accountIndex = 0
	}

	// 使用自定义索引的BIP44路径
	customPath := fmt.Sprintf("m/44'/60'/0'/0/%d", accountIndex)
	return createWalletFromMnemonic(mnemonic, customPath)
}

// NewWithPrivateKey 使用私钥创建账户
func NewWithPrivateKey(privateKeyHex string) (*EVMAccount, error) {
	if privateKeyHex == "" {
		return nil, fmt.Errorf("私钥不能为空")
	}

	// 移除可能的0x前缀
	if len(privateKeyHex) > 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key hex: %w", err)
	}

	walletAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	return &EVMAccount{
		privateKey: privateKey,
		address:    walletAddress,
	}, nil
}

// PrivateKey 获取账户私钥的字节表示
func (a *EVMAccount) PrivateKey() ([]byte, error) {
	if a.privateKey == nil {
		return nil, fmt.Errorf("private key not initialized")
	}
	return crypto.FromECDSA(a.privateKey), nil
}

// PrivateKeyHex 获取账户私钥的十六进制表示
func (a *EVMAccount) PrivateKeyHex() (string, error) {
	keyBytes, err := a.PrivateKey()
	if err != nil {
		return "", err
	}
	return hexutil.Encode(keyBytes), nil
}

// PublicKey 获取账户公钥的字节表示
func (a *EVMAccount) PublicKey() []byte {
	return crypto.FromECDSAPub(&a.privateKey.PublicKey)
}

// PublicKeyHex 获取账户公钥的十六进制表示
func (a *EVMAccount) PublicKeyHex() string {
	pubKeyBytes := crypto.FromECDSAPub(&a.privateKey.PublicKey)
	return hexutil.Encode(pubKeyBytes)
}

// Address 获取账户的以太坊地址
func (a *EVMAccount) Address() string {
	return a.address.Hex()
}

// Sign 用途: 签名用户消息，如登录认证、授权操作
func (a *EVMAccount) Sign(message []byte, password string) (string, error) {
	if a.privateKey == nil {
		return "", fmt.Errorf("私钥未初始化")
	}
	if len(message) == 0 {
		return "", fmt.Errorf("消息不能为空")
	}
	signatureBytes, err := SignMessage(a.privateKey, message)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(signatureBytes), err
}

// SignHex 对十六进制格式的消息进行签名
func (a *EVMAccount) SignHex(messageHex string, password string) (string, error) {
	messageBytes, err := hexutil.Decode(messageHex)
	if err != nil {
		return "", fmt.Errorf("解码hex消息失败: %w", err)
	}

	signature, err := a.Sign(messageBytes, password)
	if err != nil {
		return "", fmt.Errorf("签名消息失败: %w", err)
	}

	return signature, nil
}

// SignHash 直接签名32字节的哈希值，区块链交易签名，智能合约调用签名
func (a *EVMAccount) SignHash(hash []byte) ([]byte, error) {
	if a.privateKey == nil {
		return nil, fmt.Errorf("私钥未初始化")
	}
	if len(hash) != 32 {
		return nil, fmt.Errorf("哈希长度必须为32字节")
	}

	signature, err := crypto.Sign(hash, a.privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign hash failed: %w", err)
	}
	// Transform V from 0/1 to 27/28 according to the yellow paper
	signature[crypto.RecoveryIDOffset] += 27
	return signature, nil
}
