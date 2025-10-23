package ievm

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

// ==================== 账户创建和管理 ====================

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
	k44, _ := master.NewChildKey(harden(44))
	k60, _ := k44.NewChildKey(harden(60))
	k0h, _ := k60.NewChildKey(harden(0))
	k0, _ := k0h.NewChildKey(0)
	ki, _ := k0.NewChildKey(uint32(index))

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

// ==================== 基础账户信息 ====================

// Address 获取账户地址
func (a *IAccount) Address() string {
	return a.address.Hex()
}

// PrivateKey 获取私钥
func (a *IAccount) PrivateKey() *ecdsa.PrivateKey {
	return a.key
}

// PrivateKeyHex 获取十六进制私钥
func (a *IAccount) PrivateKeyHex() string {
	return hexutil.Encode(crypto.FromECDSA(a.key))
}

// Close 关闭连接
func (a *IAccount) Close() {
	if a != nil && a.EInnerClient != nil {
		a.EInnerClient.Close()
	}
}

// ==================== 网络和连接状态 ====================

// IsConnected 检查网络连接状态
func (a *IAccount) IsConnected() bool {
	if a == nil || a.EInnerClient == nil {
		return false
	}
	ctx := context.Background()
	_, err := a.EInnerClient.ChainID(ctx)
	return err == nil
}

// GetNetworkInfo 获取网络信息
func (a *IAccount) GetNetworkInfo() (chainID *big.Int, networkName string, err error) {
	if a == nil || a.EInnerClient == nil {
		return nil, "", fmt.Errorf("客户端未初始化")
	}

	chainID = a.ChainID
	if chainID == nil {
		ctx := context.Background()
		chainID, err = a.EInnerClient.ChainID(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("获取链ID失败: %w", err)
		}
	}

	// 根据链ID返回网络名称
	networkName = getNetworkName(chainID)
	return chainID, networkName, nil
}

// getNetworkName 根据链ID获取网络名称
func getNetworkName(chainID *big.Int) string {
	switch chainID.Uint64() {
	case 1:
		return "Ethereum Mainnet"
	case 3:
		return "Ropsten Testnet"
	case 4:
		return "Rinkeby Testnet"
	case 5:
		return "Goerli Testnet"
	case 42:
		return "Kovan Testnet"
	case 56:
		return "BSC Mainnet"
	case 97:
		return "BSC Testnet"
	case 137:
		return "Polygon Mainnet"
	case 80001:
		return "Polygon Mumbai Testnet"
	case 250:
		return "Fantom Opera"
	case 4002:
		return "Fantom Testnet"
	case 43114:
		return "Avalanche C-Chain"
	case 43113:
		return "Avalanche Fuji Testnet"
	case 10:
		return "Optimism"
	case 420:
		return "Optimism Goerli"
	case 42161:
		return "Arbitrum One"
	case 421613:
		return "Arbitrum Goerli"
	default:
		return fmt.Sprintf("Unknown Network (ChainID: %s)", chainID.String())
	}
}
