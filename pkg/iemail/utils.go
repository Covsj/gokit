package iemail

import (
	"crypto/rand"
	"math/big"
)

const (
	usernameCharset = "abcdefghijklmnopqrstuvwxyz0123456789"
	passwordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	passwordLength  = 8
)

func generateRandom(usernameLength int) (string, string) {
	username := randomStringFromCharset(usernameLength, usernameCharset)
	password := randomStringFromCharset(passwordLength, passwordCharset)
	return username, password
}

func randomStringFromCharset(length int, charset string) string {
	b := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))
	for i := range b {
		randIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			panic("Failed to generate secure random number: " + err.Error())
		}
		b[i] = charset[randIndex.Int64()]
	}
	return string(b)
}
