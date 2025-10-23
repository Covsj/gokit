package ievm

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// BuildLegacy 构造并签名 Legacy 交易
//
// 用途：
// - 适用于未启用伦敦升级的链，或作为 EIP-1559 回退方案。
// - 使用单一 gasPrice 计费模型：最终手续费 = gasUsed * gasPrice。
//
// 参数：
// - acc: IAccount 实例（必须包含私钥、RPC 客户端与 ChainID）。
// - to: 目标地址；若为空字符串，则视为合约创建交易（使用 NewContractCreation）。
// - value: 转账金额（以 wei 为单位）。
// - data: 调用数据；当合约创建时为合约字节码（可包含构造参数编码）。
// - gasLimit: 明确指定 gas 上限；为 0 时将自动 EstimateGas 估算。
// - gasPrice: 明确指定 gasPrice；为 nil 时将自动 SuggestGasPrice 获取。
// - nonceOpt: 可选 nonce；为 nil 时自动查询 PendingNonce。
//
// 行为：
// - 当 to 为空：构造合约创建交易（types.NewContractCreation）。
// - 当 to 不为空：构造普通转账/合约调用交易（types.NewTransaction）。
// - 使用 EIP-155 签名（types.NewEIP155Signer）。
//
// 返回：
// - 已签名的 *types.Transaction；错误时返回详细提示（地址校验、估算/建议价格失败等）。
func BuildLegacy(acc *IAccount, to string, value *big.Int, data []byte,
	gasLimit uint64, gasPrice *big.Int, nonceOpt *uint64) (*types.Transaction, error) {

	if acc == nil || acc.EInnerClient == nil {
		return nil, fmt.Errorf("client 未初始化")
	}
	var toPtr *common.Address
	if to != "" {
		if !IsValidAddress(to) {
			return nil, fmt.Errorf("无效 to 地址: %s", to)
		}
		addr := common.HexToAddress(to)
		toPtr = &addr
	}
	var nonce uint64
	var err error
	if nonceOpt != nil {
		nonce = *nonceOpt
	} else {
		nonce, err = acc.Nonce(acc.Address())
		if err != nil {
			return nil, err
		}
	}
	if gasPrice == nil {
		gasPrice, err = acc.SuggestGasPrice()
		if err != nil {
			return nil, err
		}
	}
	if gasLimit == 0 {
		gasLimit, err = acc.EstimateGas(acc.Address(), to, value, data)
		if err != nil {
			return nil, err
		}
	}
	var unsigned *types.Transaction
	if toPtr == nil {
		// 合约创建（legacy 模式）
		unsigned = types.NewContractCreation(nonce, value, gasLimit, gasPrice, data)
	} else {
		unsigned = types.NewTransaction(nonce, *toPtr, value, gasLimit, gasPrice, data)
	}
	signed, err := types.SignTx(unsigned, types.NewEIP155Signer(acc.ChainID), acc.key)
	if err != nil {
		return nil, err
	}
	return signed, nil
}

// BuildDynamic 构造并签名 EIP-1559（Dynamic Fee）交易
//
// 用途：
// - 适用于已启用伦敦升级（EIP-1559）的链。
// - 采用 baseFee + priorityFee（小费）模型；用户设置上限，实际支付不超过上限。
//
// 参数：
// - acc: IAccount 实例（必须包含私钥、RPC 客户端与 ChainID）。
// - to: 目标地址；可为空（一般 EIP-1559 创建合约推荐 EIP-1559 类型，to 为空即创建）。
// - value: 转账金额（wei）。
// - data: 调用数据/合约字节码。
// - gasLimit: 明确指定 gas 上限；为 0 时自动 EstimateGas。
// - maxFeePerGas: 总费用上限（包含 baseFee + priorityFee）；为 nil 时依据 SuggestGasPrice 推导（约 2x）。
// - maxPriorityFeePerGas: 小费上限（支付给打包者）；为 nil 时使用 SuggestGasTipCap。
// - nonceOpt: 可选 nonce；nil 时自动查询 PendingNonce。
//
// 行为：
// - 组装 types.DynamicFeeTx，并使用 London 签名（types.NewLondonSigner）。
// - 若未指定 maxFeePerGas，会以建议 gasPrice 近似推导，并确保 >= maxPriorityFeePerGas。
// - 实际有效单价 = min(maxFeePerGas, baseFee + maxPriorityFeePerGas)。
//
// 返回：
// - 已签名的 *types.Transaction；错误时返回详细提示（地址校验、建议/估算失败等）。
func BuildDynamic(acc *IAccount, to string, value *big.Int, data []byte,
	gasLimit uint64, maxFeePerGas, maxPriorityFeePerGas *big.Int,
	nonceOpt *uint64) (*types.Transaction, error) {

	if acc == nil || acc.EInnerClient == nil {
		return nil, fmt.Errorf("client 未初始化")
	}
	var toPtr *common.Address
	if to != "" {
		if !IsValidAddress(to) {
			return nil, fmt.Errorf("无效 to 地址: %s", to)
		}
		addr := common.HexToAddress(to)
		toPtr = &addr
	}
	var nonce uint64
	var err error
	if nonceOpt != nil {
		nonce = *nonceOpt
	} else {
		nonce, err = acc.Nonce(acc.Address())
		if err != nil {
			return nil, err
		}
	}
	if maxPriorityFeePerGas == nil {
		maxPriorityFeePerGas, err = acc.SuggestGasTipCap()
		if err != nil {
			return nil, err
		}
	}
	if maxFeePerGas == nil {
		base, err := acc.SuggestGasPrice()
		if err != nil {
			return nil, err
		}
		maxFeePerGas = new(big.Int).Mul(base, big.NewInt(2))
		if maxFeePerGas.Cmp(maxPriorityFeePerGas) < 0 {
			maxFeePerGas = new(big.Int).Set(maxPriorityFeePerGas)
		}
	}
	if gasLimit == 0 {
		gasLimit, err = acc.EstimateGas(acc.Address(), to, value, data)
		if err != nil {
			return nil, err
		}
	}
	unsigned := &types.DynamicFeeTx{
		ChainID:   acc.ChainID,
		Nonce:     nonce,
		GasTipCap: new(big.Int).Set(maxPriorityFeePerGas),
		GasFeeCap: new(big.Int).Set(maxFeePerGas),
		Gas:       gasLimit,
		To:        toPtr,
		Value:     value,
		Data:      data,
	}
	tx := types.NewTx(unsigned)
	signer := types.NewLondonSigner(acc.ChainID)
	signed, err := types.SignTx(tx, signer, acc.key)
	if err != nil {
		return nil, err
	}
	return signed, nil
}

// SendContractMethod 打 ABI 数据并发送交易（优先 EIP-1559）
func SendContractMethod(
	acc *IAccount, contract, abiJSON, method string, value *big.Int, args ...interface{}) (*types.Transaction, error) {

	if !IsValidAddress(contract) {
		return nil, fmt.Errorf("无效合约地址: %s", contract)
	}
	data, err := Pack(abiJSON, method, args...)
	if err != nil {
		return nil, err
	}
	signed, err := BuildDynamic(acc, contract, value, data, 0, nil, nil, nil)
	if err != nil {
		signed, err = BuildLegacy(acc, contract, value, data, 0, nil, nil)
	}
	if err != nil {
		return nil, err
	}
	_, err = acc.SendTx(signed)
	if err != nil {
		return nil, err
	}
	return signed, nil
}

// PreflightTx 在发送前模拟交易，尝试预测是否会失败（revert/无余额/参数错误等）
// 它会：
// 1) 使用 EstimateGas（若执行会 revert，节点通常返回错误）
// 2) 使用 CallContract 进行只读执行（再次捕获可能的 revert 原因）
func (a *IAccount) PreflightTx(from, to string, value *big.Int, data []byte) error {
	ctx := context.Background()

	var fromAddr common.Address
	if from != "" {
		if !IsValidAddress(from) {
			return fmt.Errorf("无效 from 地址: %s", from)
		}
		fromAddr = common.HexToAddress(from)
	}
	var toPtr *common.Address
	if to != "" {
		if !IsValidAddress(to) {
			return fmt.Errorf("无效 to 地址: %s", to)
		}
		addr := common.HexToAddress(to)
		toPtr = &addr
	}

	msg := ethereum.CallMsg{From: fromAddr, To: toPtr, Value: value, Data: data}

	// 先估算 gas：如果会 revert，许多链节点会直接返回错误
	if _, err := a.EInnerClient.EstimateGas(ctx, msg); err != nil {
		return err
	}

	// 再次用只读调用执行，捕获潜在的 revert 及原因
	if _, err := a.EInnerClient.CallContract(ctx, msg, nil); err != nil {
		return err
	}
	return nil
}

// PreflightContractMethod 对合约方法进行模拟，便于在发送前判断是否会失败
func (a *IAccount) PreflightContractMethod(from, contract, abiJSON, method string, value *big.Int, args ...interface{}) error {
	if !IsValidAddress(contract) {
		return fmt.Errorf("无效合约地址: %s", contract)
	}
	data, err := Pack(abiJSON, method, args...)
	if err != nil {
		return err
	}
	return a.PreflightTx(from, contract, value, data)
}
