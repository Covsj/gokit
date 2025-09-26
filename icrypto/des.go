package icrypto

import (
	"strings"

	"gitee.com/golang-module/dongle"
)

func NewDESCipher(mode, padding string, key, iv interface{}) *dongle.Cipher {
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
	case "CTR":
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
	case "ISO97971":
		cipher.SetPadding(dongle.ISO97971)
	}

	cipher.SetKey(key)
	if iv != nil {
		cipher.SetIV(iv)
	}
	return cipher
}

func EncryptDES(data interface{}, mode, padding string, desKey, desIv interface{}) dongle.Encrypter {
	cipher := NewDESCipher(mode, padding, desKey, desIv)
	var encrypter dongle.Encrypter
	switch v := data.(type) {
	case string:
		encrypter = dongle.Encrypt.FromString(v).ByDes(cipher)
	case []byte:
		encrypter = dongle.Encrypt.FromBytes(v).ByDes(cipher)
	default:
		return encrypter
	}
	return encrypter
}

func DecryptDES(encryptedData interface{}, mode, padding string, desKey, desIv interface{}, encodingMode string) dongle.Decrypter {
	cipher := NewDESCipher(mode, padding, desKey, desIv)

	encodingMode = strings.ToLower(encodingMode)
	switch encodingMode {
	case "raw":
		return dongle.Decrypt.FromRawString(encryptedData.(string)).ByDes(cipher)
	case "hex":
		return dongle.Decrypt.FromHexString(encryptedData.(string)).ByDes(cipher)
	case "base64":
		return dongle.Decrypt.FromBase64String(encryptedData.(string)).ByDes(cipher)
	case "bytes":
		return dongle.Decrypt.FromRawBytes(encryptedData.([]byte)).ByDes(cipher)
	case "hex-bytes":
		return dongle.Decrypt.FromHexBytes(encryptedData.([]byte)).ByDes(cipher)
	case "base64-bytes":
		return dongle.Decrypt.FromBase64Bytes(encryptedData.([]byte)).ByDes(cipher)
	}

	return dongle.Decrypter{}
}

func Encrypt3DES(data interface{}, mode, padding string, desKey, desIv interface{}) dongle.Encrypter {
	cipher := NewDESCipher(mode, padding, desKey, desIv)
	var encrypter dongle.Encrypter
	switch v := data.(type) {
	case string:
		encrypter = dongle.Encrypt.FromString(v).By3Des(cipher)
	case []byte:
		encrypter = dongle.Encrypt.FromBytes(v).By3Des(cipher)
	default:
		return encrypter
	}
	return encrypter
}

func Decrypt3DES(encryptedData interface{}, mode, padding string, desKey, desIv interface{}, encodingMode string) dongle.Decrypter {
	cipher := NewDESCipher(mode, padding, desKey, desIv)

	encodingMode = strings.ToLower(encodingMode)
	switch encodingMode {
	case "bytes":
		return dongle.Decrypt.FromRawBytes(encryptedData.([]byte)).By3Des(cipher)
	case "hex-bytes":
		return dongle.Decrypt.FromHexBytes(encryptedData.([]byte)).By3Des(cipher)
	case "base64-bytes":
		return dongle.Decrypt.FromBase64Bytes(encryptedData.([]byte)).By3Des(cipher)
	case "hex":
		return dongle.Decrypt.FromHexString(encryptedData.(string)).By3Des(cipher)
	case "base64":
		return dongle.Decrypt.FromBase64String(encryptedData.(string)).By3Des(cipher)
	default:
		return dongle.Decrypt.FromRawString(encryptedData.(string)).By3Des(cipher)
	}
}
