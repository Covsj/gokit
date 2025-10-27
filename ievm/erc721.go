package ievm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ERC721ABI 标准 ERC721 ABI（常用方法）
const ERC721ABI = `[{"constant":true,"inputs":[],"name":"name","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_tokenId","type":"uint256"}],"name":"ownerOf","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_tokenId","type":"uint256"}],"name":"tokenURI","outputs":[{"name":"","type":"string"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_owner","type":"address"},{"name":"_operator","type":"address"}],"name":"isApprovedForAll","outputs":[{"name":"","type":"bool"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":true,"inputs":[{"name":"_tokenId","type":"uint256"}],"name":"getApproved","outputs":[{"name":"","type":"address"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_tokenId","type":"uint256"}],"name":"transfer","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_tokenId","type":"uint256"}],"name":"transferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_approved","type":"address"},{"name":"_tokenId","type":"uint256"}],"name":"approve","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_operator","type":"address"},{"name":"_approved","type":"bool"}],"name":"setApprovalForAll","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_tokenId","type":"uint256"}],"name":"safeTransferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"},{"constant":false,"inputs":[{"name":"_from","type":"address"},{"name":"_to","type":"address"},{"name":"_tokenId","type":"uint256"}],"name":"safeTransferFrom","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`

// ERC721 轻量封装
type ERC721 struct {
	TokenContract string
	IAcc          *IAccount
}

func NewERC721(tokenContract string, acc *IAccount) *ERC721 {
	return &ERC721{
		TokenContract: tokenContract,
		IAcc:          acc,
	}
}

func (e *ERC721) Name() (string, error) {
	data, err := Pack(ERC721ABI, "name")
	if err != nil {
		return "", err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return "", err
	}
	vals, err := Unpack(ERC721ABI, "name", out)
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

func (e *ERC721) Symbol() (string, error) {
	data, err := Pack(ERC721ABI, "symbol")
	if err != nil {
		return "", err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return "", err
	}
	vals, err := Unpack(ERC721ABI, "symbol", out)
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

func (e *ERC721) TotalSupply() (*big.Int, error) {
	data, err := Pack(ERC721ABI, "totalSupply")
	if err != nil {
		return nil, err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return nil, err
	}
	vals, err := Unpack(ERC721ABI, "totalSupply", out)
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

func (e *ERC721) OwnerOf(tokenId *big.Int) (string, error) {
	data, err := Pack(ERC721ABI, "ownerOf", tokenId)
	if err != nil {
		return "", err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return "", err
	}
	vals, err := Unpack(ERC721ABI, "ownerOf", out)
	if err != nil {
		return "", err
	}
	if len(vals) > 0 {
		if addr, ok := vals[0].(common.Address); ok {
			return addr.Hex(), nil
		}
	}
	return "", nil
}

func (e *ERC721) BalanceOf(owner string) (*big.Int, error) {
	data, err := Pack(ERC721ABI, "balanceOf", common.HexToAddress(owner))
	if err != nil {
		return nil, err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return nil, err
	}
	vals, err := Unpack(ERC721ABI, "balanceOf", out)
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

func (e *ERC721) TokenURI(tokenId *big.Int) (string, error) {
	data, err := Pack(ERC721ABI, "tokenURI", tokenId)
	if err != nil {
		return "", err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return "", err
	}
	vals, err := Unpack(ERC721ABI, "tokenURI", out)
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

func (e *ERC721) GetApproved(tokenId *big.Int) (string, error) {
	data, err := Pack(ERC721ABI, "getApproved", tokenId)
	if err != nil {
		return "", err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return "", err
	}
	vals, err := Unpack(ERC721ABI, "getApproved", out)
	if err != nil {
		return "", err
	}
	if len(vals) > 0 {
		if addr, ok := vals[0].(common.Address); ok {
			return addr.Hex(), nil
		}
	}
	return "", nil
}

func (e *ERC721) IsApprovedForAll(owner, operator string) (bool, error) {
	data, err := Pack(ERC721ABI, "isApprovedForAll", common.HexToAddress(owner), common.HexToAddress(operator))
	if err != nil {
		return false, err
	}
	out, err := e.IAcc.OnlyReadCall(e.TokenContract, data)
	if err != nil {
		return false, err
	}
	vals, err := Unpack(ERC721ABI, "isApprovedForAll", out)
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

func (e *ERC721) Transfer(to string, tokenId *big.Int) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC721ABI,
		"transfer", big.NewInt(0), common.HexToAddress(to), tokenId)
}

func (e *ERC721) TransferFrom(from, to string, tokenId *big.Int) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC721ABI,
		"transferFrom", big.NewInt(0), common.HexToAddress(from), common.HexToAddress(to), tokenId)
}

func (e *ERC721) Approve(approved string, tokenId *big.Int) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC721ABI,
		"approve", big.NewInt(0), common.HexToAddress(approved), tokenId)
}

func (e *ERC721) SetApprovalForAll(operator string, approved bool) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC721ABI,
		"setApprovalForAll", big.NewInt(0), common.HexToAddress(operator), approved)
}

func (e *ERC721) SafeTransferFrom(to string, tokenId *big.Int) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC721ABI,
		"safeTransferFrom", big.NewInt(0), common.HexToAddress(to), tokenId)
}

func (e *ERC721) SafeTransferFromWithFrom(from, to string, tokenId *big.Int) (*types.Transaction, error) {
	return SendContractMethod(e.IAcc, e.TokenContract, ERC721ABI,
		"safeTransferFrom", big.NewInt(0), common.HexToAddress(from), common.HexToAddress(to), tokenId)
}
