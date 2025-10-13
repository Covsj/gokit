package evm

import (
	"testing"

	"github.com/Covsj/gokit/ilog"
)

func TestNewWithMnemonicIndex(t *testing.T) {
	mn, err := GenerateMnemonic()
	if err != nil {
		ilog.Error("生成助记词失败", "错误", err)
		return
	}
	ilog.Info("生成助记词成功", "助记词", mn)

	acc, err := NewWithMnemonicIndex(mn, 0, "https://arbitrum-one.public.blastapi.io")
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
}
