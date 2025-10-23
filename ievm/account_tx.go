package ievm

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ==================== 核心交易发送方法 ====================

// SendTx 发送已签名交易
func (a *IAccount) SendTx(tx *types.Transaction) (common.Hash, error) {
	ctx := context.Background()
	if err := a.EInnerClient.SendTransaction(ctx, tx); err != nil {
		return common.Hash{}, err
	}

	// 等待交易上链并检查执行结果
	waitCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		receipt, err := a.EInnerClient.TransactionReceipt(waitCtx, tx.Hash())

		if err == ethereum.NotFound {
			select {
			case <-waitCtx.Done():
				return tx.Hash(), fmt.Errorf("等待交易上链超时: %s", tx.Hash().Hex())
			case <-ticker.C:
				continue
			}
		}
		if err != nil {
			return tx.Hash(), err
		}

		if receipt != nil {
			if receipt.Status == types.ReceiptStatusFailed {
				return tx.Hash(), fmt.Errorf("交易执行失败(status=0), tx: %s", tx.Hash().Hex())
			}
			return tx.Hash(), nil
		}

		select {
		case <-waitCtx.Done():
			return tx.Hash(), fmt.Errorf("等待交易上链超时: %s", tx.Hash().Hex())
		case <-ticker.C:
		}
	}
}

// ==================== 交易构建和发送 ====================

// SendETH 发送 ETH 转账
func (a *IAccount) SendETH(to string, amount *big.Int) (common.Hash, error) {
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return common.Hash{}, fmt.Errorf("转账金额必须大于0")
	}
	if !IsValidAddress(to) {
		return common.Hash{}, fmt.Errorf("无效的目标地址: %s", to)
	}

	// 检查余额是否足够
	hasEnough, err := a.HasEnoughBalance(amount)
	if err != nil {
		return common.Hash{}, fmt.Errorf("检查余额失败: %w", err)
	}
	if !hasEnough {
		return common.Hash{}, fmt.Errorf("余额不足")
	}

	// 构建并发送交易
	tx, err := a.BuildTxWithGas(to, nil, amount, 0, nil, nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("构建交易失败: %w", err)
	}

	return a.SendTx(tx)
}

// SendETHWithGas 发送 ETH 转账（带自定义 Gas）
func (a *IAccount) SendETHWithGas(to string, amount *big.Int, gasLimit uint64, gasPrice *big.Int) (common.Hash, error) {
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return common.Hash{}, fmt.Errorf("转账金额必须大于0")
	}
	if !IsValidAddress(to) {
		return common.Hash{}, fmt.Errorf("无效的目标地址: %s", to)
	}

	// 智能处理 Gas 参数
	gasParams, err := a.ProcessGasParams(gasLimit, gasPrice, to, amount, nil)
	if err != nil {
		return common.Hash{}, err
	}

	// 检查余额是否足够（包括 Gas 费用）
	totalCost := a.CalculateTotalCostWithGas(amount, gasParams.GasLimit, gasParams.GasPrice)
	hasEnough, err := a.HasEnoughBalance(totalCost)
	if err != nil {
		return common.Hash{}, fmt.Errorf("检查余额失败: %w", err)
	}
	if !hasEnough {
		return common.Hash{}, fmt.Errorf("余额不足，需要 %s wei", totalCost.String())
	}

	// 构建交易
	tx, err := a.BuildTxWithGas(to, nil, amount, gasParams.GasLimit, gasParams.GasPrice, nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("构建交易失败: %w", err)
	}

	return a.SendTx(tx)
}

// SendContractCall 发送合约调用交易
func (a *IAccount) SendContractCall(to string, data []byte, value *big.Int) (common.Hash, error) {
	return a.SendContractCallWithGas(to, data, value, 0, nil)
}

// SendContractCallWithGas 发送合约调用交易（带自定义 Gas）
func (a *IAccount) SendContractCallWithGas(to string, data []byte, value *big.Int, gasLimit uint64, gasPrice *big.Int) (common.Hash, error) {
	if !IsValidAddress(to) {
		return common.Hash{}, fmt.Errorf("无效的合约地址: %s", to)
	}
	if len(data) == 0 {
		return common.Hash{}, fmt.Errorf("调用数据不能为空")
	}
	if value == nil {
		value = big.NewInt(0)
	}

	// 智能处理 Gas 参数
	gasParams, err := a.ProcessGasParams(gasLimit, gasPrice, to, value, data)
	if err != nil {
		return common.Hash{}, err
	}

	// 检查余额是否足够（包括 Gas 费用）
	if value.Cmp(big.NewInt(0)) > 0 {
		totalCost := a.CalculateTotalCostWithGas(value, gasParams.GasLimit, gasParams.GasPrice)
		hasEnough, err := a.HasEnoughBalance(totalCost)
		if err != nil {
			return common.Hash{}, fmt.Errorf("检查余额失败: %w", err)
		}
		if !hasEnough {
			return common.Hash{}, fmt.Errorf("余额不足，需要 %s wei", totalCost.String())
		}
	}

	// 构建并发送交易
	tx, err := a.BuildTxWithGas(to, data, value, gasParams.GasLimit, gasParams.GasPrice, nil)
	if err != nil {
		return common.Hash{}, fmt.Errorf("构建交易失败: %w", err)
	}

	return a.SendTx(tx)
}

// BuildTxWithGas 构建交易（带自定义 Gas 参数）
func (a *IAccount) BuildTxWithGas(
	to string, data []byte, value *big.Int,
	gasLimit uint64, gasPrice *big.Int,
	nonce *uint64) (*types.Transaction, error) {
	if value == nil {
		value = big.NewInt(0)
	}
	if data == nil {
		data = []byte{}
	}

	// 智能处理 Gas 参数
	gasParams, err := a.ProcessGasParams(gasLimit, gasPrice, to, value, data)
	if err != nil {
		return nil, err
	}

	// 优先使用 EIP-1559 交易
	tx, err := BuildDynamic(a, to, value, data,
		gasParams.GasLimit, nil, nil, nonce)
	if err != nil {
		// 如果 EIP-1559 失败，回退到 Legacy 交易
		tx, err = BuildLegacy(a, to, value, data,
			gasParams.GasLimit, gasParams.GasPrice, nonce)
	}

	// 否则使用 EIP-1559 交易
	return tx, err
}

// ==================== 交易状态查询 ====================

// GetTransactionReceipt 获取交易回执
func (a *IAccount) GetTransactionReceipt(txHash common.Hash) (*types.Receipt, error) {
	ctx := context.Background()
	receipt, err := a.EInnerClient.TransactionReceipt(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("获取交易回执失败: %w", err)
	}
	return receipt, nil
}

// GetTransaction 获取交易详情
func (a *IAccount) GetTransaction(txHash common.Hash) (*types.Transaction, error) {
	ctx := context.Background()
	tx, isPending, err := a.EInnerClient.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("获取交易详情失败: %w", err)
	}
	if isPending {
		return nil, fmt.Errorf("交易仍在待处理状态")
	}
	return tx, nil
}

// WaitForTransaction 等待交易确认
func (a *IAccount) WaitForTransaction(txHash common.Hash, confirmations int) (*types.Receipt, error) {
	if confirmations <= 0 {
		confirmations = 1
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	var receipt *types.Receipt
	var err error

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("等待交易确认超时: %s", txHash.Hex())
		case <-ticker.C:
			receipt, err = a.GetTransactionReceipt(txHash)
			if err != nil {
				// 交易可能还在待处理状态，继续等待
				continue
			}

			// 检查确认数
			currentBlock, err := a.GetLatestBlockNumber()
			if err != nil {
				return nil, fmt.Errorf("获取最新区块号失败: %w", err)
			}

			confirmCount := new(big.Int).Sub(currentBlock, receipt.BlockNumber)
			if confirmCount.Int64() >= int64(confirmations) {
				return receipt, nil
			}
		}
	}
}

// IsTransactionSuccessful 检查交易是否成功
func (a *IAccount) IsTransactionSuccessful(txHash common.Hash) (bool, error) {
	receipt, err := a.GetTransactionReceipt(txHash)
	if err != nil {
		return false, err
	}
	return receipt.Status == types.ReceiptStatusSuccessful, nil
}

// CalculateTransactionFee 计算交易费用
func (a *IAccount) CalculateTransactionFee(tx *types.Transaction) (*big.Int, error) {
	if tx == nil {
		return nil, fmt.Errorf("交易不能为空")
	}

	receipt, err := a.GetTransactionReceipt(tx.Hash())
	if err != nil {
		return nil, fmt.Errorf("获取交易回执失败: %w", err)
	}

	// 计算实际使用的 Gas
	gasUsed := big.NewInt(int64(receipt.GasUsed))

	// 根据交易类型计算费用
	if tx.Type() == types.LegacyTxType {
		// Legacy 交易：gasUsed * gasPrice
		return new(big.Int).Mul(gasUsed, tx.GasPrice()), nil
	} else if tx.Type() == types.DynamicFeeTxType {
		// EIP-1559 交易：gasUsed * gasPrice (简化处理)
		return new(big.Int).Mul(gasUsed, tx.GasPrice()), nil
	}

	return nil, fmt.Errorf("不支持的交易类型")
}

// ==================== 批量交易操作 ====================

// BatchSendTransactions 批量发送交易
func (a *IAccount) BatchSendTransactions(txs []*types.Transaction) ([]common.Hash, error) {
	if len(txs) == 0 {
		return nil, fmt.Errorf("交易列表不能为空")
	}

	results := make([]common.Hash, len(txs))
	errors := make([]error, len(txs))

	for i, tx := range txs {
		if tx == nil {
			errors[i] = fmt.Errorf("第 %d 个交易为空", i+1)
			continue
		}

		hash, err := a.SendTx(tx)
		if err != nil {
			errors[i] = fmt.Errorf("发送第 %d 个交易失败: %w", i+1, err)
			continue
		}

		results[i] = hash
	}

	// 检查是否有错误
	var hasError bool
	for _, err := range errors {
		if err != nil {
			hasError = true
			break
		}
	}

	if hasError {
		return results, fmt.Errorf("批量发送交易时发生错误")
	}

	return results, nil
}

// ==================== 交易重试机制 ====================

// SendTxWithRetry 带重试的交易发送
func (a *IAccount) SendTxWithRetry(tx *types.Transaction, maxRetries int) (common.Hash, error) {
	if maxRetries <= 0 {
		maxRetries = 3
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		hash, err := a.SendTx(tx)
		if err == nil {
			return hash, nil
		}

		lastErr = err

		// 如果不是最后一次重试，等待一段时间
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return common.Hash{}, fmt.Errorf("重试 %d 次后仍然失败: %w", maxRetries, lastErr)
}

// SendTxWithSmartGas 智能 Gas 价格调整的交易发送
func (a *IAccount) SendTxWithSmartGas(tx *types.Transaction) (common.Hash, error) {
	// 首先尝试使用当前 Gas 价格发送
	hash, err := a.SendTx(tx)
	if err == nil {
		return hash, nil
	}

	// 如果失败，尝试提高 Gas 价格
	gasPrice, err := a.SuggestGasPrice()
	if err != nil {
		return common.Hash{}, fmt.Errorf("获取建议 Gas 价格失败: %w", err)
	}

	// 提高 20% 的 Gas 价格
	higherGasPrice := new(big.Int).Mul(gasPrice, big.NewInt(120))
	higherGasPrice.Div(higherGasPrice, big.NewInt(100))

	// 重新构建交易
	var newTx *types.Transaction
	nonce := tx.Nonce()
	if tx.Type() == types.LegacyTxType {
		newTx, err = a.BuildTxWithGas(
			tx.To().Hex(),
			tx.Data(),
			tx.Value(),
			tx.Gas(),
			higherGasPrice,
			&nonce,
		)
	} else {
		// 对于 EIP-1559 交易，调整 maxFeePerGas
		newTx, err = a.BuildTxWithGas(
			tx.To().Hex(),
			tx.Data(),
			tx.Value(),
			tx.Gas(),
			nil, // 使用 EIP-1559
			&nonce,
		)
	}

	if err != nil {
		return common.Hash{}, fmt.Errorf("重新构建交易失败: %w", err)
	}

	return a.SendTx(newTx)
}
