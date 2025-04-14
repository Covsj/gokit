package evm

import (
	"testing"

	log "github.com/Covsj/gokit/pkg/ilog"
)

func TestContract(t *testing.T) {
	contract := NewContract("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", "eth", "", "")
	log.Info("测试Name", "Name", contract.Name())

	log.Info("测试Symbol", "Symbol", contract.Symbol())

	decimals := contract.Decimals()
	log.Info("测试Decimals", "Decimals", decimals)

	totalSupply := contract.TotalSupply()
	log.Info("测试TotalSupply", "TotalSupply", FromDecimals(totalSupply, int64(decimals)))

	allowance := contract.Allowance("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599")
	log.Info("测试Allowance", "Allowance", FromDecimals(allowance, int64(decimals)))

	balance := contract.BalanceOf("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599")
	log.Info("测试BalanceOf", "BalanceOf", FromDecimals(balance, int64(decimals)))
}
