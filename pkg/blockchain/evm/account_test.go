package evm

import (
	"testing"

	"github.com/Covsj/gokit/pkg/log"
)

type TestAccountCase struct {
	mnemonic string
	address  string
}

var accountCase1 = &TestAccountCase{
	mnemonic: "unaware oxygen allow method allow property predict various slice travel please priority",
	address:  "0x7161ada3EA6e53E5652A45988DdfF1cE595E09c2",
}
var accountCase2 = &TestAccountCase{
	mnemonic: "police saddle quote salon run split notice taxi expand uniform zone excess",
	address:  "0xD32D26054099DbB5A14387d0cF15Df4452EFE4a9",
}

var errorCase = &TestAccountCase{mnemonic: "unaware oxygen allow method allow property predict various slice travel please wrong"}

func TestNewAccountWithMnemonic(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		address  string // If the generated address can match, there is no problem.
		wantErr  bool
	}{
		{name: "valid account 1", mnemonic: accountCase1.mnemonic, address: accountCase1.address},
		{name: "valid account 2", mnemonic: accountCase2.mnemonic, address: accountCase2.address},
		{name: "error mnemonic", mnemonic: errorCase.mnemonic, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewAccountWithMnemonic(tt.mnemonic)
			if (err != nil) != tt.wantErr {
				log.Error("NewAccountWithMnemonic error", "error", err.Error(), "wantErr", err.Error())
				return
			}
			if (err == nil) && got.Address() != tt.address {
				log.Error("NewAccountWithMnemonic error", "want", tt.address, "got", got)
			}
		})
	}
}

func TestETHWallet_Privatekey_Publickey_Address(t *testing.T) {
	// 从 coming 的 trust wallet 库计算的测试用例
	// private key = 0x8c3083c24062f065ff2ee71b21f665375b266cebffa920e8909ec7c48006725d
	// public key  = 0xc66cbe3908fda67d2fb229b13a63aa1a2d8428acef2ff67bc31f6a79f2e2085f // Curve25519
	// public key  = 0xb34ec4ec2ebc84b04d9170bed91f65306c7045863efb9175d721104a8ecc17f2 // Ed25519
	// public key  = 0x011e56a004e205db53ae3cc7291ffb8a28181aed4b4e95813c17b9a96db2d769 // Ed25519Blake2b
	// public key  = 0x04bd6d7af856d20188fcfdb8ff38b978bc7c72fd028b67a6fab3d2120dd9bd1db61c5d44e242001dce224188a8b88150e16e9748438703bbf2dc417135c4f9377e // Secp256k1 compressed false
	// public key  = 0x02bd6d7af856d20188fcfdb8ff38b978bc7c72fd028b67a6fab3d2120dd9bd1db6 // Secp256k1 compressed true
	// public key  = 0x027bcb5a6edf262eca9602b8343baa1cd5dd7811e540e850b05661b6524e504222 // Nist256p1
	// address     = 0x7161ada3EA6e53E5652A45988DdfF1cE595E09c2

	mnemonic := "unaware oxygen allow method allow property predict various slice travel please priority"
	account, err := NewAccountWithMnemonic(mnemonic)
	if err != nil {
		log.Error("测试公私钥相关", "error", err.Error())
		return
	}
	privateKeyHex, err := account.PrivateKeyHex()
	if err != nil {
		log.Error("测试公私钥相关", "error", err.Error())
		return
	}
	log.Info("测试公私钥相关", "privateKeyHex", privateKeyHex)

	publicKeyHex := account.PublicKeyHex()
	log.Info("测试公私钥相关", "publicKeyHex", publicKeyHex)

	address := account.Address()
	log.Info("测试公私钥相关", "address", address)
}
