package evm

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
)

// EVMAccount 表示以太坊账户，包含私钥和地址信息
type EVMAccount struct {
	privateKey *ecdsa.PrivateKey // 账户的ECDSA私钥
	address    common.Address    // 对应的以太坊地址
}

// NewAccount 创建一个空的EVMAccount实例
func NewAccount() *EVMAccount {
	return &EVMAccount{}
}

// createWalletFromMnemonic 通过助记词和派生路径创建钱包账户
func createWalletFromMnemonic(mnemonic, derivationPath string) (*EVMAccount, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, fmt.Errorf("generate seed from mnemonic failed: %w", err)
	}

	privateKey, err := deriveKeyFromSeed(seed, derivationPath)
	if err != nil {
		return nil, fmt.Errorf("derive private key from seed failed: %w", err)
	}

	walletAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	return &EVMAccount{
		privateKey: privateKey,
		address:    walletAddress,
	}, nil
}

// deriveKeyFromSeed 使用BIP44标准从种子派生私钥
func deriveKeyFromSeed(seed []byte, derivationPath string) (*ecdsa.PrivateKey, error) {
	hdPath, err := accounts.ParseDerivationPath(derivationPath)
	if err != nil {
		return nil, fmt.Errorf("invalid derivation path: %w", err)
	}

	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("create master key failed: %w", err)
	}

	derivedKey := masterKey
	for _, index := range hdPath {
		if derivedKey.IsAffectedByIssue172() {
			return nil, fmt.Errorf("key derivation affected by btcd issue #172")
		}

		derivedKey, err = derivedKey.Derive(index)
		if err != nil {
			return nil, fmt.Errorf("derive child key failed: %w", err)
		}
	}

	privateKey, err := derivedKey.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("extract private key failed: %w", err)
	}

	// IMPORTANT: Construct the ECDSA key using go-ethereum's curve implementation
	// to ensure compatibility with crypto.Sign, which requires crypto.S256().
	// Using btcec's ToECDSA() yields an ECDSA key with a different curve type
	// and causes: "private key curve is not secp256k1".
	privBytes := privateKey.Serialize()
	ethPrivKey, err := crypto.ToECDSA(privBytes)
	if err != nil {
		return nil, fmt.Errorf("convert private key to geth ecdsa failed: %w", err)
	}

	return ethPrivKey, nil
}

// NewAccountWithMnemonic 使用助记词创建账户，使用默认的BIP44路径
func NewAccountWithMnemonic(mnemonic string) (*EVMAccount, error) {
	// 使用标准的以太坊BIP44路径: m/44'/60'/0'/0/0
	defaultPath := "m/44'/60'/0'/0/0"
	return createWalletFromMnemonic(mnemonic, defaultPath)
}

// NewAccountWithMnemonicIndex 使用助记词和账户索引创建账户
func NewAccountWithMnemonicIndex(mnemonic string, accountIndex int) (*EVMAccount, error) {
	if accountIndex < 0 {
		accountIndex = 0
	}

	// 使用自定义索引的BIP44路径
	customPath := fmt.Sprintf("m/44'/60'/0'/0/%d", accountIndex)
	return createWalletFromMnemonic(mnemonic, customPath)
}

// NewAccountWithPrivateKey 使用私钥创建账户
func NewAccountWithPrivateKey(privateKeyHex string) (*EVMAccount, error) {
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

// Sign 对消息进行签名
func (a *EVMAccount) Sign(message []byte, password string) ([]byte, error) {
	return SignMessage(a.privateKey, message)
}

// SignHex 对十六进制格式的消息进行签名
func (a *EVMAccount) SignHex(messageHex string, password string) (string, error) {
	messageBytes, err := hexutil.Decode(messageHex)
	if err != nil {
		return "", fmt.Errorf("decode hex message failed: %w", err)
	}

	signature, err := a.Sign(messageBytes, password)
	if err != nil {
		return "", fmt.Errorf("sign message failed: %w", err)
	}

	return hexutil.Encode(signature), nil
}

// SignHash 对哈希进行签名，并按以太坊黄皮书要求调整V值
func (a *EVMAccount) SignHash(hash []byte) ([]byte, error) {
	signature, err := crypto.Sign(hash, a.privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign hash failed: %w", err)
	}
	// Transform V from 0/1 to 27/28 according to the yellow paper
	signature[crypto.RecoveryIDOffset] += 27
	return signature, nil
}
