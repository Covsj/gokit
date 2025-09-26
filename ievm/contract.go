package ievm

import (
	"fmt"
	"math/big"

	log "github.com/Covsj/gokit/ilog"
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

	// 缓存合约信息，避免重复调用
	ContractName        string       `json:"contract_name"`
	ContractSymbol      string       `json:"contract_symbol"`
	ContractDecimals    uint8        `json:"contract_decimals"`
	ContractTotalSupply *big.Int     `json:"contract_total_supply"`
	Utils               *utils.Utils `json:"utils"`

	// 标记缓存是否已加载
	nameLoaded        bool
	symbolLoaded      bool
	decimalsLoaded    bool
	totalSupplyLoaded bool
}

func NewContract(contractAddress string, evm string, abiStr string, privateKey string) (*Contract, error) {
	if contractAddress == "" {
		return nil, fmt.Errorf("合约地址不能为空")
	}

	if !ValidateAddress(contractAddress) {
		return nil, fmt.Errorf("无效的合约地址格式: %s", contractAddress)
	}

	if abiStr == "" {
		abiStr = ERC20ABI
	}

	c := &Contract{
		ContractAddress: contractAddress,
		Evm:             evm,
		ChainId:         GetChainIDByEVM(evm),
		AbiStr:          abiStr,
	}

	rpcURL := GetRPCByChainID(c.ChainId)
	if rpcURL == "" {
		return nil, fmt.Errorf("无法获取链ID %d 的RPC端点", c.ChainId)
	}

	web3, err := web3.NewWeb3(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("初始化web3失败: %w", err)
	}

	web3.Eth.SetChainId(c.ChainId)
	if privateKey != "" {
		if err = web3.Eth.SetAccount(privateKey); err != nil {
			return nil, fmt.Errorf("设置账户失败: %w", err)
		}
	}

	c.Web3 = web3
	c.Utils = web3.Utils
	contract, err := web3.Eth.NewContract(c.AbiStr, c.ContractAddress)
	if err != nil {
		return nil, fmt.Errorf("初始化合约失败: %w", err)
	}
	c.Contract = contract
	log.Info("初始化合约实例成功", "合约地址", contractAddress, "公链", c.Evm)
	return c, nil
}
func (c *Contract) Name() (string, error) {
	if c.nameLoaded {
		return c.ContractName, nil
	}

	name, err := c.Contract.Call("name")
	if err != nil {
		return "", fmt.Errorf("获取合约名称失败: %w", err)
	}

	c.ContractName = name.(string)
	c.nameLoaded = true
	return c.ContractName, nil
}

func (c *Contract) Symbol() (string, error) {
	if c.symbolLoaded {
		return c.ContractSymbol, nil
	}

	symbol, err := c.Contract.Call("symbol")
	if err != nil {
		return "", fmt.Errorf("获取合约符号失败: %w", err)
	}

	c.ContractSymbol = symbol.(string)
	c.symbolLoaded = true
	return c.ContractSymbol, nil
}

func (c *Contract) Decimals() (uint8, error) {
	if c.decimalsLoaded {
		return c.ContractDecimals, nil
	}

	decimals, err := c.Contract.Call("decimals")
	if err != nil {
		return 0, fmt.Errorf("获取合约精度失败: %w", err)
	}

	c.ContractDecimals = decimals.(uint8)
	c.decimalsLoaded = true
	return c.ContractDecimals, nil
}

func (c *Contract) BalanceOf(address string) (*big.Int, error) {
	if !ValidateAddress(address) {
		return nil, fmt.Errorf("无效的地址格式: %s", address)
	}

	balance, err := c.Contract.Call("balanceOf", common.HexToAddress(address))
	if err != nil {
		return nil, fmt.Errorf("获取余额失败: %w", err)
	}
	return balance.(*big.Int), nil
}

func (c *Contract) TotalSupply() (*big.Int, error) {
	if c.totalSupplyLoaded {
		return c.ContractTotalSupply, nil
	}

	totalSupply, err := c.Contract.Call("totalSupply")
	if err != nil {
		return nil, fmt.Errorf("获取代币总量失败: %w", err)
	}

	c.ContractTotalSupply = totalSupply.(*big.Int)
	c.totalSupplyLoaded = true
	return c.ContractTotalSupply, nil
}

func (c *Contract) Allowance(owner string, spender string) (*big.Int, error) {
	if !ValidateAddress(owner) {
		return nil, fmt.Errorf("无效的owner地址格式: %s", owner)
	}
	if !ValidateAddress(spender) {
		return nil, fmt.Errorf("无效的spender地址格式: %s", spender)
	}

	allowance, err := c.Contract.Call("allowance", common.HexToAddress(owner), common.HexToAddress(spender))
	if err != nil {
		return nil, fmt.Errorf("获取授权额度失败: %w", err)
	}
	return allowance.(*big.Int), nil
}

func (c *Contract) Approve(spender string, value *big.Int) error {
	if !ValidateAddress(spender) {
		return fmt.Errorf("无效的spender地址格式: %s", spender)
	}
	if value == nil {
		return fmt.Errorf("授权金额不能为空")
	}

	approveInputData, err := c.Contract.Methods("approve").
		Inputs.Pack(common.HexToAddress(spender), value)
	if err != nil {
		return fmt.Errorf("授权构建数据失败: %w", err)
	}

	if err = c.Call(approveInputData); err != nil {
		return fmt.Errorf("授权失败: %w", err)
	}
	return nil
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
