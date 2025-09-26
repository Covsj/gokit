package icrypto

import (
	"strings"

	"gitee.com/golang-module/dongle"
)

// des key 长度必须是 8 字节 iv 长度必须是 8 字节
// 3des key 长度必须是 24 iv 长度必须是 8
// aes key 长度必须是 16、24 或 32 字节, iv 长度必须是 16 字节，ECB 模式不需要设置 iv
func NewAESCipher(mode, padding string, key, iv interface{}) *dongle.Cipher {
	mode, padding = strings.ToUpper(mode), strings.ToUpper(padding)
	cipher := dongle.NewCipher()
	// CBC、ECB、CFB、OFB、CTR
	switch mode {
	case "CBC":
		cipher.SetMode(dongle.CBC)
	case "ECB":
		cipher.SetMode(dongle.ECB)
	case "CFB":
		cipher.SetMode(dongle.CFB)
	case "OFB":
		cipher.SetMode(dongle.OFB)
	default:
		cipher.SetMode(dongle.CTR)
	}
	// No、Empty、Zero、PKCS5、PKCS7、AnsiX923、ISO97971
	switch padding {
	case "NO":
		cipher.SetPadding(dongle.No)
	case "EMPTY":
		cipher.SetPadding(dongle.Empty)
	case "ZERO":
		cipher.SetPadding(dongle.Zero)
	case "PKCS5":
		cipher.SetPadding(dongle.PKCS5)
	case "PKCS7":
		cipher.SetPadding(dongle.PKCS7)
	case "ANSIX923":
		cipher.SetPadding(dongle.AnsiX923)
	default:
		cipher.SetPadding(dongle.ISO97971)
	}

	cipher.SetKey(key)
	if iv != nil {
		cipher.SetIV(iv)
	}
	return cipher
}

func EncryptAES(data interface{}, mode, padding string, aesKey, aesIv interface{}) dongle.Encrypter {
	cipher := NewAESCipher(mode, padding, aesKey, aesIv)
	var encrypter dongle.Encrypter
	switch v := data.(type) {
	case string:
		encrypter = dongle.Encrypt.FromString(v).ByAes(cipher)
	case []byte:
		encrypter = dongle.Encrypt.FromBytes(v).ByAes(cipher)
	default:
		return encrypter
	}
	return encrypter
}

func DecryptAES(encryptedData interface{}, mode, padding string, aesKey, aesIv interface{}, encodingMode string) dongle.Decrypter {

	cipher := NewAESCipher(mode, padding, aesKey, aesIv)

	encodingMode = strings.ToLower(encodingMode)
	switch encodingMode {
	case "raw":
		return dongle.Decrypt.FromRawString(encryptedData.(string)).ByAes(cipher)
	case "hex":
		return dongle.Decrypt.FromHexString(encryptedData.(string)).ByAes(cipher)
	case "base64":
		return dongle.Decrypt.FromBase64String(encryptedData.(string)).ByAes(cipher)
	case "bytes":
		return dongle.Decrypt.FromRawBytes(encryptedData.([]byte)).ByAes(cipher)
	case "hex-bytes":
		return dongle.Decrypt.FromHexBytes(encryptedData.([]byte)).ByAes(cipher)
	case "base64-bytes":
		return dongle.Decrypt.FromBase64Bytes(encryptedData.([]byte)).ByAes(cipher)
	}

	return dongle.Decrypter{}
}
