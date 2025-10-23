package ievm

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// ==================== 智能 Gas 处理使用示例 ====================

// ExampleSmartGasUsage 展示智能 Gas 处理的使用方法
func ExampleSmartGasUsage() {
	// 创建账户
	account, err := NewWithPrivateKey("0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", "https://eth-mainnet.g.alchemy.com/v2/test")
	if err != nil {
		fmt.Printf("创建账户失败: %v\n", err)
		return
	}
	defer account.Close()

	to := "0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6"
	amount := big.NewInt(1000000000000000000) // 1 ETH

	fmt.Println("=== 智能 Gas 处理示例 ===")

	// 1. 完全自动 Gas（推荐）
	fmt.Println("\n1. 完全自动 Gas 处理:")
	txHash1, err := account.SendETHWithGas(to, amount, 0, nil)
	if err != nil {
		fmt.Printf("自动 Gas 发送失败: %v\n", err)
	} else {
		fmt.Printf("自动 Gas 发送成功: %s\n", txHash1.Hex())
	}

	// 2. 自定义 Gas 限制，自动 Gas 价格
	fmt.Println("\n2. 自定义 Gas 限制，自动 Gas 价格:")
	txHash2, err := account.SendETHWithGas(to, amount, 21000, nil)
	if err != nil {
		fmt.Printf("自定义 Gas 限制发送失败: %v\n", err)
	} else {
		fmt.Printf("自定义 Gas 限制发送成功: %s\n", txHash2.Hex())
	}

	// 3. 自动 Gas 限制，自定义 Gas 价格
	fmt.Println("\n3. 自动 Gas 限制，自定义 Gas 价格:")
	customGasPrice := big.NewInt(20000000000) // 20 Gwei
	txHash3, err := account.SendETHWithGas(to, amount, 0, customGasPrice)
	if err != nil {
		fmt.Printf("自定义 Gas 价格发送失败: %v\n", err)
	} else {
		fmt.Printf("自定义 Gas 价格发送成功: %s\n", txHash3.Hex())
	}

	// 4. 完全自定义 Gas
	fmt.Println("\n4. 完全自定义 Gas:")
	customGasLimit := uint64(25000)
	customGasPrice2 := big.NewInt(25000000000) // 25 Gwei
	txHash4, err := account.SendETHWithGas(to, amount, customGasLimit, customGasPrice2)
	if err != nil {
		fmt.Printf("完全自定义 Gas 发送失败: %v\n", err)
	} else {
		fmt.Printf("完全自定义 Gas 发送成功: %s\n", txHash4.Hex())
	}

	// 5. 合约调用示例
	fmt.Println("\n5. 合约调用示例:")
	contractData := []byte{0x70, 0xa0, 0x82, 0x31} // 示例合约调用数据
	txHash5, err := account.SendContractCallWithGas(to, contractData, big.NewInt(0), 0, nil)
	if err != nil {
		fmt.Printf("合约调用失败: %v\n", err)
	} else {
		fmt.Printf("合约调用成功: %s\n", txHash5.Hex())
	}

	// 6. 获取 Gas 建议
	fmt.Println("\n6. Gas 建议信息:")
	gasPrice, gasTipCap, gasFeeCap, err := account.GetGasSuggestions()
	if err != nil {
		fmt.Printf("获取 Gas 建议失败: %v\n", err)
	} else {
		fmt.Printf("建议 Gas 价格: %s Wei\n", gasPrice.String())
		fmt.Printf("建议优先费: %s Wei\n", gasTipCap.String())
		fmt.Printf("建议费用上限: %s Wei\n", gasFeeCap.String())
	}

	// 7. 智能 Gas 参数处理
	fmt.Println("\n7. 智能 Gas 参数处理:")
	gasParams, err := account.ProcessGasParams(0, nil, to, amount, nil)
	if err != nil {
		fmt.Printf("处理 Gas 参数失败: %v\n", err)
	} else {
		fmt.Printf("处理后的 Gas 限制: %d\n", gasParams.GasLimit)
		fmt.Printf("处理后的 Gas 价格: %s Wei\n", gasParams.GasPrice.String())

		// 计算总成本
		totalCost := account.CalculateTotalCostWithGas(amount, gasParams.GasLimit, gasParams.GasPrice)
		fmt.Printf("总成本: %s Wei\n", totalCost.String())
		fmt.Printf("总成本: %s ETH\n", FormatETH(totalCost))
	}
}

// ==================== Gas 策略示例 ====================

// GasStrategy 定义 Gas 策略
type GasStrategy int

const (
	// GasStrategyAuto 自动 Gas 策略
	GasStrategyAuto GasStrategy = iota
	// GasStrategyFast 快速 Gas 策略
	GasStrategyFast
	// GasStrategyStandard 标准 Gas 策略
	GasStrategyStandard
	// GasStrategySlow 慢速 Gas 策略
	GasStrategySlow
)

// GetGasByStrategy 根据策略获取 Gas 参数
func (a *IAccount) GetGasByStrategy(strategy GasStrategy) (gasPrice *big.Int, err error) {
	basePrice, err := a.SuggestGasPrice()
	if err != nil {
		return nil, err
	}

	switch strategy {
	case GasStrategyAuto:
		return a.GetOptimalGasPrice()
	case GasStrategyFast:
		// 快速策略：提高 50%
		fastPrice := new(big.Int).Mul(basePrice, big.NewInt(150))
		fastPrice.Div(fastPrice, big.NewInt(100))
		return fastPrice, nil
	case GasStrategyStandard:
		// 标准策略：使用建议价格
		return basePrice, nil
	case GasStrategySlow:
		// 慢速策略：降低 20%
		slowPrice := new(big.Int).Mul(basePrice, big.NewInt(80))
		slowPrice.Div(slowPrice, big.NewInt(100))
		return slowPrice, nil
	default:
		return basePrice, nil
	}
}

// SendETHWithStrategy 使用策略发送 ETH
func (a *IAccount) SendETHWithStrategy(to string, amount *big.Int, strategy GasStrategy) (common.Hash, error) {
	gasPrice, err := a.GetGasByStrategy(strategy)
	if err != nil {
		return common.Hash{}, fmt.Errorf("获取策略 Gas 价格失败: %w", err)
	}

	return a.SendETHWithGas(to, amount, 0, gasPrice)
}

// SendContractCallWithStrategy 使用策略发送合约调用
func (a *IAccount) SendContractCallWithStrategy(to string, data []byte, value *big.Int, strategy GasStrategy) (common.Hash, error) {
	gasPrice, err := a.GetGasByStrategy(strategy)
	if err != nil {
		return common.Hash{}, fmt.Errorf("获取策略 Gas 价格失败: %w", err)
	}

	return a.SendContractCallWithGas(to, data, value, 0, gasPrice)
}
