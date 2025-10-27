package ievm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ERC1155ABI 标准 ERC1155 ABI（常用方法）
const ERC1155ABI = `[{"constant":true,"inputs":[{"name":"account","type":"address"},{"name":"id","type":"uint256"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"accounts","type":"address[]"},{"name":"ids","type":"uint256[]"}],"name":"balanceOfBatch","outputs":[{"name":"","type":"uint256[]"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"id","type":"uint256"}],"name":"uri","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"account","type":"address"},{"name":"operator","type":"address"}],"name":"isApprovedForAll","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"id","type":"uint256"},{"name":"amount","type":"uint256"},{"name":"data","type":"bytes"}],"name":"safeTransferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"from","type":"address"},{"name":"to","type":"address"},{"name":"ids","type":"uint256[]"},{"name":"amounts","type":"uint256[]"},{"name":"data","type":"bytes"}],"name":"safeBatchTransferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"operator","type":"address"},{"name":"approved","type":"bool"}],"name":"setApprovalForAll","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"id","type":"uint256"},{"name":"amount","type":"uint256"},{"name":"data","type":"bytes"}],"name":"mint","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"to","type":"address"},{"name":"ids","type":"uint256[]"},{"name":"amounts","type":"uint256[]"},{"name":"data","type":"bytes"}],"name":"mintBatch","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"from","type":"address"},{"name":"id","type":"uint256"},{"name":"amount","type":"uint256"}],"name":"burn","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"from","type":"address"},{"name":"ids","type":"uint256[]"},{"name":"amounts","type":"uint256[]"}],"name":"burnBatch","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

// ERC1155 轻量封装
type ERC1155 struct {
	TokenContract string
	IAcc          *IAccount
}

func NewERC1155(tokenContract string, acc *IAccount) *ERC1155 {
	return &ERC1155{
		TokenContract: tokenContract,
		IAcc:          acc,
	}
}

// 只读方法

func (e *ERC1155) BalanceOf(account string, id *big.Int) (*big.Int, error) {
	data, err := Pack(ERC1155ABI, "balanceOf", common.HexToAddress(account), id)
	if err != nil {
		return nil, err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return nil, err
	}
	vals, err := Unpack(ERC1155ABI, "balanceOf", out)
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

func (e *ERC1155) BalanceOfBatch(accounts []string, ids []*big.Int) ([]*big.Int, error) {
	// 转换地址数组
	addresses := make([]common.Address, len(accounts))
	for i, addr := range accounts {
		addresses[i] = common.HexToAddress(addr)
	}

	data, err := Pack(ERC1155ABI, "balanceOfBatch", addresses, ids)
	if err != nil {
		return nil, err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return nil, err
	}
	vals, err := Unpack(ERC1155ABI, "balanceOfBatch", out)
	if err != nil {
		return nil, err
	}
	if len(vals) > 0 {
		if balances, ok := vals[0].([]*big.Int); ok {
			return balances, nil
		}
	}
	return nil, nil
}

func (e *ERC1155) URI(id *big.Int) (string, error) {
	data, err := Pack(ERC1155ABI, "uri", id)
	if err != nil {
		return "", err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return "", err
	}
	vals, err := Unpack(ERC1155ABI, "uri", out)
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

func (e *ERC1155) IsApprovedForAll(account, operator string) (bool, error) {
	data, err := Pack(ERC1155ABI, "isApprovedForAll", common.HexToAddress(account), common.HexToAddress(operator))
	if err != nil {
		return false, err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return false, err
	}
	vals, err := Unpack(ERC1155ABI, "isApprovedForAll", out)
	if err != nil {
		return false, err
	}
	if len(vals) > 0 {
		if b, ok := vals[0].(bool); ok {
			return b, nil
		}
	}
	return false, nil
}

// 写入方法

func (e *ERC1155) SafeTransferFrom(from, to string, id, amount *big.Int, data []byte) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC1155ABI,
		"safeTransferFrom", big.NewInt(0), common.HexToAddress(from), common.HexToAddress(to), id, amount, data)
}

func (e *ERC1155) SafeBatchTransferFrom(from, to string, ids, amounts []*big.Int, data []byte) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC1155ABI,
		"safeBatchTransferFrom", big.NewInt(0), common.HexToAddress(from), common.HexToAddress(to), ids, amounts, data)
}

func (e *ERC1155) SetApprovalForAll(operator string, approved bool) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC1155ABI,
		"setApprovalForAll", big.NewInt(0), common.HexToAddress(operator), approved)
}

func (e *ERC1155) Mint(to string, id, amount *big.Int, data []byte) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC1155ABI,
		"mint", big.NewInt(0), common.HexToAddress(to), id, amount, data)
}

func (e *ERC1155) MintBatch(to string, ids, amounts []*big.Int, data []byte) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC1155ABI,
		"mintBatch", big.NewInt(0), common.HexToAddress(to), ids, amounts, data)
}

func (e *ERC1155) Burn(from string, id, amount *big.Int) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC1155ABI,
		"burn", big.NewInt(0), common.HexToAddress(from), id, amount)
}

func (e *ERC1155) BurnBatch(from string, ids, amounts []*big.Int) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC1155ABI,
		"burnBatch", big.NewInt(0), common.HexToAddress(from), ids, amounts)
}

// 便捷方法

// SafeTransferFromSimple 简化版本，不需要data参数
func (e *ERC1155) SafeTransferFromSimple(from, to string, id, amount *big.Int) (*types.Transaction, error) {
	return e.SafeTransferFrom(from, to, id, amount, []byte{})
}

// SafeBatchTransferFromSimple 简化版本，不需要data参数
func (e *ERC1155) SafeBatchTransferFromSimple(from, to string, ids, amounts []*big.Int) (*types.Transaction, error) {
	return e.SafeBatchTransferFrom(from, to, ids, amounts, []byte{})
}

// MintSimple 简化版本，不需要data参数
func (e *ERC1155) MintSimple(to string, id, amount *big.Int) (*types.Transaction, error) {
	return e.Mint(to, id, amount, []byte{})
}

// MintBatchSimple 简化版本，不需要data参数
func (e *ERC1155) MintBatchSimple(to string, ids, amounts []*big.Int) (*types.Transaction, error) {
	return e.MintBatch(to, ids, amounts, []byte{})
}
