package ievm

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/Covsj/gokit/ilog"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/chenzhijie/go-web3/utils"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

// GetRPCByChainID 根据链ID获取随机RPC端点
func GetRPCByChainID(chainID int64) string {
	rpcList, exists := RPCEndpoints[chainID]
	if !exists {
		// 默认使用以太坊主网
		rpcList = RPCEndpoints[ChainIDEthereum]
	}

	if len(rpcList) == 0 {
		return ""
	}

	// 使用全局随机数生成器
	return rpcList[globalRand.Intn(len(rpcList))]
}

// GetChainIDByEVM 根据EVM名称获取链ID
func GetChainIDByEVM(evm string) int64 {
	switch strings.ToLower(evm) {
	case "bsc", "binance":
		return ChainIDBSC
	case "eth", "ethereum":
		return ChainIDEthereum
	default:
		return ChainIDEthereum // 默认以太坊
	}
}

// FromDecimals 将代币数量从小数位转换为实际数量
func FromDecimals(value *big.Int, decimals int64) *big.Float {
	return utils.NewUtils().FromDecimals(value, decimals)
}

// ToDecimals 将实际数量转换为代币数量（考虑小数位）
func ToDecimals(value uint64, decimals int64) *big.Int {
	return utils.NewUtils().ToDecimals(value, decimals)
}

// ValidateAddress 验证以太坊地址格式
func ValidateAddress(address string) bool {
	if len(address) != 42 {
		return false
	}
	if !strings.HasPrefix(address, "0x") {
		return false
	}
	// 可以添加更多验证逻辑
	return true
}

// 通过助记词和派生路径创建钱包账户
func createWalletFromMnemonic(mnemonic, derivationPath string) (*EVMAccount, error) {
	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, fmt.Errorf("从助记词生成seed失败: %w", err)
	}

	privateKey, err := deriveKeyFromSeed(seed, derivationPath)
	if err != nil {
		return nil, fmt.Errorf("从seed加载私钥失败: %w", err)
	}

	walletAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	return &EVMAccount{
		privateKey: privateKey,
		address:    walletAddress,
	}, nil
}

// 使用BIP44标准从种子派生私钥
func deriveKeyFromSeed(seed []byte, derivationPath string) (*ecdsa.PrivateKey, error) {
	hdPath, err := accounts.ParseDerivationPath(derivationPath)
	if err != nil {
		return nil, fmt.Errorf("错误的派生路径: %w", err)
	}

	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("创建HD Master失败: %w", err)
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

// GenerateMnemonic 生成助记词
func GenerateMnemonic() string {
	// 生成128位熵（12个单词）
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		ilog.Error("生成熵失败", "错误", err)
		return ""
	}
	// 生成助记词
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		ilog.Error("生成助记词失败", "错误", err)
		return ""
	}
	return mnemonic
}

// GenerateEVMAddress 根据助记词生成ETH地址和私钥
func GenerateEVMAddress(mnemonic string) (string, string) {
	harden := func(i uint32) uint32 { return i | bip32.FirstHardenedChild }
	seed := bip39.NewSeed(mnemonic, "") // 64 bytes
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return "", ""
	}

	// path m/44'/60'/0'/0/0
	k1, _ := masterKey.NewChildKey(harden(44))
	k2, _ := k1.NewChildKey(harden(60))
	k3, _ := k2.NewChildKey(harden(0))
	k4, _ := k3.NewChildKey(0)
	k5, _ := k4.NewChildKey(0)
	ethPrivKeyBytes := k5.Key
	// Ensure length 32
	if len(ethPrivKeyBytes) != 32 {
		// Some implementations include a leading zero, trim/pad if necessary
		if len(ethPrivKeyBytes) > 32 {
			ethPrivKeyBytes = ethPrivKeyBytes[len(ethPrivKeyBytes)-32:]
		} else {
			padded := make([]byte, 32)
			copy(padded[32-len(ethPrivKeyBytes):], ethPrivKeyBytes)
			ethPrivKeyBytes = padded
		}
	}

	ethKey, err := crypto.ToECDSA(ethPrivKeyBytes)
	if err != nil {
		return "", ""
	}
	ethAddr := crypto.PubkeyToAddress(ethKey.PublicKey)
	evmPrivateKey, evmAddr := hex.EncodeToString(ethPrivKeyBytes), ethAddr.Hex()
	return evmPrivateKey, evmAddr
}
