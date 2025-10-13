package ievm

import (
	"fmt"
	"testing"

	"github.com/Covsj/gokit/ilog"
)

func TestNewWithMnemonicIndex(t *testing.T) {
	fmt.Println(ToDecimalsFloat("1.12", 6))
	mn, err := GenerateMnemonic()
	if err != nil {
		ilog.Error("生成助记词失败", "错误", err)
		return
	}
	ilog.Info("生成助记词成功", "助记词", mn)

	acc, err := NewWithMnemonicIndex(mn, 4, "https://base.llamarpc.com")
	if err != nil {
		ilog.Error("生成客户端失败", "错误", err)
		return
	}
	address := acc.Address()
	privateKey := acc.PrivateKeyHex()
	ilog.Info("加载数据成功", "地址", address, "私钥", privateKey)

	nonce, err := acc.Nonce(address)
	if err != nil {
		ilog.Error("调用方法失败", "错误", err)
		return
	}
	banlance, err := acc.Balance(address)
	if err != nil {
		ilog.Error("调用方法失败", "错误", err)
		return
	}

	ilog.Info("加载数据成功", "nonce", nonce, "余额", FromDecimals(banlance, 18))

	iErc20 := NewERC20("0x833589fcd6edb6e08f4c7c32d4f71b54bda02913", acc)
	usdc_balance, err := iErc20.BalanceOf(address)
	if err != nil {
		ilog.Error("调用方法失败", "错误", err)
		return
	}
	ilog.Info("加载数据成功", "USDC余额", FromDecimals(usdc_balance, 6))

	limitless01 := ""
	usdc_allowance, err := iErc20.Allowance(address, limitless01)
	if err != nil {
		ilog.Error("调用方法失败", "错误", err)
		return
	}
	ilog.Info("加载数据成功", "USDC授权额度1", usdc_allowance)

	// tx, err := iErc20.Approve(limitless01, ToDecimals(21, 6))
	// if err != nil {
	// 	ilog.Error("调用方法失败", "错误", err)
	// 	return
	// }
	// ilog.Info("加载数据成功", "tx", tx.Hash())

	// usdc_allowance, err = iErc20.Allowance(address, limitless01)
	// if err != nil {
	// 	ilog.Error("调用方法失败", "错误", err)
	// 	return
	// }
	// ilog.Info("加载数据成功", "USDC授权额度2", usdc_allowance)

	// tx, err := iErc20.Transfer(limitless01, ToDecimals(1, 6))
	// if err != nil {
	// 	ilog.Error("调用方法失败", "错误", err)
	// 	return
	// }
	// ilog.Info("加载数据成功", "tx", tx.Hash())

}
