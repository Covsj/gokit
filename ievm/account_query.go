package ievm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ==================== 余额查询 ====================

// Balance 查询地址余额（最新块）
func (a *IAccount) Balance(address string) (*big.Int, error) {
	ctx := context.Background()
	if !IsValidAddress(address) {
		return nil, fmt.Errorf("无效地址: %s", address)
	}
	return a.EInnerClient.BalanceAt(ctx, common.HexToAddress(address), nil)
}

// GetETHBalance 获取账户的 ETH 余额
func (a *IAccount) GetETHBalance() (*big.Int, error) {
	return a.Balance(a.Address())
}

// HasEnoughBalance 检查账户是否有足够余额
func (a *IAccount) HasEnoughBalance(amount *big.Int) (bool, error) {
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return true, nil
	}

	balance, err := a.GetETHBalance()
	if err != nil {
		return false, err
	}

	return balance.Cmp(amount) >= 0, nil
}

// BatchGetBalances 批量查询余额
func (a *IAccount) BatchGetBalances(addresses []string) (map[string]*big.Int, error) {
	if len(addresses) == 0 {
		return nil, fmt.Errorf("地址列表不能为空")
	}

	results := make(map[string]*big.Int)
	ctx := context.Background()

	for _, addr := range addresses {
		if !IsValidAddress(addr) {
			results[addr] = nil
			continue
		}

		balance, err := a.EInnerClient.BalanceAt(ctx, common.HexToAddress(addr), nil)
		if err != nil {
			results[addr] = nil
			continue
		}

		results[addr] = balance
	}

	return results, nil
}

// ==================== Nonce 查询 ====================

// Nonce 查询挂起 nonce
func (a *IAccount) Nonce(address string) (uint64, error) {
	ctx := context.Background()
	if !IsValidAddress(address) {
		return 0, fmt.Errorf("无效地址: %s", address)
	}
	return a.EInnerClient.PendingNonceAt(ctx, common.HexToAddress(address))
}

// BatchGetNonces 批量查询 Nonce
func (a *IAccount) BatchGetNonces(addresses []string) (map[string]uint64, error) {
	if len(addresses) == 0 {
		return nil, fmt.Errorf("地址列表不能为空")
	}

	results := make(map[string]uint64)
	ctx := context.Background()

	for _, addr := range addresses {
		if !IsValidAddress(addr) {
			results[addr] = 0
			continue
		}

		nonce, err := a.EInnerClient.PendingNonceAt(ctx, common.HexToAddress(addr))
		if err != nil {
			results[addr] = 0
			continue
		}

		results[addr] = nonce
	}

	return results, nil
}

// ==================== Gas 相关查询 ====================

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

// SuggestGasFeeCap 建议费用上限（EIP-1559）
func (a *IAccount) SuggestGasFeeCap() (*big.Int, error) {
	// 注意：go-ethereum 客户端可能没有 SuggestGasFeeCap 方法
	// 这里使用建议的 gasPrice 乘以 2 作为费用上限
	gasPrice, err := a.SuggestGasPrice()
	if err != nil {
		return nil, err
	}

	// 费用上限通常是 gasPrice 的 2 倍
	feeCap := new(big.Int).Mul(gasPrice, big.NewInt(2))
	return feeCap, nil
}

// GetGasSuggestions 获取完整的 Gas 建议
func (a *IAccount) GetGasSuggestions() (gasPrice, gasTipCap, gasFeeCap *big.Int, err error) {
	ctx := context.Background()

	// 获取 Legacy Gas 价格
	gasPrice, err = a.EInnerClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("获取建议 Gas 价格失败: %w", err)
	}

	// 获取 EIP-1559 Gas 建议
	gasTipCap, err = a.EInnerClient.SuggestGasTipCap(ctx)
	if err != nil {
		return gasPrice, nil, nil, fmt.Errorf("获取建议优先费失败: %w", err)
	}

	gasFeeCap, err = a.SuggestGasFeeCap()
	if err != nil {
		return gasPrice, gasTipCap, nil, fmt.Errorf("获取建议费用上限失败: %w", err)
	}

	return gasPrice, gasTipCap, gasFeeCap, nil
}

// GetDynamicGasPrice 动态 Gas 价格调整提高10%
func (a *IAccount) GetDynamicGasPrice() (*big.Int, error) {
	gasPrice, err := a.SuggestGasPrice()
	if err != nil {
		return nil, err
	}

	// 提高 10% 的 Gas 价格以确保交易被快速处理
	higherGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(110))
	higherGasPrice.Div(higherGasPrice, big.NewInt(100))

	return higherGasPrice, nil
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

// ==================== 区块查询 ====================

// GetLatestBlockNumber 获取最新区块号
func (a *IAccount) GetLatestBlockNumber() (*big.Int, error) {
	ctx := context.Background()
	blockNumber, err := a.EInnerClient.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取最新区块号失败: %w", err)
	}
	return big.NewInt(int64(blockNumber)), nil
}

// GetBlock 获取区块信息
func (a *IAccount) GetBlock(blockNumber *big.Int) (*types.Block, error) {
	ctx := context.Background()
	block, err := a.EInnerClient.BlockByNumber(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("获取区块信息失败: %w", err)
	}
	return block, nil
}

// GetBlockByHash 根据哈希获取区块信息
func (a *IAccount) GetBlockByHash(blockHash common.Hash) (*types.Block, error) {
	ctx := context.Background()
	block, err := a.EInnerClient.BlockByHash(ctx, blockHash)
	if err != nil {
		return nil, fmt.Errorf("获取区块信息失败: %w", err)
	}
	return block, nil
}

// GetBlockHeader 获取区块头信息
func (a *IAccount) GetBlockHeader(blockNumber *big.Int) (*types.Header, error) {
	ctx := context.Background()
	header, err := a.EInnerClient.HeaderByNumber(ctx, blockNumber)
	if err != nil {
		return nil, fmt.Errorf("获取区块头失败: %w", err)
	}
	return header, nil
}

// ==================== 合约调用 ====================

// OnlyReadCall 只读调用，不会发起交易
func (a *IAccount) OnlyReadCall(to string, data []byte) ([]byte, error) {
	ctx := context.Background()
	if !IsValidAddress(to) {
		return nil, fmt.Errorf("无效合约地址: %s", to)
	}
	addr := common.HexToAddress(to)
	msg := ethereum.CallMsg{To: &addr, Data: data}
	return a.EInnerClient.CallContract(ctx, msg, nil)
}

// ==================== 网络状态查询 ====================

// GetNetworkStatus 获取网络状态信息
func (a *IAccount) GetNetworkStatus() (map[string]interface{}, error) {
	status := make(map[string]interface{})

	// 获取最新区块号
	latestBlock, err := a.GetLatestBlockNumber()
	if err != nil {
		return nil, fmt.Errorf("获取最新区块号失败: %w", err)
	}
	status["latestBlock"] = latestBlock

	// 获取网络信息
	chainID, networkName, err := a.GetNetworkInfo()
	if err != nil {
		return nil, fmt.Errorf("获取网络信息失败: %w", err)
	}
	status["chainID"] = chainID
	status["networkName"] = networkName

	// 获取 Gas 建议
	gasPrice, gasTipCap, gasFeeCap, err := a.GetGasSuggestions()
	if err != nil {
		status["gasError"] = err.Error()
	} else {
		status["gasPrice"] = gasPrice
		status["gasTipCap"] = gasTipCap
		status["gasFeeCap"] = gasFeeCap
	}

	// 检查连接状态
	status["connected"] = a.IsConnected()

	return status, nil
}
