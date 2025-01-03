package evm

import (
	"math/big"

	"github.com/Covsj/gokit/pkg/log"
	"github.com/chenzhijie/go-web3"
	"github.com/chenzhijie/go-web3/eth"
	"github.com/chenzhijie/go-web3/types"
	"github.com/chenzhijie/go-web3/utils"
	"github.com/ethereum/go-ethereum/common"
)

type Contract struct {
	Evm             string `json:"evm"`
	ChainId         int64  `json:"chain_id"`
	ContractAddress string `json:"contract_address"`
	AbiStr          string `json:"abi_str"`

	Contract *eth.Contract `json:"contract"`
	Web3     *web3.Web3    `json:"web3"`

	ContractName        string       `json:"contract_name"`
	ContractSymbol      string       `json:"contract_symbol"`
	ContractDecimals    uint8        `json:"contract_decimals"`
	ContractTotalSupply *big.Int     `json:"contract_total_supply"`
	Utils               *utils.Utils `json:"utils"`
}

func NewContract(ContractAddress string, evm string, abiStr string, privateKey string) *Contract {
	if abiStr == "" {
		abiStr = erc20_abi
	}
	c := &Contract{
		ContractAddress: ContractAddress,
		Evm:             evm,
		ChainId:         chainIdByEvm(evm),
		AbiStr:          abiStr,
	}
	web3, err := web3.NewWeb3(rpcByChainId(c.ChainId))
	if err != nil {
		log.Error("初始化合约实例失败", "详情", "初始化web3错误", "合约地址", ContractAddress, "错误", err.Error())
		return nil
	}

	web3.Eth.SetChainId(c.ChainId)
	if privateKey != "" {
		err = web3.Eth.SetAccount(privateKey)
		log.Error("初始化合约实例失败", "详情", "初始化账户错误", "合约地址", ContractAddress, "错误", err.Error())
	}

	c.Web3 = web3
	c.Utils = web3.Utils
	contract, err := web3.Eth.NewContract(c.AbiStr, c.ContractAddress)
	if err != nil {
		log.Error("初始化合约实例失败", "详情", "初始化合约错误", "合约地址", ContractAddress, "错误", err.Error())
		return nil
	}
	c.Contract = contract
	log.Info("初始化合约实例成功", "合约地址", ContractAddress, "公链", c.Evm)
	return c
}
func (c *Contract) Name() string {
	if c.ContractName != "" {
		return c.ContractName
	}
	name, err := c.Contract.Call("name")
	if err != nil {
		log.Error("获取合约名称失败", "合约地址", c.ContractAddress, "公链", c.Evm, "错误", err)
		return ""
	}
	c.ContractName = name.(string)
	return c.ContractName
}

func (c *Contract) Symbol() string {
	if c.ContractSymbol != "" {
		return c.ContractSymbol
	}
	symbol, err := c.Contract.Call("symbol")
	if err != nil {
		log.Error("获取合约符号失败", "合约地址", c.ContractAddress, "公链", c.Evm, "错误", err)
		return ""
	}
	c.ContractSymbol = symbol.(string)
	return c.ContractSymbol
}

func (c *Contract) Decimals() uint8 {
	if c.ContractDecimals != 0 {
		return c.ContractDecimals
	}
	decimals, err := c.Contract.Call("decimals")
	if err != nil {
		log.Error("获取合约精度失败", "合约地址", c.ContractAddress, "公链", c.Evm, "错误", err)
		return 0
	}

	c.ContractDecimals = decimals.(uint8)
	return c.ContractDecimals
}

func (c *Contract) BalanceOf(address string) *big.Int {
	balance, err := c.Contract.Call("balanceOf", common.HexToAddress(address))
	if err != nil {
		log.Error("获取余额失败", "合约地址", c.ContractAddress, "公链", c.Evm, "错误", err)
		return nil
	}
	return balance.(*big.Int)
}

func (c *Contract) TotalSupply() *big.Int {
	if c.ContractTotalSupply != nil {
		return c.ContractTotalSupply
	}
	totalSupply, err := c.Contract.Call("totalSupply")
	if err != nil {
		log.Error("获取代币总量失败", "合约地址", c.ContractAddress, "公链", c.Evm, "错误", err)
		return nil
	}
	c.ContractTotalSupply = totalSupply.(*big.Int)
	return c.ContractTotalSupply
}

func (c *Contract) Allowance(owner string, spender string) *big.Int {
	allowance, err := c.Contract.Call("allowance", common.HexToAddress(owner), common.HexToAddress(spender))
	if err != nil {
		log.Error("获取授权额度失败", "合约地址", c.ContractAddress, "公链", c.Evm, "错误", err)
		return nil
	}
	return allowance.(*big.Int)
}

func (c *Contract) Approve(spender string, value *big.Int) bool {
	approveInputData, err := c.Contract.Methods("approve").
		Inputs.Pack(common.HexToAddress(spender), value)
	if err != nil {
		log.Error("授权构建数据失败", "合约地址", c.ContractAddress, "公链", c.Evm, "spender", spender, "value", value, "错误", err)
		return false
	}
	err = c.Call(approveInputData)
	if err != nil {
		log.Error("授权失败", "合约地址", c.ContractAddress, "公链", c.Evm, "spender", spender, "value", value, "错误", err)
		return false
	}
	return true
}

func (c *Contract) Call(data []byte) error {

	call := &types.CallMsg{
		From: c.Web3.Eth.Address(),
		To:   common.HexToAddress(c.ContractAddress),
		Data: data,
		Gas:  types.NewCallMsgBigInt(big.NewInt(types.MAX_GAS_LIMIT)),
	}
	gasLimit, err := c.Web3.Eth.EstimateGas(call)
	if err != nil {
		log.Error("交易获取gasLimit失败", "合约地址", c.ContractAddress, "公链", c.Evm, "错误", err)
		return err
	}
	call.Gas = types.NewCallMsgBigInt(big.NewInt(int64(gasLimit)))
	nonce, err := c.Web3.Eth.GetNonce(c.Web3.Eth.Address(), nil)
	if err != nil {
		log.Error("交易获取nonce失败", "合约地址", c.ContractAddress, "公链", c.Evm, "错误", err)
		return err
	}
	txHash, err := c.Web3.Eth.SyncSendRawTransaction(
		common.HexToAddress(c.ContractAddress),
		big.NewInt(0),
		nonce,
		gasLimit,
		c.Web3.Utils.ToGWei(1),
		data,
	)
	if err != nil {
		log.Error("交易失败", "合约地址", c.ContractAddress, "公链", c.Evm, "错误", err)
		return err
	}
	log.Info("交易成功", "合约地址", c.ContractAddress, "公链", c.Evm, "txHash", txHash)
	return nil
}
