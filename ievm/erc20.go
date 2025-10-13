package ievm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ERC20ABI 标准 ERC20 ABI（常用方法）
const ERC20ABI = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transferFrom","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_spender","type":"address"},{"name":"_value","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_spender","type":"address"}],"name":"allowance","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"}]`

// ERC20 轻量封装
type ERC20 struct {
	Address string
	IAcc    *IAccount
}

func NewERC20(addr string, acc *IAccount) *ERC20 {
	return &ERC20{
		Address: addr,
		IAcc:    acc,
	}
}

func (e *ERC20) Erc20Name() (string, error) {
	data, _ := Pack(ERC20ABI, "name")
	out, err := e.IAcc.OnlyReadCall(e.Address, data)
	if err != nil {
		return "", err
	}
	vals, err := Unpack(ERC20ABI, "name", out)
	if err != nil {
		return "", err
	}
	if len(vals) > 0 {
		if s, ok := vals[0].(string); ok {
			return s, nil
		}
	}
	return "", nil
}

func (e *ERC20) Symbol() (string, error) {
	data, _ := Pack(ERC20ABI, "symbol")
	out, err := e.IAcc.OnlyReadCall(e.Address, data)
	if err != nil {
		return "", err
	}
	vals, err := Unpack(ERC20ABI, "symbol", out)
	if err != nil {
		return "", err
	}
	if len(vals) > 0 {
		if s, ok := vals[0].(string); ok {
			return s, nil
		}
	}
	return "", nil
}

func (e *ERC20) Decimals() (uint8, error) {
	data, _ := Pack(ERC20ABI, "decimals")
	out, err := e.IAcc.OnlyReadCall(e.Address, data)
	if err != nil {
		return 0, err
	}
	vals, err := Unpack(ERC20ABI, "decimals", out)
	if err != nil {
		return 0, err
	}
	if len(vals) > 0 {
		if d, ok := vals[0].(uint8); ok {
			return d, nil
		}
	}
	return 0, nil
}

func (e *ERC20) TotalSupply() (*big.Int, error) {
	data, _ := Pack(ERC20ABI, "totalSupply")
	out, err := e.IAcc.OnlyReadCall(e.Address, data)
	if err != nil {
		return nil, err
	}
	vals, err := Unpack(ERC20ABI, "totalSupply", out)
	if err != nil {
		return nil, err
	}
	if len(vals) > 0 {
		if s, ok := vals[0].(*big.Int); ok {
			return s, nil
		}
	}
	return big.NewInt(0), nil
}

func (e *ERC20) BalanceOf(owner string) (*big.Int, error) {
	data, _ := Pack(ERC20ABI, "balanceOf", common.HexToAddress(owner))
	out, err := e.IAcc.OnlyReadCall(e.Address, data)
	if err != nil {
		return nil, err
	}
	vals, err := Unpack(ERC20ABI, "balanceOf", out)
	if err != nil {
		return nil, err
	}
	if len(vals) > 0 {
		if b, ok := vals[0].(*big.Int); ok {
			return b, nil
		}
	}
	return big.NewInt(0), nil
}

func (e *ERC20) Allowance(owner, spender string) (*big.Int, error) {
	data, _ := Pack(ERC20ABI, "allowance", common.HexToAddress(owner), common.HexToAddress(spender))
	out, err := e.IAcc.OnlyReadCall(e.Address, data)
	if err != nil {
		return nil, err
	}
	vals, err := Unpack(ERC20ABI, "allowance", out)
	if err != nil {
		return nil, err
	}
	if len(vals) > 0 {
		if a, ok := vals[0].(*big.Int); ok {
			return a, nil
		}
	}
	return big.NewInt(0), nil
}

func (e *ERC20) Transfer(to string, amount *big.Int) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.Address, ERC20ABI,
		"transfer", big.NewInt(0), common.HexToAddress(to), amount)
}

func (e *ERC20) TransferFrom(from, to string, amount *big.Int) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.Address, ERC20ABI,
		"transferFrom", big.NewInt(0), common.HexToAddress(from), common.HexToAddress(to), amount)
}

func (e *ERC20) Approve(spender string, amount *big.Int) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.Address, ERC20ABI,
		"approve", big.NewInt(0), common.HexToAddress(spender), amount)
}
