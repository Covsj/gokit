package evm

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// BuildLegacy 构造并签名 Legacy 交易
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
	unsigned := types.NewTransaction(nonce, common.Address{}, value, gasLimit, gasPrice, data)
	if toPtr != nil {
		unsigned = types.NewTransaction(nonce, *toPtr, value, gasLimit, gasPrice, data)
	}
	signed, err := types.SignTx(unsigned, types.NewEIP155Signer(acc.ChainID), acc.key)
	if err != nil {
		return nil, err
	}
	return signed, nil
}

// BuildDynamic 构造并签名 EIP-1559 交易
func BuildDynamic(acc *IAccount, to string, value *big.Int, data []byte, gasLimit uint64, maxFeePerGas, maxPriorityFeePerGas *big.Int, nonceOpt *uint64) (*types.Transaction, error) {

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
	unsigned := &types.DynamicFeeTx{ChainID: new(big.Int).Set(acc.ChainID), Nonce: nonce, GasTipCap: new(big.Int).Set(maxPriorityFeePerGas), GasFeeCap: new(big.Int).Set(maxFeePerGas), Gas: gasLimit, To: toPtr, Value: value, Data: data}
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
	acc *IAccount, contract, abiJSON, method string, args ...interface{}) (*types.Transaction, error) {

	if !IsValidAddress(contract) {
		return nil, fmt.Errorf("无效合约地址: %s", contract)
	}
	data, err := Pack(abiJSON, method, args...)
	if err != nil {
		return nil, err
	}
	signed, err := BuildDynamic(acc, contract, big.NewInt(0), data, 0, nil, nil, nil)
	if err != nil {
		signed, err = BuildLegacy(acc, contract, big.NewInt(0), data, 0, nil, nil)
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

// SignTypedData 对 EIP-712 进行签名
func SignTypedData(hashToSign []byte, acc *IAccount) ([]byte, error) {
	if len(hashToSign) != 32 {
		return nil, fmt.Errorf("hash size must be 32")
	}
	return crypto.Sign(hashToSign, acc.key)
}
