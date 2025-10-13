package evm

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

// IAccount 简化账户：持有私钥与地址, 封装 go-ethereum 客户端，提供简洁易用的多链 EVM 能力
type IAccount struct {
	RPC          string
	ChainID      *big.Int
	key          *ecdsa.PrivateKey
	address      common.Address
	EInnerClient *ethclient.Client
}

// NewWithMnemonicIndex 使用助记词和账户索引创建账户，派生路径 m/44'/60'/0'/0/{index}
func NewWithMnemonicIndex(mnemonic string, index int,
	rpcURL string) (*IAccount, error) {
	if mnemonic == "" {
		return nil, fmt.Errorf("助记词不能为空")
	}
	if index < 0 {
		index = 0
	}
	// 校验助记词
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("无效助记词")
	}
	// 生成种子（不使用密码短语）
	seed := bip39.NewSeed(mnemonic, "")
	// BIP32 根
	master, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, fmt.Errorf("创建主密钥失败: %w", err)
	}
	harden := func(i uint32) uint32 { return i | bip32.FirstHardenedChild }
	// m/44'/60'/0'/0/{index}
	k44, err := master.NewChildKey(harden(44))
	if err != nil {
		return nil, err
	}
	k60, err := k44.NewChildKey(harden(60))
	if err != nil {
		return nil, err
	}
	k0h, err := k60.NewChildKey(harden(0))
	if err != nil {
		return nil, err
	}
	k0, err := k0h.NewChildKey(0)
	if err != nil {
		return nil, err
	}
	ki, err := k0.NewChildKey(uint32(index))
	if err != nil {
		return nil, err
	}
	// 32字节私钥
	privBytes := ki.Key
	if len(privBytes) != 32 {
		if len(privBytes) > 32 {
			privBytes = privBytes[len(privBytes)-32:]
		} else {
			padded := make([]byte, 32)
			copy(padded[32-len(privBytes):], privBytes)
			privBytes = padded
		}
	}
	priv, err := crypto.ToECDSA(privBytes)
	if err != nil {
		return nil, fmt.Errorf("转换私钥失败: %w", err)
	}

	addr := crypto.PubkeyToAddress(priv.PublicKey)

	ctx := context.Background()
	eInnerClient, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("连接 RPC 失败: %w", err)
	}
	chainId, err := eInnerClient.ChainID(ctx)
	if err != nil {
		chainId = big.NewInt(0)
	}

	return &IAccount{
		key:          priv,
		address:      addr,
		ChainID:      chainId,
		RPC:          rpcURL,
		EInnerClient: eInnerClient,
	}, nil
}

// NewWithPrivateKey 从 0x 私钥创建账户
func NewWithPrivateKey(hexKey string, rpcURL string) (*IAccount, error) {
	if hexKey == "" {
		return nil, fmt.Errorf("私钥不能为空")
	}
	if len(hexKey) > 2 && hexKey[:2] == "0x" {
		hexKey = hexKey[2:]
	}
	priv, err := crypto.HexToECDSA(hexKey)
	if err != nil {
		return nil, fmt.Errorf("错误的私钥: %w", err)
	}
	addr := crypto.PubkeyToAddress(priv.PublicKey)
	ctx := context.Background()
	eInnerClient, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, fmt.Errorf("连接 RPC 失败: %w", err)
	}
	chainId, err := eInnerClient.ChainID(ctx)
	if err != nil {
		chainId = big.NewInt(0)
	}
	return &IAccount{
		key:          priv,
		address:      addr,
		ChainID:      chainId,
		RPC:          rpcURL,
		EInnerClient: eInnerClient,
	}, nil
}

func (a *IAccount) Address() string {
	return a.address.Hex()
}

func (a *IAccount) PrivateKey() *ecdsa.PrivateKey {
	return a.key
}

func (a *IAccount) PrivateKeyHex() string {
	return hexutil.Encode(crypto.FromECDSA(a.key))
}

// Close 关闭连接
func (a *IAccount) Close() {
	if a != nil && a.EInnerClient != nil {
		a.EInnerClient.Close()
	}
}

// Balance 查询地址余额（最新块）
func (a *IAccount) Balance(address string) (*big.Int, error) {
	ctx := context.Background()
	if !IsValidAddress(address) {
		return nil, fmt.Errorf("无效地址: %s", address)
	}
	return a.EInnerClient.BalanceAt(ctx, common.HexToAddress(address), nil)
}

// Nonce 查询挂起 nonce
func (a *IAccount) Nonce(address string) (uint64, error) {
	ctx := context.Background()
	if !IsValidAddress(address) {
		return 0, fmt.Errorf("无效地址: %s", address)
	}
	return a.EInnerClient.PendingNonceAt(ctx, common.HexToAddress(address))
}

// SuggestGasPrice 建议 gasPrice（legacy）
func (a *IAccount) SuggestGasPrice() (*big.Int, error) {
	ctx := context.Background()
	return a.EInnerClient.SuggestGasPrice(ctx)
}

// SuggestGasTipCap 建议优先费（EIP-1559）
func (a *IAccount) SuggestGasTipCap() (*big.Int, error) {
	ctx := context.Background()
	return a.EInnerClient.SuggestGasTipCap(ctx)
}

// EstimateGas 估算 gasLimit
func (a *IAccount) EstimateGas(from, to string, value *big.Int, data []byte) (uint64, error) {
	ctx := context.Background()
	var fromAddr common.Address
	if from != "" {
		if !IsValidAddress(from) {
			return 0, fmt.Errorf("无效 from 地址: %s", from)
		}
		fromAddr = common.HexToAddress(from)
	}
	var toPtr *common.Address
	if to != "" {
		if !IsValidAddress(to) {
			return 0, fmt.Errorf("无效 to 地址: %s", to)
		}
		addr := common.HexToAddress(to)
		toPtr = &addr
	}
	msg := ethereum.CallMsg{From: fromAddr, To: toPtr, Value: value, Data: data}
	return a.EInnerClient.EstimateGas(ctx, msg)
}

// Call 只读调用
func (a *IAccount) OnlyReadCall(to string, data []byte) ([]byte, error) {
	ctx := context.Background()
	if !IsValidAddress(to) {
		return nil, fmt.Errorf("无效合约地址: %s", to)
	}
	addr := common.HexToAddress(to)
	msg := ethereum.CallMsg{To: &addr, Data: data}
	return a.EInnerClient.CallContract(ctx, msg, nil)
}

// SendTx 发送已签名交易
func (a *IAccount) SendTx(tx *types.Transaction) (common.Hash, error) {
	ctx := context.Background()
	if err := a.EInnerClient.SendTransaction(ctx, tx); err != nil {
		return common.Hash{}, err
	}
	return tx.Hash(), nil
}

// SignPersonal 对消息执行 personal_sign（带前缀）
func (a *IAccount) SignPersonal(message []byte) (string, error) {
	if len(message) == 0 {
		return "", fmt.Errorf("消息不能为空")
	}
	prefix := fmt.Sprintf("\x19Ethereum Signed Message:\n%d", len(message))
	hash := crypto.Keccak256([]byte(prefix), message)
	sig, err := crypto.Sign(hash, a.key)
	if err != nil {
		return "", err
	}
	return hexutil.Encode(sig), nil
}
