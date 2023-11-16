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

// generate crypto random for Session ID as well PW Salt.
func randomCryptoBytes() ([]byte, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// For session_id use
func GenerateCryptoString() (string, error) {
	bytes, err := randomCryptoBytes()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hasher(salt []byte, plainTxtPW string) ([]byte, error) {
	passwordBytes := []byte(plainTxtPW)

	hash, err := scrypt.Key(passwordBytes, salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	passwordHash := append(salt, hash...)
	return passwordHash, nil
} // base64.StdEncoding.EncodeToString(passwordHash),

func VerifyPassword(storedHashStr, password string) (bool, error) {
	byteHash, err := base64.StdEncoding.DecodeString(storedHashStr)
	if err != nil {
		return false, err
	}
	salt := byteHash[:16]
	storedHashBytes := byteHash[16:]

	pwHashed, err := hasher(salt, password)
	if err != nil {
		return false, err

	} else if len(pwHashed) != len(storedHashBytes) {
		return false, nil
	}

	for i := range pwHashed {
		if pwHashed[i] != storedHashBytes[i] {
			return false, nil
		}
	}
	return true, nil
}

func GetPasswordHash(plainTxtPW string) string {
	salt, _ := randomCryptoBytes()
	hash, _ := hasher(salt, plainTxtPW)
	return base64.StdEncoding.EncodeToString(hash)
}
