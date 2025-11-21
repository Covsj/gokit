package ievm

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

// IAccount 简化账户：持有私钥与地址, 封装 go-ethereum 客户端，提供简洁易用的多链 EVM 能力
type IAccount struct {
	RpcOpt       *RPCOptions
	ChainID      *big.Int
	key          *ecdsa.PrivateKey
	address      common.Address
	EInnerClient *ethclient.Client
}

// RPCOptions RPC 连接配置选项
// 使用示例：
//
//	// 方式1: 使用自定义头部（推荐用于 API Key 验证）
//	opts := &RPCOptions{
//	    RpcUrl: "https://eth-mainnet.g.alchemy.com/v2/your-api-key",
//	    Headers: map[string]string{
//	        "Authorization": "Bearer your-token",
//	        "X-API-Key": "your-api-key",
//	    },
//	    Timeout: 60, // 60秒超时
//	}
//	account, err := NewWithPrivateKey(privateKey, opts)
//
//	// 方式2: 仅设置超时
//	opts := &RPCOptions{
//	    RpcUrl: "https://eth-mainnet.g.alchemy.com/v2/your-api-key",
//	    Timeout: 120,
//	}
//	account, err := NewWithMnemonicIndex(mnemonic, 0, opts)
//
//	// 方式3: 使用完全自定义的 HTTP 客户端
//	customClient := &http.Client{
//	    Timeout: 90 * time.Second,
//	    Transport: &customTransport{...},
//	}
//	opts := &RPCOptions{
//	    RpcUrl: "https://eth-mainnet.g.alchemy.com/v2/your-api-key",
//	    CustomHTTPClient: customClient,
//	}
//	account, err := NewWithPrivateKey(privateKey, opts)
type RPCOptions struct {
	RpcUrl string
	// Headers 自定义 HTTP 请求头，用于 API Key 等验证
	// 例如: map[string]string{"Authorization": "Bearer token", "X-API-Key": "your-key"}
	Headers map[string]string

	// Timeout 连接超时时间（秒），默认 30 秒
	Timeout int

	// CustomHTTPClient 自定义 HTTP 客户端，如果设置则忽略 Headers 和 Timeout
	// 适用于需要完全控制 HTTP 客户端行为的场景
	CustomHTTPClient *http.Client
}

// ==================== 账户创建和管理 ====================

// dialRPCWithOptions 使用自定义选项创建 RPC 客户端
func dialRPCWithOptions(ctx context.Context, opts *RPCOptions) (*ethclient.Client, error) {
	if opts == nil || opts.RpcUrl == "" {
		return nil, fmt.Errorf("RPC URL 不能为空")
	}

	var httpClient *http.Client

	if opts.CustomHTTPClient != nil {
		// 使用自定义 HTTP 客户端
		httpClient = opts.CustomHTTPClient
	} else {
		// 创建默认 HTTP 客户端
		timeout := 30 * time.Second
		if opts.Timeout > 0 {
			timeout = time.Duration(opts.Timeout) * time.Second
		}

		httpClient = &http.Client{
			Timeout: timeout,
		}
	}

	// 构建 DialOptions
	var dialOpts []rpc.ClientOption

	// 添加 HTTP 客户端选项
	dialOpts = append(dialOpts, rpc.WithHTTPClient(httpClient))

	// 添加自定义头部选项
	if len(opts.Headers) > 0 {
		for k, v := range opts.Headers {
			dialOpts = append(dialOpts, rpc.WithHeader(k, v))
		}
	}

	// 使用 DialOptions 创建 RPC 客户端（推荐方式，替代已弃用的 DialHTTPWithClient）
	rpcClient, err := rpc.DialOptions(ctx, opts.RpcUrl, dialOpts...)
	if err != nil {
		return nil, err
	}

	// 创建 ethclient
	return ethclient.NewClient(rpcClient), nil
}

// NewWithMnemonicIndex 使用助记词和账户索引创建账户，派生路径 m/44'/60'/0'/0/{index}
// rpcURL: RPC 服务地址
// opts: 可选的 RPC 配置选项，支持自定义 HTTP 头部、超时等
func NewWithMnemonicIndex(mnemonic string, index int, opt *RPCOptions) (*IAccount, error) {
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

	eInnerClient, err := dialRPCWithOptions(ctx, opt)
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
		RpcOpt:       opt,
		EInnerClient: eInnerClient,
	}, nil
}

// NewWithPrivateKey 从 0x 私钥创建账户
// hexKey: 十六进制私钥（可带或不带 0x 前缀）
// rpcURL: RPC 服务地址
// opts: 可选的 RPC 配置选项，支持自定义 HTTP 头部、超时等
func NewWithPrivateKey(hexKey string, opt *RPCOptions) (*IAccount, error) {
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

	eInnerClient, err := dialRPCWithOptions(ctx, opt)
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
		RpcOpt:       opt,
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
