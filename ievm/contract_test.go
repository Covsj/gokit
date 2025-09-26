package ievm

import (
	"testing"

	log "github.com/Covsj/gokit/ilog"
)

func TestContract(t *testing.T) {
	contract, err := NewContract("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", "eth", "", "")
	if err != nil {
		t.Fatalf("Failed to create contract: %v", err)
	}

	name, err := contract.Name()
	if err != nil {
		t.Errorf("Failed to get name: %v", err)
	} else {
		log.Info("测试Name", "Name", name)
	}

	symbol, err := contract.Symbol()
	if err != nil {
		t.Errorf("Failed to get symbol: %v", err)
	} else {
		log.Info("测试Symbol", "Symbol", symbol)
	}

	decimals, err := contract.Decimals()
	if err != nil {
		t.Errorf("Failed to get decimals: %v", err)
	} else {
		log.Info("测试Decimals", "Decimals", decimals)
	}

	totalSupply, err := contract.TotalSupply()
	if err != nil {
		t.Errorf("Failed to get total supply: %v", err)
	} else {
		log.Info("测试TotalSupply", "TotalSupply", FromDecimals(totalSupply, int64(decimals)))
	}

	allowance, err := contract.Allowance("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599", "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599")
	if err != nil {
		t.Errorf("Failed to get allowance: %v", err)
	} else {
		log.Info("测试Allowance", "Allowance", FromDecimals(allowance, int64(decimals)))
	}

	balance, err := contract.BalanceOf("0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599")
	if err != nil {
		t.Errorf("Failed to get balance: %v", err)
	} else {
		log.Info("测试BalanceOf", "BalanceOf", FromDecimals(balance, int64(decimals)))
	}
}
