package icrypto

import (
	"fmt"
	"testing"
)

func TestAes(t *testing.T) {
	mode, padding, key, iv := "cbc", "PKCS7", "0123456789abcdef", "0123456789abcdef"

	decrypter := DecryptAES("5e80323dfd3d8f812e5b88bd32ef56a53dbb346ed0415a123f8b7c99e3006fd4", mode, padding, key, iv, "hex")
	fmt.Println(decrypter.ToString())

	encrypter := EncryptAES(decrypter.ToString(), mode, padding, key, iv)
	fmt.Println(encrypter.ToHexString())
}
