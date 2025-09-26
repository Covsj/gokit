package ievm

import (
	"testing"

	"github.com/Covsj/gokit/ilog"
	log "github.com/Covsj/gokit/ilog"
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

func TestNewWithMnemonic(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		address  string // If the generated address can match, there is no problem.
		wantErr  bool
	}{
		{name: "valid account 1", mnemonic: accountCase1.mnemonic, address: accountCase1.address},
		{name: "valid account 2", mnemonic: accountCase2.mnemonic, address: accountCase2.address},
		{name: "error mnemonic", mnemonic: errorCase.mnemonic, wantErr: true},
		{name: "empty mnemonic", mnemonic: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWithMnemonic(tt.mnemonic)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWithMnemonic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && got.Address() != tt.address {
				t.Errorf("NewWithMnemonic() address = %v, want %v", got.Address(), tt.address)
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
	account, err := NewWithMnemonic(mnemonic)
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
	signature, err := account.Sign([]byte("1ww12e21e23"), "")
	ilog.Info("测试签名", "sig", signature, "err", err)
}

func TestNewWithPrivateKey(t *testing.T) {
	tests := []struct {
		name        string
		privateKey  string
		expectedErr bool
	}{
		{
			name:        "valid private key with 0x prefix",
			privateKey:  "0x8c3083c24062f065ff2ee71b21f665375b266cebffa920e8909ec7c48006725d",
			expectedErr: false,
		},
		{
			name:        "valid private key without 0x prefix",
			privateKey:  "8c3083c24062f065ff2ee71b21f665375b266cebffa920e8909ec7c48006725d",
			expectedErr: false,
		},
		{
			name:        "empty private key",
			privateKey:  "",
			expectedErr: true,
		},
		{
			name:        "invalid private key",
			privateKey:  "invalid",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := NewWithPrivateKey(tt.privateKey)
			if (err != nil) != tt.expectedErr {
				t.Errorf("NewWithPrivateKey() error = %v, expectedErr %v", err, tt.expectedErr)
				return
			}
			if err == nil && account == nil {
				t.Error("NewWithPrivateKey() returned nil account without error")
			}
		})
	}
}

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name    string
		address string
		want    bool
	}{
		{
			name:    "valid address",
			address: "0x7161ada3EA6e53E5652A45988DdfF1cE595E09c2",
			want:    true,
		},
		{
			name:    "invalid address - too short",
			address: "0x7161ada3EA6e53E5652A45988DdfF1cE595E09c",
			want:    false,
		},
		{
			name:    "invalid address - no 0x prefix",
			address: "7161ada3EA6e53E5652A45988DdfF1cE595E09c2",
			want:    false,
		},
		{
			name:    "empty address",
			address: "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateAddress(tt.address); got != tt.want {
				t.Errorf("ValidateAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}
