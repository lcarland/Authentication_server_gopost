package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/scrypt"
)

type JwtToken struct {
	Header  header
	Payload payload
	Sig     string
}

type header struct {
	Alg string
	Typ string
}

type payload struct {
	Sub       int
	Name      string
	SuperUser bool
	Staff     bool
}

func randomCryptoBytes() ([]byte, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func GenerateCryptoString() (string, error) {
	bytes, err := randomCryptoBytes()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func PasswordHasher(plainTxtPW string) (string, error) {
	salt, err := randomCryptoBytes()
	if err != nil {
		return "", nil
	}
	passwordBytes := []byte(plainTxtPW)

	hash, err := scrypt.Key(passwordBytes, salt, 32768, 8, 1, 32)
	if err != nil {
		return "", nil
	}
	passwordHash := append(salt, hash...)
	return base64.StdEncoding.EncodeToString(passwordHash), nil
}
