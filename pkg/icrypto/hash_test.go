package icrypto

import (
	"fmt"
	"testing"
)

func TestHash(t *testing.T) {
	md5 := HashGenerator("hello world", "Sha512-224")
	fmt.Println(md5.ToHexString())

	encode := BaseEncode(md5.ToHexString(), "100")
	fmt.Println(encode.ToString())
	decode := BaseDecode(encode.String(), "100")
	fmt.Println(decode.ToString())
	fmt.Println(decode.ToString() == md5.ToHexString())
}

func TestHex(t *testing.T) {
	raw := "61615616"
	decode := HexDecode(raw)
	fmt.Println(decode.ToString())
	encode := HexEncode(decode.ToString())
	fmt.Println(encode.ToString())
}

func TestHmac(t *testing.T) {
	hmac := HmacGenerator("hello world", "Sha512-224", "hmac-sha512-224")
	fmt.Println(hmac.ToHexString())

	hmac = HmacGenerator(hmac.ToHexString(), "TestHmac", "hmac-ripemd160")
	fmt.Println(hmac.ToHexString())

}
