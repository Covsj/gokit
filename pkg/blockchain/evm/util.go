package evm

import (
	"math/big"
	"math/rand"
	"time"

	"github.com/chenzhijie/go-web3/utils"
)

func rpcByChainId(chainId int64) string {
	var rpc_list []string
	if chainId == 56 {
		rpc_list = rpc_bsc_list
	} else {
		rpc_list = rpc_eth_list
	}
	rand.New(rand.NewSource(time.Now().Unix()))
	return rpc_list[rand.Intn(len(rpc_list))]
}

func chainIdByEvm(evm string) int64 {
	if evm == "bsc" {
		return 56
	}
	return 1
}

func FromDecimals(value *big.Int, decimals int64) *big.Float {
	return utils.NewUtils().FromDecimals(value, decimals)
}

func ToDecimals(value uint64, decimals int64) *big.Int {
	return utils.NewUtils().ToDecimals(value, decimals)
}
