package icrypto

import (
	"fmt"
	"testing"
)

func TestDes(t *testing.T) {
	mode, padding, key, iv := "ecb", "PKCS7", "12349568", ""

	decrypter := DecryptDES("3352f48d3009d7ccd43598d13bf8d6e47117521f4f8eca1bc3b7751b40a0160b", mode, padding, key, iv, "hex")
	fmt.Println(decrypter.ToString())

	encrypter := EncryptDES(decrypter.ToString(), mode, padding, key, iv)
	fmt.Println(encrypter.ToHexString())
}

func Test3Des(t *testing.T) {
	rawStr := "hello world"
	mode, padding, key, iv := "ecb", "PKCS7", "12349568", ""
	encrypter := Encrypt3DES(rawStr, mode, padding, key, iv)
	fmt.Println(encrypter.ToBase64String())

	decrypter := Decrypt3DES("3352f48d3009d7ccd43598d13bf8d6e47117521f4f8eca1bc3b7751b40a0160b", mode, padding, key, iv, "hex")
	fmt.Println(decrypter.ToString())
}
