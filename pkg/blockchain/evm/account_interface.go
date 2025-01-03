package evm

// Account 定义了账户的基本接口
type IAccount interface {
	// 密钥相关
	PrivateKey() ([]byte, error)
	PrivateKeyHex() (string, error)
	PublicKey() []byte
	PublicKeyHex() string
	Address() string

	// 签名相关
	Sign(message []byte, password string) ([]byte, error)
	SignHex(messageHex string, password string) (string, error)
	SignHash(hash []byte) ([]byte, error)
}
