package ievm

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// ==================== 智能 Gas 处理 ====================

// GasParams Gas 参数结构
type GasParams struct {
	GasLimit uint64
	GasPrice *big.Int
}

// ProcessGasParams 智能处理 Gas 参数
func (a *IAccount) ProcessGasParams(gasLimit uint64, gasPrice *big.Int, to string, value *big.Int, data []byte) (*GasParams, error) {
	var finalGasLimit uint64
	var finalGasPrice *big.Int
	var err error

	// 处理 Gas 限制
	if gasLimit == 0 {
		// 自动估算 Gas 限制
		finalGasLimit, err = a.EstimateGas(a.Address(), to, value, data)
		if err != nil {
			return nil, fmt.Errorf("估算 Gas 限制失败: %w", err)
		}
	} else {
		finalGasLimit = gasLimit
	}

	// 处理 Gas 价格
	if gasPrice == nil {
		// 自动获取 Gas 价格建议
		finalGasPrice, err = a.SuggestGasPrice()
		if err != nil {
			return nil, fmt.Errorf("获取 Gas 价格建议失败: %w", err)
		}
	} else {
		finalGasPrice = gasPrice
	}

	return &GasParams{
		GasLimit: finalGasLimit,
		GasPrice: finalGasPrice,
	}, nil
}

// CalculateTotalCostWithGas 计算包含 Gas 费用的总成本
func (a *IAccount) CalculateTotalCostWithGas(amount *big.Int, gasLimit uint64, gasPrice *big.Int) *big.Int {
	if amount == nil {
		amount = big.NewInt(0)
	}
	if gasPrice == nil {
		return amount
	}

	gasCost := new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
	return new(big.Int).Add(amount, gasCost)
}

// ValidateGasParams 验证 Gas 参数
func (a *IAccount) ValidateGasParams(gasLimit uint64, gasPrice *big.Int) error {
	if gasLimit == 0 {
		return fmt.Errorf("Gas 限制不能为0")
	}

	if gasPrice != nil && gasPrice.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("Gas 价格必须大于0")
	}

	return nil
}

// GetOptimalGasPrice 获取最优 Gas 价格（考虑网络拥堵情况）
func (a *IAccount) GetOptimalGasPrice() (*big.Int, error) {
	// 获取基础 Gas 价格
	basePrice, err := a.SuggestGasPrice()
	if err != nil {
		return nil, err
	}

	// 获取动态 Gas 价格（提高10%）
	dynamicPrice, err := a.GetDynamicGasPrice()
	if err != nil {
		return basePrice, nil // 如果获取动态价格失败，使用基础价格
	}

	// 选择较高的价格以确保交易被快速处理
	if dynamicPrice.Cmp(basePrice) > 0 {
		return dynamicPrice, nil
	}

	return basePrice, nil
}

// ==================== 账户特定工具 ====================

// IsContractAddress 检查是否为合约地址
func (a *IAccount) IsContractAddress(address string) (bool, error) {
	if !IsValidAddress(address) {
		return false, fmt.Errorf("无效地址格式")
	}

	// 获取地址的代码
	code, err := a.EInnerClient.CodeAt(context.Background(), common.HexToAddress(address), nil)
	if err != nil {
		return false, fmt.Errorf("获取地址代码失败: %w", err)
	}

	// 如果代码长度大于0，则为合约地址
	return len(code) > 0, nil
}

// CalculateGasCost 计算 Gas 成本
func (a *IAccount) CalculateGasCost(gasLimit uint64, gasPrice *big.Int) *big.Int {
	if gasPrice == nil {
		return big.NewInt(0)
	}

	return new(big.Int).Mul(big.NewInt(int64(gasLimit)), gasPrice)
}

// EstimateTotalCost 估算总成本（转账金额 + Gas 费用）
func (a *IAccount) EstimateTotalCost(amount *big.Int, gasLimit uint64, gasPrice *big.Int) *big.Int {
	if amount == nil {
		amount = big.NewInt(0)
	}

	gasCost := a.CalculateGasCost(gasLimit, gasPrice)
	return new(big.Int).Add(amount, gasCost)
}

// ==================== 调试工具 ====================

// GetAccountInfo 获取账户信息摘要
func (a *IAccount) GetAccountInfo() map[string]interface{} {
	info := make(map[string]interface{})

	info["address"] = a.Address()
	info["truncatedAddress"] = TruncateAddress(a.Address())
	info["chainID"] = a.ChainID
	info["rpc"] = a.RPC

	// 获取余额
	balance, err := a.GetETHBalance()
	if err != nil {
		info["balance"] = "获取失败"
		info["balanceError"] = err.Error()
	} else {
		info["balance"] = balance.String()
		info["balanceETH"] = FormatETH(balance)
	}

	// 获取网络信息
	chainID, networkName, err := a.GetNetworkInfo()
	if err != nil {
		info["network"] = "获取失败"
		info["networkError"] = err.Error()
	} else {
		info["network"] = networkName
		info["chainID"] = chainID
	}

	// 连接状态
	info["connected"] = a.IsConnected()

	return info
}

// PrintAccountInfo 打印账户信息
func (a *IAccount) PrintAccountInfo() {
	info := a.GetAccountInfo()

	fmt.Println("=== 账户信息 ===")
	fmt.Printf("地址: %s\n", info["address"])
	fmt.Printf("地址(截断): %s\n", info["truncatedAddress"])
	fmt.Printf("网络: %s\n", info["network"])
	fmt.Printf("链ID: %s\n", info["chainID"])
	fmt.Printf("余额: %s ETH (%s Wei)\n", info["balanceETH"], info["balance"])
	fmt.Printf("连接状态: %t\n", info["connected"])
	fmt.Printf("RPC: %s\n", info["rpc"])
}

// ==================== 错误处理工具 ====================

// IsRevertError 检查是否为合约回滚错误
func (a *IAccount) IsRevertError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "revert") ||
		strings.Contains(errStr, "execution reverted") ||
		strings.Contains(errStr, "VM execution error")
}

// IsInsufficientFundsError 检查是否为余额不足错误
func (a *IAccount) IsInsufficientFundsError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "insufficient funds") ||
		strings.Contains(errStr, "余额不足")
}

// IsGasLimitError 检查是否为 Gas 限制错误
func (a *IAccount) IsGasLimitError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return strings.Contains(errStr, "gas limit") ||
		strings.Contains(errStr, "out of gas")
}
