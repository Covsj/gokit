package icrypto

import (
	"strings"

	"gitee.com/golang-module/dongle"
)

func HexEncode(encodeStr interface{}) dongle.Encoder {
	switch v := encodeStr.(type) {
	case []byte:
		return dongle.Encode.FromBytes(v).ByHex()
	default:
		return dongle.Encode.FromString(v.(string)).ByHex()
	}
}

func HexDecode(decodeStr interface{}) dongle.Decoder {
	switch v := decodeStr.(type) {
	case []byte:
		return dongle.Decode.FromBytes(v).ByHex()
	default:
		return dongle.Decode.FromString(v.(string)).ByHex()
	}
}

func HashGenerator(encryptData interface{}, encryptMode string) dongle.Encrypter {
	// 定义一个映射表，将加密模式与加密操作关联
	encryptMode = strings.ToLower(encryptMode)
	encryptFuncMap := map[string]func([]byte) dongle.Encrypter{
		"md2":        func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByMd2() },
		"md4":        func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByMd4() },
		"md5":        func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByMd5() },
		"sha1":       func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).BySha1() },
		"sha3-224":   func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).BySha3(224) },
		"sha3-256":   func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).BySha3(256) },
		"sha3-384":   func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).BySha3(384) },
		"sha3-512":   func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).BySha3(512) },
		"sha224":     func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).BySha224() },
		"sha256":     func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).BySha256() },
		"sha384":     func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).BySha384() },
		"sha512":     func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).BySha512() },
		"sha512-224": func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).BySha512(224) },
		"sha512-256": func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).BySha512(256) },
		"shake128-256": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByShake128(256)
		},
		"shake128-512": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByShake128(512)
		},
		"shake256-384": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByShake256(384)
		},
		"shake256-512": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByShake256(512)
		},
		"ripemd160": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByRipemd160()
		},
		"blake2b-256": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByBlake2b(256)
		},
		"blake2b-384": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByBlake2b(384)
		},
		"blake2b-512": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByBlake2b(512)
		},
		"blake2s-256": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByBlake2s(256)
		},
	}

	// 获取加密数据的字节形式
	var dataBytes []byte
	switch v := encryptData.(type) {
	case []byte:
		dataBytes = v
	case string:
		dataBytes = []byte(v)
	default:
		panic("unsupported encrypt data type")
	}

	// 根据加密模式执行加密操作，默认为 Md5
	if encryptFunc, exists := encryptFuncMap[encryptMode]; exists {
		return encryptFunc(dataBytes)
	}

	// 如果模式不匹配，默认使用 Md5
	return dongle.Encrypt.FromBytes(dataBytes).ByMd5()
}

func HmacGenerator(encryptData, encryptKey interface{}, encryptMode string) dongle.Encrypter {
	// 定义一个映射表，将加密模式与加密操作关联
	encryptMode = strings.ToLower(encryptMode)

	encryptFuncMap := map[string]func([]byte) dongle.Encrypter{
		"hmac-md2":      func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacMd2(encryptKey) },
		"hmac-md4":      func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacMd4(encryptKey) },
		"hmac-md5":      func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacMd5(encryptKey) },
		"hmac-sha1":     func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacSha1(encryptKey) },
		"hmac-sha3-224": func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacSha3(encryptKey, 224) },
		"hmac-sha3-256": func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacSha3(encryptKey, 256) },
		"hmac-sha3-384": func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacSha3(encryptKey, 384) },
		"hmac-sha3-512": func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacSha3(encryptKey, 512) },
		"hmac-sha224":   func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacSha224(encryptKey) },
		"hmac-sha256":   func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacSha256(encryptKey) },
		"hmac-sha384":   func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacSha384(encryptKey) },
		"hmac-sha512":   func(data []byte) dongle.Encrypter { return dongle.Encrypt.FromBytes(data).ByHmacSha512(encryptKey) },
		"hmac-sha512-224": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByHmacSha512(encryptKey, 224)
		},
		"hmac-sha512-256": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByHmacSha512(encryptKey, 256)
		},
		"hmac-ripemd160": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByHmacRipemd160(encryptKey)
		},
		"hmac-sm3": func(data []byte) dongle.Encrypter {
			return dongle.Encrypt.FromBytes(data).ByHmacSm3(encryptKey)
		},
	}

	// 获取加密数据的字节形式
	var dataBytes []byte
	switch v := encryptData.(type) {
	case []byte:
		dataBytes = v
	case string:
		dataBytes = []byte(v)
	default:
		panic("unsupported encrypt data type")
	}

	// 根据加密模式执行加密操作，默认为 Md5
	if encryptFunc, exists := encryptFuncMap[encryptMode]; exists {
		return encryptFunc(dataBytes)
	}

	// 如果模式不匹配，默认使用 Md5
	return dongle.Encrypt.FromBytes(dataBytes).ByHmacMd5(encryptKey)
}

func BaseEncode(encodeStr interface{}, baseMode string) dongle.Encoder {
	// 定义一个映射表，将编码模式与对应的编码函数关联
	encodeFuncMap := map[string]func([]byte) dongle.Encoder{
		"16":    func(data []byte) dongle.Encoder { return dongle.Encode.FromBytes(data).ByBase16() },
		"32":    func(data []byte) dongle.Encoder { return dongle.Encode.FromBytes(data).ByBase32() },
		"45":    func(data []byte) dongle.Encoder { return dongle.Encode.FromBytes(data).ByBase45() },
		"58":    func(data []byte) dongle.Encoder { return dongle.Encode.FromBytes(data).ByBase58() },
		"62":    func(data []byte) dongle.Encoder { return dongle.Encode.FromBytes(data).ByBase62() },
		"64":    func(data []byte) dongle.Encoder { return dongle.Encode.FromBytes(data).ByBase64() },
		"64URL": func(data []byte) dongle.Encoder { return dongle.Encode.FromBytes(data).ByBase64URL() },
		"85":    func(data []byte) dongle.Encoder { return dongle.Encode.FromBytes(data).ByBase85() },
		"91":    func(data []byte) dongle.Encoder { return dongle.Encode.FromBytes(data).ByBase91() },
		"100":   func(data []byte) dongle.Encoder { return dongle.Encode.FromBytes(data).ByBase100() },
	}

	// 获取编码数据的字节形式
	var dataBytes []byte
	switch v := encodeStr.(type) {
	case []byte:
		dataBytes = v
	case string:
		dataBytes = []byte(v)
	default:
		panic("unsupported encode data type")
	}

	// 根据编码模式执行相应的编码操作，默认为 Base64
	if encodeFunc, exists := encodeFuncMap[baseMode]; exists {
		return encodeFunc(dataBytes)
	}

	// 如果模式不匹配，默认使用 Base64 编码
	return dongle.Encode.FromBytes(dataBytes).ByBase64()
}

func BaseDecode(decodeStr interface{}, baseMode string) dongle.Decoder {
	// 定义一个映射表，将解码模式与对应的解码函数关联
	decodeFuncMap := map[string]func([]byte) dongle.Decoder{
		"16":    func(data []byte) dongle.Decoder { return dongle.Decode.FromBytes(data).ByBase16() },
		"32":    func(data []byte) dongle.Decoder { return dongle.Decode.FromBytes(data).ByBase32() },
		"45":    func(data []byte) dongle.Decoder { return dongle.Decode.FromBytes(data).ByBase45() },
		"58":    func(data []byte) dongle.Decoder { return dongle.Decode.FromBytes(data).ByBase58() },
		"62":    func(data []byte) dongle.Decoder { return dongle.Decode.FromBytes(data).ByBase62() },
		"64":    func(data []byte) dongle.Decoder { return dongle.Decode.FromBytes(data).ByBase64() },
		"64URL": func(data []byte) dongle.Decoder { return dongle.Decode.FromBytes(data).ByBase64URL() },
		"85":    func(data []byte) dongle.Decoder { return dongle.Decode.FromBytes(data).ByBase85() },
		"91":    func(data []byte) dongle.Decoder { return dongle.Decode.FromBytes(data).ByBase91() },
		"100":   func(data []byte) dongle.Decoder { return dongle.Decode.FromBytes(data).ByBase100() },
	}

	// 获取解码数据的字节形式
	var dataBytes []byte
	switch v := decodeStr.(type) {
	case []byte:
		dataBytes = v
	case string:
		dataBytes = []byte(v)
	default:
		panic("unsupported decode data type")
	}

	// 根据解码模式执行相应的解码操作，默认为 Base64
	if decodeFunc, exists := decodeFuncMap[baseMode]; exists {
		return decodeFunc(dataBytes)
	}

	// 如果模式不匹配，默认使用 Base64 解码
	return dongle.Decode.FromBytes(dataBytes).ByBase64()
}
