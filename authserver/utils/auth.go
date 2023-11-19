package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/scrypt"
)

var SECRET []byte = []byte(os.Getenv("SECRET_KEY"))

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

	pwHashed, err := hasher(salt, password)
	if err != nil {
		fmt.Println(err)
		return false, err

	} else if len(pwHashed) != len(byteHash) {
		fmt.Println("Hash not same len")
		return false, nil
	}

	for i := range pwHashed {
		if pwHashed[i] != byteHash[i] {
			fmt.Println("Hash Byte mismatch")
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

type TokenClaims struct {
	user_id  int
	username string
	is_staff bool
	IAT      time.Time
}

func GenerateAccessToken(id int, username string, is_staff bool) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  id,
		"username": username,
		"is_staff": is_staff,
		"IAT":      time.Now().Add(time.Minute * 15),
	})
	signedToken, err := token.SignedString(SECRET)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return SECRET, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("Invalid Claims")
	}
	return claims, nil
}
